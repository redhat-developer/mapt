package rhelai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	apiRHELAI "github.com/redhat-developer/mapt/pkg/target/host/rhelai"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	imageOwnerSubscriptionId = "02db6bd4-035c-4074-b699-468f3d914744"
	imageOwnerResourceGroup  = "aipcc-productization"
	// $1 accelerator $2 version
	imageNameRegex = "rhel-ai-%s-azure-%s"
	// $1 subscriptionId $2 rgName $3 galleryName $4 imageName
	imageIdRegex = "/subscriptions/%s/resourceGroups/" + imageOwnerResourceGroup + "/providers/Microsoft.Compute/galleries/%s/images/%s/versions/1.0.0"

	username = "azureuser"
)

func imageIdFromName(imageName string) string {
	gName := strings.ReplaceAll(imageName, "-", "_")
	return fmt.Sprintf(imageIdRegex,
		imageOwnerSubscriptionId,
		gName,
		imageName)
}

func imageId(accelerator, version string) string {
	return imageIdFromName(fmt.Sprintf(imageNameRegex, accelerator, version))
}

// isGPUCapableSize returns true for ND-series and NC-series Azure VM sizes,
// which are the compute GPU families supported for RHEL AI workloads.
// NV-series (visualization GPUs) is intentionally excluded.
func isGPUCapableSize(vmSize string) bool {
	lower := strings.ToLower(vmSize)
	return strings.HasPrefix(lower, "standard_nd") || strings.HasPrefix(lower, "standard_nc")
}

func Create(mCtxArgs *maptContext.ContextArgs, args *apiRHELAI.RHELAIArgs) (err error) {
	if args == nil || args.ComputeRequest == nil {
		return fmt.Errorf("RHEL AI: args and ComputeRequest must not be nil")
	}
	logging.Debug("Creating RHEL AI Server")
	sharedImageID := imageId(args.Accelerator, args.Version)
	if args.CustomImage != "" {
		sharedImageID = imageIdFromName(args.CustomImage)
	}
	// Shallow-copy to avoid mutating the caller's ComputeRequestArgs.
	computeReq := *args.ComputeRequest
	if len(computeReq.ComputeSizes) > 0 {
		ctx := mCtxArgs.Context
		if ctx == nil {
			ctx = context.Background()
		}
		computeReq.ComputeSizes, err = data.FilterNoLocalStorageSizes(
			ctx, computeReq.ComputeSizes)
		if err != nil {
			return err
		}
		if len(computeReq.ComputeSizes) == 0 {
			return fmt.Errorf("no valid compute sizes: all provided sizes have NVMe-only local storage, incompatible with RHEL AI")
		}
	}
	// Ensure GPU-capable instance selection for auto-selection paths.
	if computeReq.GPUs == 0 {
		logging.Debug("RHEL AI: GPUs not set, defaulting to 1 for GPU-capable instance selection")
		computeReq.GPUs = 1
	}
	// All explicitly specified sizes must be GPU-capable; a single non-GPU entry
	// could get allocated and vllm would fail silently.
	for _, s := range computeReq.ComputeSizes {
		if !isGPUCapableSize(s) {
			return fmt.Errorf("RHEL AI: %q is not GPU-capable (expected ND-series or NC-series for vllm)", s)
		}
	}
	azureLinuxRequest :=
		&azureLinux.LinuxArgs{
			Prefix:         args.Prefix,
			ComputeRequest: &computeReq,
			Spot:           args.Spot,
			ImageRef: &data.ImageReference{
				SharedImageID: sharedImageID,
				// Belt-and-suspenders: set SCSI explicitly so Azure never infers a
				// conflicting default. resolveImageRef will also derive this from the
				// gallery image's Features, but the static value protects against API
				// failures or future images with multiple supported types.
				DiskControllerType: "SCSI",
			},
			Username:         username,
			ReadinessCommand: command.CommandPing}
	if err = azureLinux.Create(mCtxArgs, azureLinuxRequest); err != nil && len(computeReq.ComputeSizes) == 0 {
		return fmt.Errorf("RHEL AI: failed to provision a GPU-capable instance (ND/NC-series required for vllm); verify GPU quota in the target location/subscription: %w", err)
	}
	return err
}

func Destroy(mCtxArgs *maptContext.ContextArgs) error {
	return azureLinux.Destroy(mCtxArgs)
}

// ListVersions returns available RHEL AI version strings for the given accelerator,
// sorted in ascending order. Versions are derived from Azure Compute Gallery names
// in the image owner's subscription (e.g. gallery "rhel_ai_cuda_azure_3.4.0_ea.2"
// yields version "3.4.0-ea.2").
func ListVersions(ctx context.Context, accelerator string) ([]string, error) {
	acc := strings.ToLower(strings.TrimSpace(accelerator))
	switch acc {
	case "cuda", "rocm":
	default:
		return nil, fmt.Errorf("unsupported accelerator %q (expected: cuda or rocm)", accelerator)
	}
	prefix := fmt.Sprintf("rhel_ai_%s_azure_", strings.ReplaceAll(acc, "-", "_"))
	galleries, err := data.ListGalleriesByPrefix(ctx, imageOwnerSubscriptionId, imageOwnerResourceGroup, prefix)
	if err != nil {
		return nil, fmt.Errorf("listing RHEL AI versions for accelerator %q: %w", accelerator, err)
	}
	versions := make([]string, 0, len(galleries))
	for _, g := range galleries {
		raw := strings.TrimPrefix(g, prefix)
		versions = append(versions, strings.ReplaceAll(raw, "_", "-"))
	}
	sort.Strings(versions)
	return versions, nil
}
