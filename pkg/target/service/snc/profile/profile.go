package profile

import (
	"fmt"
	"maps"
	"slices"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	ProfileVirtualization     = "virtualization"
	ProfileServerlessServing  = "serverless-serving"
	ProfileServerlessEventing = "serverless-eventing"
	ProfileServerless         = "serverless"
	ProfileServiceMesh        = "servicemesh"
)

// profileEffect describes what a profile requires when deployed.
type profileEffect struct {
	nestedVirt bool
	serving    bool
	eventing   bool
	deployFn   func(ctx *pulumi.Context, args *DeployArgs) (pulumi.Resource, error)
}

// profileRegistry is the single source of truth for supported profiles
// and their effects.
var profileRegistry = map[string]profileEffect{
	ProfileVirtualization:     {nestedVirt: true, deployFn: deployVirtualization},
	ProfileServerlessServing:  {serving: true},
	ProfileServerlessEventing: {eventing: true},
	ProfileServerless:         {serving: true, eventing: true},
	ProfileServiceMesh:        {deployFn: deployServiceMesh},
}

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
		if _, ok := profileRegistry[p]; !ok {
			return fmt.Errorf("profile %q is not supported for SNC. Supported profiles: %v",
				p, slices.Sorted(maps.Keys(profileRegistry)))
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
		effect := profileRegistry[p]
		if effect.deployFn != nil {
			if _, err := effect.deployFn(ctx, args); err != nil {
				return err
			}
		}
		needServing = needServing || effect.serving
		needEventing = needEventing || effect.eventing
	}

	if needServing || needEventing {
		if err := deployServerless(ctx, args, needServing, needEventing); err != nil {
			return err
		}
	}

	return nil
}

// RequireNestedVirt returns true if any of the given profiles
// requires nested virtualization on the compute instance.
func RequireNestedVirt(profiles []string) bool {
	for _, p := range profiles {
		if effect, ok := profileRegistry[p]; ok && effect.nestedVirt {
			return true
		}
	}
	return false
}

// NewK8sProvider creates a Pulumi Kubernetes provider from a kubeconfig string output.
func NewK8sProvider(ctx *pulumi.Context, name string, kubeconfig pulumi.StringOutput) (*kubernetes.Provider, error) {
	return kubernetes.NewProvider(ctx, name, &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	})
}
