package rhelai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
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

	// Marketplace image coordinates
	marketplacePublisher     = "RedHat"
	marketplaceOffer         = "rh-rhel-ai"
	marketplacePlanPublisher = "redhat"
	// SKU pattern: rh-rhelai-{nvidia|amd}-{N}gpu (gen2 handled by SkuG2Support)
	marketplaceSkuRegex = "rh-rhelai-%s-%dgpu"

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

var acceleratorToMarketplace = map[string]string{
	"cuda": "nvidia",
	"rocm": "amd",
}

var validMarketplaceGPUCounts = map[int32]bool{1: true, 2: true, 4: true, 8: true}

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
	// Shallow-copy to avoid mutating the caller's ComputeRequestArgs.
	computeReq := *args.ComputeRequest
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
	imageRef, err := resolveImageSource(args, &computeReq)
	if err != nil {
		return err
	}
	azureLinuxRequest :=
		&azureLinux.LinuxArgs{
			Prefix:           args.Prefix,
			ComputeRequest:   &computeReq,
			Spot:             args.Spot,
			ImageRef:         imageRef,
			Username:         username,
			ReadinessCommand: command.CommandPing}
	if err = azureLinux.Create(mCtxArgs, azureLinuxRequest); err != nil {
		if args.Marketplace && imageRef.Plan != nil &&
			(strings.Contains(err.Error(), "ResourcePurchaseValidationFailed") ||
				strings.Contains(err.Error(), "MarketplacePurchaseEligibilityFailed")) {
			return fmt.Errorf("RHEL AI marketplace: terms not accepted; run: az vm image terms accept --publisher %s --offer %s --plan %s\n%w",
				imageRef.Plan.Publisher, marketplaceOffer, imageRef.Plan.Name, err)
		}
		if len(computeReq.ComputeSizes) == 0 {
			return fmt.Errorf("RHEL AI: failed to provision a GPU-capable instance (ND/NC-series required for vllm); verify GPU quota in the target location/subscription: %w", err)
		}
	}
	return err
}

func resolveImageSource(args *apiRHELAI.RHELAIArgs, computeReq *cr.ComputeRequestArgs) (*data.ImageReference, error) {
	if args.Marketplace {
		gpus := computeReq.GPUs
		if !validMarketplaceGPUCounts[gpus] {
			return nil, fmt.Errorf("RHEL AI marketplace: --gpus must be 1, 2, 4, or 8 (got %d)", gpus)
		}
		accName, ok := acceleratorToMarketplace[strings.ToLower(args.Accelerator)]
		if !ok {
			return nil, fmt.Errorf("RHEL AI marketplace: unsupported accelerator %q (expected cuda or rocm)", args.Accelerator)
		}
		sku := fmt.Sprintf(marketplaceSkuRegex, accName, gpus)
		return &data.ImageReference{
			Publisher: marketplacePublisher,
			Offer:     marketplaceOffer,
			Sku:       sku,
			Plan: &data.MarketplacePlan{
				Name:      sku,
				Product:   marketplaceOffer,
				Publisher: marketplacePlanPublisher,
			},
		}, nil
	}
	if args.CustomImage != "" {
		return &data.ImageReference{
			SharedImageID:      imageIdFromName(args.CustomImage),
			DiskControllerType: "SCSI",
		}, nil
	}
	return &data.ImageReference{
		SharedImageID:      imageId(args.Accelerator, args.Version),
		DiskControllerType: "SCSI",
	}, nil
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
