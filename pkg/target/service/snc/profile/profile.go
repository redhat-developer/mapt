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
	ProfileOpenShiftAI        = "ai"
)

// profileEffect describes what a profile requires when deployed.
type profileEffect struct {
	nestedVirt bool
	serving    bool
	eventing   bool
	minCPUs    int32
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
	ProfileOpenShiftAI:        {serving: true, minCPUs: 16},
}

// incompatibleProfiles lists pairs of profiles that cannot be combined.
var incompatibleProfiles = [][2]string{
	// AI uses Service Mesh v2 (Maistra); the servicemesh profile deploys v3 (Sail).
	// Both target istio-system and are incompatible on the same cluster.
	{ProfileOpenShiftAI, ProfileServiceMesh},
}

// DeployArgs holds the arguments needed by a profile to deploy
// its resources on the SNC cluster.
type DeployArgs struct {
	K8sProvider *kubernetes.Provider
	Kubeconfig  pulumi.StringOutput
	Prefix      string
	Deps        []pulumi.Resource
}

// Validate checks that all requested profiles are supported and
// that there are no incompatible combinations.
func Validate(profiles []string) error {
	for _, p := range profiles {
		if _, ok := profileRegistry[p]; !ok {
			return fmt.Errorf("profile %q is not supported for SNC. Supported profiles: %v",
				p, slices.Sorted(maps.Keys(profileRegistry)))
		}
	}
	for _, pair := range incompatibleProfiles {
		if slices.Contains(profiles, pair[0]) && slices.Contains(profiles, pair[1]) {
			return fmt.Errorf("profiles %q and %q cannot be combined", pair[0], pair[1])
		}
	}
	return nil
}

// Deploy deploys all requested profiles on the SNC cluster.
// It ensures shared dependencies (e.g. the Serverless operator) are only
// installed once, even when multiple profiles require them.
// The AI profile implicitly brings in Service Mesh v2 (Maistra) and
// serverless-serving as prerequisites for Kserve.
func Deploy(ctx *pulumi.Context, profiles []string, args *DeployArgs) error {
	needServing := false
	needEventing := false
	needAI := false

	for _, p := range profiles {
		effect := profileRegistry[p]
		if effect.deployFn != nil {
			if _, err := effect.deployFn(ctx, args); err != nil {
				return err
			}
		}
		needServing = needServing || effect.serving
		needEventing = needEventing || effect.eventing
		if p == ProfileOpenShiftAI {
			needAI = true
		}
	}

	// Collect readiness outputs from prerequisite profiles so that
	// dependent profiles (e.g. AI) can wait for them.
	var aiPrereqs []pulumi.StringOutput

	// AI requires Service Mesh v2 (Maistra) — separate from the v3 (Sail) profile
	if needAI {
		_, smcpReady, err := deployServiceMeshV2(ctx, args)
		if err != nil {
			return err
		}
		aiPrereqs = append(aiPrereqs, smcpReady)
	}

	if needServing || needEventing {
		if err := deployServerlessWithPrereqs(ctx, args, needServing, needEventing, needAI, &aiPrereqs); err != nil {
			return err
		}
	}

	if needAI {
		if _, err := deployOpenShiftAI(ctx, args, aiPrereqs); err != nil {
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

// MinCPUs returns the minimum number of CPUs required by the
// given set of profiles. If no profile needs extra resources it returns 0
// (meaning "use the default").
func MinCPUs(profiles []string) int32 {
	var max int32
	for _, p := range profiles {
		if effect, ok := profileRegistry[p]; ok && effect.minCPUs > max {
			max = effect.minCPUs
		}
	}
	return max
}

// NewK8sProvider creates a Pulumi Kubernetes provider from a kubeconfig string output.
func NewK8sProvider(ctx *pulumi.Context, name string, kubeconfig pulumi.StringOutput) (*kubernetes.Provider, error) {
	return kubernetes.NewProvider(ctx, name, &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	})
}
