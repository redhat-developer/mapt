package eks

import (
	"encoding/json"
	"maps"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// clusterOwnershipTag carries the cluster name as its value. The key stays
// fixed because the resources that need it cannot take a cluster-specific key:
// the EBS CSI add-on rejects controller.k8sTagClusterId, and the driver rejects
// kubernetes.io-prefixed entries in extraVolumeTags.
const clusterOwnershipTag = "mapt-cluster-name"

func controllerTags(tags map[string]string, cluster string) map[string]string {
	result := make(map[string]string, len(tags)+1)
	maps.Copy(result, tags)
	// This tag is reserved for exact cluster ownership, including teardown.
	result[clusterOwnershipTag] = cluster
	return result
}

func addonConfiguration(addon string, tags map[string]string, cluster string) (string, error) {
	configuration := map[string]interface{}{}
	switch addon {
	case "vpc-cni":
		eniTags, err := json.Marshal(controllerTags(tags, cluster))
		if err != nil {
			return "", err
		}
		configuration["env"] = map[string]string{"ADDITIONAL_ENI_TAGS": string(eniTags)}
	case "aws-ebs-csi-driver":
		configuration["defaultStorageClass"] = map[string]bool{"enabled": true}
		configuration["controller"] = map[string]interface{}{"extraVolumeTags": controllerTags(tags, cluster)}
	}
	encoded, err := json.Marshal(configuration)
	return string(encoded), err
}

// addonConfigurationValues returns nil for add-ons that do not need managed
// configuration, so their add-on schema does not receive an empty object.
func addonConfigurationValues(addon string, tags map[string]string, cluster pulumi.StringOutput) pulumi.StringPtrInput {
	switch addon {
	case "vpc-cni", "aws-ebs-csi-driver":
	default:
		return nil
	}
	return cluster.ApplyT(func(name string) (string, error) {
		return addonConfiguration(addon, tags, name)
	}).(pulumi.StringOutput)
}
