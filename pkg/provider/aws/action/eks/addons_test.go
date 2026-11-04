package eks

import (
	"encoding/json"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestAddonTagConfiguration(t *testing.T) {
	tags := map[string]string{"build-id": "123", clusterOwnershipTag: "wrong"}
	for _, addon := range []string{"vpc-cni", "aws-ebs-csi-driver"} {
		encoded, err := addonConfiguration(addon, tags, "cluster-123")
		if err != nil {
			t.Fatal(err)
		}
		var config struct {
			Env        map[string]string `json:"env"`
			Controller struct {
				ExtraVolumeTags map[string]string `json:"extraVolumeTags"`
			} `json:"controller"`
			DefaultStorageClass struct {
				Enabled bool `json:"enabled"`
			} `json:"defaultStorageClass"`
		}
		if err := json.Unmarshal([]byte(encoded), &config); err != nil {
			t.Fatal(err)
		}
		got := config.Controller.ExtraVolumeTags
		if addon == "vpc-cni" {
			if err := json.Unmarshal([]byte(config.Env["ADDITIONAL_ENI_TAGS"]), &got); err != nil {
				t.Fatal(err)
			}
		} else if !config.DefaultStorageClass.Enabled {
			t.Fatal("default storage class was lost")
		}
		if got["build-id"] != "123" || got[clusterOwnershipTag] != "cluster-123" {
			t.Fatalf("incorrect ownership: %v", got)
		}
	}
	if tags[clusterOwnershipTag] != "wrong" {
		t.Fatal("input tags mutated")
	}
	// Add-ons without managed configuration must not receive an empty object.
	if values := addonConfigurationValues("kube-proxy", tags, pulumi.String("cluster-123").ToStringOutput()); values != nil {
		t.Fatal("kube-proxy received configuration values")
	}
}
