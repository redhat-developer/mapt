package snc

import (
	"fmt"
	"slices"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	ProfileVirtualization = "virtualization"
	ProfileServiceMesh    = "servicemesh"
)

// validProfiles is the single source of truth for supported profile names.
var validProfiles = []string{ProfileVirtualization, ProfileServiceMesh}

// ProfileDeployArgs holds the arguments needed by a profile to deploy
// its resources on the SNC cluster.
type ProfileDeployArgs struct {
	K8sProvider *kubernetes.Provider
	Kubeconfig  pulumi.StringOutput
	Prefix      string
	Deps        []pulumi.Resource
}

// ValidateProfiles checks that all requested profiles are supported.
func ValidateProfiles(profiles []string) error {
	for _, p := range profiles {
		if !slices.Contains(validProfiles, p) {
			return fmt.Errorf("profile %q is not supported for SNC. Supported profiles: %v", p, validProfiles)
		}
	}
	return nil
}

// DeployProfile deploys the resources for a given profile on the SNC cluster.
// It returns the last resource created for dependency chaining.
func DeployProfile(ctx *pulumi.Context, profile string, args *ProfileDeployArgs) (pulumi.Resource, error) {
	switch profile {
	case ProfileVirtualization:
		return deployVirtualization(ctx, args)
	case ProfileServiceMesh:
		return deployServiceMesh(ctx, args)
	default:
		return nil, fmt.Errorf("profile %q has no deploy function", profile)
	}
}

// ProfilesRequireNestedVirt returns true if any of the given profiles
// requires nested virtualization on the compute instance.
func ProfilesRequireNestedVirt(profiles []string) bool {
	return slices.Contains(profiles, ProfileVirtualization)
}
