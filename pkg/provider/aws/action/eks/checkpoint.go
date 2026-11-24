package eks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// checkpointFields keeps each field as raw JSON so teardown decodes only the
// fields it needs and ignores the rest.
type checkpointFields map[string]json.RawMessage

func stringField(fields checkpointFields, key string) string {
	var value string
	_ = json.Unmarshal(fields[key], &value)
	return value
}

// kubernetesTeardownTargets returns wildcard targets for the Kubernetes resources
// and providers recorded in the checkpoint. Pulumi refuses to target the EKS
// cluster while a provider whose configuration depends on it is still in the
// stack (pulumi/pulumi#14650), so both are removed in earlier phases.
func kubernetesTeardownTargets(raw json.RawMessage, clusterURN string) ([]string, error) {
	prefix, err := resourceURNPrefix(clusterURN)
	if err != nil || prefix == "" {
		return nil, err
	}
	var deployment checkpointFields
	if err := json.Unmarshal(raw, &deployment); err != nil {
		return nil, err
	}
	var resources []checkpointFields
	if err := json.Unmarshal(deployment["resources"], &resources); err != nil {
		return nil, err
	}
	var hasResources, hasProvider bool
	for _, resource := range resources {
		switch resourceType := stringField(resource, "type"); {
		case resourceType == "pulumi:providers:kubernetes":
			hasProvider = true
		case strings.HasPrefix(resourceType, "kubernetes:"):
			hasResources = true
		}
	}
	var targets []string
	if hasResources {
		targets = append(targets, prefix+"kubernetes:**")
	}
	if hasProvider {
		targets = append(targets, prefix+"pulumi:providers:kubernetes:**")
	}
	return targets, nil
}

// resourceURNPrefix returns urn:pulumi:<stack>::<project>:: for a resource URN.
func resourceURNPrefix(resourceURN string) (string, error) {
	if resourceURN == "" {
		return "", nil
	}
	parts := strings.Split(resourceURN, "::")
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("unexpected resource URN: %q", resourceURN)
	}
	return parts[0] + "::" + parts[1] + "::", nil
}

func objectField(fields checkpointFields, key string) checkpointFields {
	var value checkpointFields
	_ = json.Unmarshal(fields[key], &value)
	return value
}

// createCompleted reports whether a create ever finished for the stack. mapt
// exports the kubeconfig only at the end of a successful program run, so a
// stack without that output never had a cluster to tear down.
func createCompleted(raw json.RawMessage) bool {
	var deployment checkpointFields
	if err := json.Unmarshal(raw, &deployment); err != nil {
		return false
	}
	var resources []checkpointFields
	if err := json.Unmarshal(deployment["resources"], &resources); err != nil {
		return false
	}
	for _, resource := range resources {
		if stringField(resource, "type") != "pulumi:pulumi:Stack" {
			continue
		}
		for key, value := range objectField(resource, "outputs") {
			if strings.HasSuffix(key, outputKubeconfig) && len(value) > 0 && string(value) != "null" {
				return true
			}
		}
	}
	return false
}

// cleanupIdentityFromCheckpoint returns the non-secret teardown identity and the
// EKS cluster URN recorded in the stack checkpoint. Pulumi saves stack outputs
// only when the program finishes, so a create killed after the cluster was
// created has no eksCleanup output; the identity is then read from the cluster
// resource. The checkpoint is never modified.
func cleanupIdentityFromCheckpoint(raw json.RawMessage) (*cleanupIdentity, string, error) {
	var deployment checkpointFields
	if err := json.Unmarshal(raw, &deployment); err != nil {
		return nil, "", err
	}
	var resources []checkpointFields
	if err := json.Unmarshal(deployment["resources"], &resources); err != nil {
		return nil, "", err
	}
	var identity *cleanupIdentity
	var cluster *checkpointFields
	for index, resource := range resources {
		if stringField(resource, "type") == "pulumi:pulumi:Stack" {
			if encoded := objectField(resource, "outputs")[outputCleanup]; len(encoded) > 0 {
				identity = &cleanupIdentity{}
				if err := json.Unmarshal(encoded, identity); err != nil {
					return nil, "", fmt.Errorf("reading cleanup identity: %w", err)
				}
			}
		}
		if stringField(resource, "type") == "aws:eks/cluster:Cluster" {
			if cluster != nil {
				return nil, "", fmt.Errorf("multiple EKS clusters in checkpoint")
			}
			cluster = &resources[index]
		}
	}
	clusterURN := ""
	if cluster != nil {
		clusterURN = stringField(*cluster, "urn")
		// No cluster ID means creation never completed; there are no controller
		// resources yet. A normal final destroy handles the partial network.
		if identity == nil && stringField(*cluster, "id") != "" {
			outputs := objectField(*cluster, "outputs")
			clusterARN, err := arn.Parse(stringField(outputs, "arn"))
			if err != nil {
				return nil, "", fmt.Errorf("recovering cluster ARN: %w", err)
			}
			logging.Debug("Recovering EKS teardown identity from the cluster resource")
			identity = &cleanupIdentity{
				Version: 1, AccountID: clusterARN.AccountID, Region: clusterARN.Region,
				ClusterName: stringField(*cluster, "id"),
				VPCID:       stringField(objectField(outputs, "vpcConfig"), "vpcId"),
			}
		}
	}
	if identity == nil {
		return nil, clusterURN, nil
	}
	if err := identity.validate(); err != nil {
		return nil, "", err
	}
	return identity, clusterURN, nil
}
