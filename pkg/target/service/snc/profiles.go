package snc

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
	ProfileServiceMesh        = "servicemesh"
	ProfileOpenShiftAI        = "ai"
)

// validProfiles is the single source of truth for supported profile names.
var validProfiles = []string{
	ProfileVirtualization,
	ProfileServerlessServing, ProfileServerlessEventing, ProfileServerless,
	ProfileServiceMesh,
	ProfileOpenShiftAI,
}

// ProfileDeployArgs holds the arguments needed by a profile to deploy
// its resources on the SNC cluster.
type ProfileDeployArgs struct {
	K8sProvider *kubernetes.Provider
	Kubeconfig  pulumi.StringOutput
	Prefix      string
	Deps        []pulumi.Resource
}

// ValidateProfiles checks that all requested profiles are supported and
// that there are no incompatible combinations.
func ValidateProfiles(profiles []string) error {
	for _, p := range profiles {
		if !slices.Contains(validProfiles, p) {
			return fmt.Errorf("profile %q is not supported for SNC. Supported profiles: %v", p, validProfiles)
		}
	}
	// AI uses Service Mesh v2 (Maistra); the servicemesh profile deploys v3 (Sail).
	// Both target istio-system and are incompatible on the same cluster.
	if slices.Contains(profiles, ProfileOpenShiftAI) && slices.Contains(profiles, ProfileServiceMesh) {
		return fmt.Errorf("profiles %q and %q cannot be combined: AI requires Service Mesh v2 while the servicemesh profile deploys v3",
			ProfileOpenShiftAI, ProfileServiceMesh)
	}
	return nil
}

// DeployProfiles deploys all requested profiles on the SNC cluster.
// It ensures shared dependencies (e.g. the Serverless operator) are only
// installed once, even when multiple profiles require them.
// The AI profile implicitly brings in Service Mesh v2 (Maistra) and
// serverless-serving as prerequisites for Kserve.
func DeployProfiles(ctx *pulumi.Context, profiles []string, args *ProfileDeployArgs) error {
	needVirtualization := false
	needServing := false
	needEventing := false
	needServiceMesh := false
	needAI := false

	for _, p := range profiles {
		switch p {
		case ProfileVirtualization:
			needVirtualization = true
		case ProfileServerlessServing:
			needServing = true
		case ProfileServerlessEventing:
			needEventing = true
		case ProfileServerless:
			needServing = true
			needEventing = true
		case ProfileServiceMesh:
			needServiceMesh = true
		case ProfileOpenShiftAI:
			needAI = true
			// AI requires serverless-serving for Kserve
			needServing = true
		default:
			return fmt.Errorf("profile %q has no deploy function", p)
		}
	}

	if needVirtualization {
		if _, err := deployVirtualization(ctx, args); err != nil {
			return err
		}
	}

	// Collect readiness outputs from prerequisite profiles so that
	// dependent profiles (e.g. AI) can wait for them.
	var aiPrereqs []pulumi.StringOutput

	if needServiceMesh {
		if _, _, err := deployServiceMesh(ctx, args); err != nil {
			return err
		}
	}

	// AI requires Service Mesh v2 (Maistra) — separate from the v3 (Sail) profile
	if needAI {
		_, smcpReady, err := deployServiceMeshV2(ctx, args)
		if err != nil {
			return err
		}
		aiPrereqs = append(aiPrereqs, smcpReady)
	}

	if needServing || needEventing {
		operatorReady, err := deployServerlessOperator(ctx, args)
		if err != nil {
			return err
		}
		if needServing {
			_, ksReady, err := deployKnativeServing(ctx, args, operatorReady)
			if err != nil {
				return err
			}
			if needAI {
				aiPrereqs = append(aiPrereqs, ksReady)
			}
		}
		if needEventing {
			if _, err := deployKnativeEventing(ctx, args, operatorReady); err != nil {
				return err
			}
		}
	}

	if needAI {
		if _, err := deployOpenShiftAI(ctx, args, aiPrereqs); err != nil {
			return err
		}
	}

	return nil
}

// ProfilesRequireNestedVirt returns true if any of the given profiles
// requires nested virtualization on the compute instance.
func ProfilesRequireNestedVirt(profiles []string) bool {
	return slices.Contains(profiles, ProfileVirtualization)
}

// ProfilesMinCPUs returns the minimum number of CPUs required by the
// given set of profiles. If no profile needs extra resources it returns 0
// (meaning "use the default").
func ProfilesMinCPUs(profiles []string) int32 {
	if slices.Contains(profiles, ProfileOpenShiftAI) {
		return 16
	}
	return 0
}
