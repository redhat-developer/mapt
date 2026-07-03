package profile

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	nvidiaNamespace   = "nvidia-gpu-operator"
	nvidiaOGName      = "nvidia-gpu-operator-group"
	nvidiaPackageName = "gpu-operator-certified"
	nvidiaAPIVersion  = "nvidia.com/v1"
	nvidiaKind        = "ClusterPolicy"
	nvidiaCRName      = "gpu-cluster-policy"
)

var (
	clusterPolicyGVR = schema.GroupVersionResource{
		Group:    "nvidia.com",
		Version:  "v1",
		Resource: "clusterpolicies",
	}

	// clusterPolicySpec is the default ClusterPolicy spec for OpenShift.
	// An empty spec causes the operator to fail; these are the recommended
	// defaults (CRI-O runtime, OCP driver toolkit).
	clusterPolicySpec = map[string]interface{}{
		"operator": map[string]interface{}{
			"use_ocp_driver_toolkit": true,
		},
		"driver": map[string]interface{}{
			"enabled": true,
			"upgradePolicy": map[string]interface{}{
				"autoUpgrade": true,
			},
		},
		"toolkit": map[string]interface{}{
			"enabled": true,
		},
		"devicePlugin": map[string]interface{}{
			"enabled": true,
		},
		"dcgm": map[string]interface{}{
			"enabled": true,
		},
		"dcgmExporter": map[string]interface{}{
			"enabled": true,
		},
		"gfd": map[string]interface{}{
			"enabled": true,
		},
		"migManager": map[string]interface{}{
			"enabled": true,
		},
		"nodeStatusExporter": map[string]interface{}{
			"enabled": true,
		},
		"daemonsets": map[string]interface{}{
			"priorityClassName": "system-node-critical",
			"updateStrategy":    "RollingUpdate",
		},
	}
)

// deployNvidia installs the NFD operator (prerequisite), the NVIDIA GPU Operator,
// and creates a ClusterPolicy CR to configure the GPU stack.
func deployNvidia(ctx *pulumi.Context, args *DeployArgs) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-nvidia-%s", args.Prefix, suffix)
	}

	// NFD must be ready before the GPU operator can discover GPU nodes
	nfdReady, err := deployNFD(ctx, args)
	if err != nil {
		return nil, err
	}

	// Gate the GPU operator install on NFD readiness via the namespace name.
	// The namespace name won't resolve until NFD is ready, which delays
	// the operator install.
	nsName := pulumi.All(nfdReady).ApplyT(
		func(_ []interface{}) string {
			return nvidiaNamespace
		}).(pulumi.StringOutput)

	// Install the NVIDIA GPU Operator from the certified catalog
	csvReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      nvidiaNamespace,
		nsName:         nsName,
		ogName:         nvidiaOGName,
		ogTargetNS:     []string{nvidiaNamespace},
		subName:        nvidiaPackageName,
		packageName:    nvidiaPackageName,
		catalogSource:  catalogSourceCertified,
		csvPrefix:      nvidiaPackageName,
	})
	if err != nil {
		return nil, err
	}

	// Create ClusterPolicy CR after CSV is ready.
	cpName := csvReady.ApplyT(func(_ string) string {
		return nvidiaCRName
	}).(pulumi.StringOutput)

	cp, err := apiextensions.NewCustomResource(ctx, rn("clusterpolicy"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(nvidiaAPIVersion),
			Kind:       pulumi.String(nvidiaKind),
			Metadata: &metav1.ObjectMetaArgs{
				Name: cpName,
			},
			OtherFields: map[string]interface{}{
				"spec": clusterPolicySpec,
			},
		},
		args.k8sOpts()...)
	if err != nil {
		return nil, err
	}

	// Wait for ClusterPolicy to be ready (status.state == "ready")
	cpReady := pulumi.All(cp.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, clusterPolicyGVR,
				"", nvidiaCRName,
				"state", "", "ready", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for ClusterPolicy: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("clusterPolicyReady", cpReady)

	return cp, nil
}
