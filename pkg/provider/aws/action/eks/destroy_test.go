package eks

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestCleanupIdentityFromRootOutput(t *testing.T) {
	raw := json.RawMessage(`{
		"manifest":{"unknownFutureField":"preserved"},
		"resources":[
			{"type":"aws:eks/cluster:Cluster","urn":"cluster","id":"test-cluster","outputs":{"arn":"arn:aws:eks:us-east-2:123456789012:cluster/test-cluster"}},
			{"type":"pulumi:providers:aws","urn":"provider","inputs":{"secretKey":{"ciphertext":"do-not-change"}}},
			{"type":"pulumi:pulumi:Stack","urn":"stack","outputs":{"eksCleanup":{"version":1,"accountId":"123456789012","region":"us-east-2","clusterName":"test-cluster","vpcId":"vpc-test"},"existing":"keep"}}
		]}`)
	identity, urn, err := cleanupIdentityFromCheckpoint(raw)
	if err != nil || identity == nil || *identity != testIdentity() || urn != "cluster" {
		t.Fatalf("identity=%+v urn=%s err=%v", identity, urn, err)
	}
}

func TestCleanupIdentityFromClusterResource(t *testing.T) {
	raw := json.RawMessage(`{"resources":[
		{"type":"aws:eks/cluster:Cluster","urn":"cluster","id":"test-cluster","outputs":{"arn":"arn:aws:eks:us-east-2:123456789012:cluster/test-cluster","vpcConfig":{"vpcId":"vpc-test"}}}
	]}`)
	identity, urn, err := cleanupIdentityFromCheckpoint(raw)
	if err != nil || identity == nil || *identity != testIdentity() || urn != "cluster" {
		t.Fatalf("identity=%+v urn=%s err=%v", identity, urn, err)
	}
}

func TestCleanupIdentityAbsent(t *testing.T) {
	for _, raw := range []string{
		`{"resources":[{"type":"pulumi:pulumi:Stack","outputs":{}}]}`,
		`{"resources":[{"type":"aws:eks/cluster:Cluster","urn":"cluster","id":""}]}`,
	} {
		identity, _, err := cleanupIdentityFromCheckpoint([]byte(raw))
		if err != nil || identity != nil {
			t.Fatalf("identity=%+v err=%v", identity, err)
		}
	}
}

func TestCreateCompleted(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{"kubeconfig output", `{"resources":[{"type":"pulumi:pulumi:Stack","outputs":{"main-eksKubeconfig":"apiVersion: v1"}}]}`, true},
		{"no outputs", `{"resources":[{"type":"pulumi:pulumi:Stack","outputs":{}}]}`, false},
		{"no stack resource", `{"resources":[]}`, false},
		{"null output", `{"resources":[{"type":"pulumi:pulumi:Stack","outputs":{"main-eksKubeconfig":null}}]}`, false},
		{"invalid checkpoint", `{`, false},
	} {
		if got := createCompleted([]byte(tc.raw)); got != tc.want {
			t.Fatalf("%s: createCompleted=%t want %t", tc.name, got, tc.want)
		}
	}
}

func TestCleanupIdentityRejectsUnsupportedState(t *testing.T) {
	for _, raw := range []string{
		`{"resources":[{"type":"pulumi:pulumi:Stack","outputs":{"eksCleanup":{"version":99}}}]}`,
		`{"resources":[{"type":"aws:eks/cluster:Cluster","urn":"a"},{"type":"aws:eks/cluster:Cluster","urn":"b"}]}`,
		`{"resources":[{"type":"aws:eks/cluster:Cluster","urn":"a","id":"x","outputs":{"arn":"not-an-arn"}}]}`,
	} {
		if _, _, err := cleanupIdentityFromCheckpoint([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestKubernetesTeardownTargets(t *testing.T) {
	const urn = "urn:pulumi:stackCreateEKS-eks::eks::aws:eks/cluster:Cluster::eks-cluster"
	for _, tt := range []struct {
		name string
		raw  string
		want []string
	}{
		{
			"resources and provider",
			`{"resources":[
				{"type":"kubernetes:helm.sh/v4:Chart","urn":"chart"},
				{"type":"kubernetes:apps/v1:Deployment","urn":"deployment"},
				{"type":"pulumi:providers:kubernetes","urn":"k8sProvider"},
				{"type":"aws:eks/cluster:Cluster","urn":"cluster"}
			]}`,
			[]string{"urn:pulumi:stackCreateEKS-eks::eks::kubernetes:**", "urn:pulumi:stackCreateEKS-eks::eks::pulumi:providers:kubernetes:**"},
		},
		{
			"resources only",
			`{"resources":[{"type":"kubernetes:core/v1:ServiceAccount","urn":"sa"}]}`,
			[]string{"urn:pulumi:stackCreateEKS-eks::eks::kubernetes:**"},
		},
		{
			"provider only",
			`{"resources":[{"type":"pulumi:providers:kubernetes","urn":"k8sProvider"}]}`,
			[]string{"urn:pulumi:stackCreateEKS-eks::eks::pulumi:providers:kubernetes:**"},
		},
		{
			"aws only",
			`{"resources":[{"type":"aws:ec2/vpc:Vpc","urn":"vpc"}]}`,
			nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kubernetesTeardownTargets([]byte(tt.raw), urn)
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("targets=%v err=%v", got, err)
			}
		})
	}
	if got, err := kubernetesTeardownTargets([]byte(`{"resources":[]}`), ""); err != nil || got != nil {
		t.Fatalf("targets=%v err=%v", got, err)
	}
	if _, err := kubernetesTeardownTargets([]byte(`{"resources":[]}`), "not-a-urn"); err == nil {
		t.Fatal("accepted malformed cluster URN")
	}
}

func TestEKSTeardownOrderingAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, urn, failAt string
		want              []string
	}{
		{"normal", "cluster", "", []string{"cluster", "cleanup", "final"}},
		{"resumed", "", "", []string{"cleanup", "final"}},
		{"cluster-failure", "cluster", "cluster", []string{"cluster"}},
		{"cleanup-failure", "cluster", "cleanup", []string{"cluster", "cleanup"}},
		{"final-failure", "cluster", "final", []string{"cluster", "cleanup", "final"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			call := func(name string) error {
				calls = append(calls, name)
				if name == tt.failAt {
					return errors.New("injected failure")
				}
				return nil
			}
			err := runEKSTeardown(context.Background(), tt.urn, func(targets []string) error {
				if len(targets) > 0 {
					return call("cluster")
				}
				return call("final")
			}, func(context.Context) error { return call("cleanup") })
			if !reflect.DeepEqual(calls, tt.want) || (err != nil) != (tt.failAt != "") {
				t.Fatalf("calls=%v err=%v", calls, err)
			}
		})
	}
}

type fakeWorkspaceEnv map[string]string

func (f fakeWorkspaceEnv) SetEnvVar(key, value string) { f[key] = value }
func (f fakeWorkspaceEnv) UnsetEnvVar(key string)      { delete(f, key) }

func TestKubernetesTeardownAllowsUnreachableClusters(t *testing.T) {
	env := fakeWorkspaceEnv{}
	var setDuringDestroy bool
	err := withDeleteUnreachable(env, func() error {
		setDuringDestroy = env[deleteUnreachableEnv] == "true"
		return errors.New("destroy failed")
	})
	if err == nil || !setDuringDestroy {
		t.Fatalf("err=%v setDuringDestroy=%v", err, setDuringDestroy)
	}
	if _, exists := env[deleteUnreachableEnv]; exists {
		t.Fatal("deleteUnreachable env var was not unset after the phase")
	}
}
