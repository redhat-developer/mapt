package profile

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	ProfileVirtualization     = "virtualization"
	ProfileServerlessServing  = "serverless-serving"
	ProfileServerlessEventing = "serverless-eventing"
	ProfileServerless         = "serverless"
	ProfileServiceMesh        = "servicemesh"
	ProfileOpenShiftAI        = "ai"
	ProfileNvidia             = "nvidia"
)

// profileEffect describes what a profile requires when deployed.
type profileEffect struct {
	nestedVirt      bool
	serving         bool
	eventing        bool
	minCPUs         int32
	maxCPUs         int32
	gpuManufacturer string
	deployFn        func(ctx *pulumi.Context, args *DeployArgs) (pulumi.Resource, error)
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
	ProfileNvidia:             {gpuManufacturer: "NVIDIA", minCPUs: 8, maxCPUs: 32, deployFn: deployNvidia},
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
	// DeletedWith is the compute resource (EC2 instance or ASG) that hosts
	// the cluster. When set, K8s resources are marked with pulumi.DeletedWith
	// so that Pulumi skips deleting them individually during destroy — the
	// resources disappear when the VM is terminated.
	DeletedWith pulumi.Resource
	// OperatorChannels maps operator packageName to an OLM channel override.
	OperatorChannels map[string]string
	// CatalogSources maps operator packageName to a custom index image URL.
	CatalogSources map[string]string

	// catalogSourceCRs maps packageName to the CatalogSource CR info.
	catalogSourceCRs map[string]catalogSourceInfo
}

type catalogSourceInfo struct {
	Name     string
	Resource pulumi.Resource
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
	if err := args.ensureCatalogSources(ctx); err != nil {
		return err
	}

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

// MaxCPUs returns the maximum number of CPUs allowed by the
// given set of profiles. If no profile sets a cap it returns 0
// (meaning "no upper limit").
func MaxCPUs(profiles []string) int32 {
	var max int32
	for _, p := range profiles {
		if effect, ok := profileRegistry[p]; ok && effect.maxCPUs > max {
			max = effect.maxCPUs
		}
	}
	return max
}

// GPUManufacturer returns the GPU manufacturer required by the given set
// of profiles. If no profile needs a GPU it returns an empty string.
func GPUManufacturer(profiles []string) string {
	for _, p := range profiles {
		if effect, ok := profileRegistry[p]; ok && effect.gpuManufacturer != "" {
			return effect.gpuManufacturer
		}
	}
	return ""
}

// newNamespace creates (or adopts) a Kubernetes namespace using server-side
// apply so the call succeeds even if the namespace already exists.
func (a *DeployArgs) newNamespace(ctx *pulumi.Context, name string, nsName pulumi.StringInput, extra ...pulumi.ResourceOption) (*corev1.NamespacePatch, error) {
	return corev1.NewNamespacePatch(ctx, name,
		&corev1.NamespacePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: nsName,
			},
		},
		a.k8sOpts(extra...)...)
}

func ValidateOperatorOverrides(channels, catalogs map[string]string) error {
	for pkg, ch := range channels {
		if pkg == "" || ch == "" {
			return fmt.Errorf("invalid --operator-channel: both package name and channel must be non-empty (got %q=%q)", pkg, ch)
		}
	}
	for pkg, img := range catalogs {
		if pkg == "" || img == "" {
			return fmt.Errorf("invalid --catalog-source: both package name and index image must be non-empty (got %q=%q)", pkg, img)
		}
	}
	return nil
}

// ensureCatalogSources creates CatalogSource CRs for any custom index images
// specified via --catalog-source, so that operator subscriptions can reference them.
func (a *DeployArgs) ensureCatalogSources(ctx *pulumi.Context) error {
	if len(a.CatalogSources) == 0 {
		return nil
	}
	a.catalogSourceCRs = make(map[string]catalogSourceInfo, len(a.CatalogSources))
	for pkg, indexImage := range a.CatalogSources {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(indexImage)))[:8]
		csName := fmt.Sprintf("mapt-cs-%s-%s", pkg, hash)
		cs, err := apiextensions.NewCustomResource(ctx, csName,
			&apiextensions.CustomResourceArgs{
				ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
				Kind:       pulumi.String("CatalogSource"),
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(csName),
					Namespace: pulumi.String("openshift-marketplace"),
				},
				OtherFields: map[string]interface{}{
					"spec": map[string]interface{}{
						"sourceType":  "grpc",
						"image":       indexImage,
						"displayName": fmt.Sprintf("MAPT custom catalog for %s", pkg),
						"publisher":   "MAPT",
					},
				},
			},
			a.k8sOpts(pulumi.DependsOn(a.Deps))...)
		if err != nil {
			return err
		}
		a.catalogSourceCRs[pkg] = catalogSourceInfo{Name: csName, Resource: cs}
	}
	return nil
}

// k8sOpts returns the common Pulumi resource options for K8s resources:
// the K8s provider and (when set) the DeletedWith option. Extra options
// (e.g. DependsOn) can be appended.
func (a *DeployArgs) k8sOpts(extra ...pulumi.ResourceOption) []pulumi.ResourceOption {
	opts := []pulumi.ResourceOption{pulumi.Provider(a.K8sProvider)}
	if a.DeletedWith != nil {
		opts = append(opts, pulumi.DeletedWith(a.DeletedWith))
	}
	return append(opts, extra...)
}

// NewK8sProvider creates a Pulumi Kubernetes provider from a kubeconfig string output.
func NewK8sProvider(ctx *pulumi.Context, name string, kubeconfig pulumi.StringOutput) (*kubernetes.Provider, error) {
	return kubernetes.NewProvider(ctx, name, &kubernetes.ProviderArgs{
		Kubeconfig:        kubeconfig,
		DeleteUnreachable: pulumi.Bool(true),
	})
}
