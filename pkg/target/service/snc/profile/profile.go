package profile

import (
	"fmt"
	"slices"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	ProfileVirtualization     = "virtualization"
	ProfileServerlessServing  = "serverless-serving"
	ProfileServerlessEventing = "serverless-eventing"
	ProfileServerless         = "serverless"
)

// validProfiles is the single source of truth for supported profile names.
var validProfiles = []string{ProfileVirtualization, ProfileServerlessServing, ProfileServerlessEventing, ProfileServerless}

// DeployArgs holds the arguments needed by a profile to deploy
// its resources on the SNC cluster.
type DeployArgs struct {
	K8sProvider *kubernetes.Provider
	Kubeconfig  pulumi.StringOutput
	Prefix      string
	Deps        []pulumi.Resource
}

// Validate checks that all requested profiles are supported.
func Validate(profiles []string) error {
	for _, p := range profiles {
		if !slices.Contains(validProfiles, p) {
			return fmt.Errorf("profile %q is not supported for SNC. Supported profiles: %v", p, validProfiles)
		}
	}
	return nil
}

// Deploy deploys all requested profiles on the SNC cluster.
// It ensures shared dependencies (e.g. the Serverless operator) are only
// installed once, even when multiple serverless profiles are requested.
func Deploy(ctx *pulumi.Context, profiles []string, args *DeployArgs) error {
	needServing := false
	needEventing := false

	for _, p := range profiles {
		switch p {
		case ProfileVirtualization:
			if _, err := deployVirtualization(ctx, args); err != nil {
				return err
			}
		case ProfileServerlessServing:
			needServing = true
		case ProfileServerlessEventing:
			needEventing = true
		case ProfileServerless:
			needServing = true
			needEventing = true
		default:
			return fmt.Errorf("profile %q has no deploy function", p)
		}
	}

	if needServing || needEventing {
		operatorReady, err := deployServerlessOperator(ctx, args)
		if err != nil {
			return err
		}
		if needServing {
			if _, err := deployKnativeServing(ctx, args, operatorReady); err != nil {
				return err
			}
		}
		if needEventing {
			if _, err := deployKnativeEventing(ctx, args, operatorReady); err != nil {
				return err
			}
		}
	}

	return nil
}

// RequireNestedVirt returns true if any of the given profiles
// requires nested virtualization on the compute instance.
func RequireNestedVirt(profiles []string) bool {
	return slices.Contains(profiles, ProfileVirtualization)
}

// NewK8sProvider creates a Pulumi Kubernetes provider from a kubeconfig string output.
func NewK8sProvider(ctx *pulumi.Context, name string, kubeconfig pulumi.StringOutput) (*kubernetes.Provider, error) {
	return kubernetes.NewProvider(ctx, name, &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	})
}
