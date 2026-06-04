package allocation

import (
	"fmt"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type AllocationArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	OSType         string
	ImageRef       *data.ImageReference
	Location       *string
	Spot           *spotTypes.SpotArgs
}

type AllocationResult struct {
	Location     *string
	Price        *float64
	ComputeSizes []string
	ImageRef     *data.ImageReference
}

func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error) {
	var err error
	computeSizes := args.ComputeRequest.ComputeSizes
	if len(computeSizes) == 0 {
		computeSizes, err =
			data.NewComputeSelector().Select(mCtx.Context(), args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}

	// Derive the effective image reference. For shared gallery images, resolveImageRef
	// queries the gallery API and sets DiskControllerType when the image supports
	// exactly one type; the caller's explicit value is preserved in all other cases.
	ir := resolveImageRef(mCtx, args.ImageRef)

	diskControllerType := ""
	if ir != nil {
		diskControllerType = ir.DiskControllerType
	}

	if args.Spot != nil && args.Spot.Spot {
		sArgs := &data.SpotInfoArgs{
			ComputeSizes:          computeSizes,
			OSType:                args.OSType,
			SpotTolerance:         &args.Spot.Tolerance,
			ExcludedLocations:     args.Spot.ExcludedHostingPlaces,
			SpotPriceIncreaseRate: &args.Spot.IncreaseRate,
		}
		if ir != nil {
			sArgs.ImageRef = ir
		}
		bsc, err := data.SpotInfo(mCtx, sArgs)
		if err != nil {
			return nil, err
		}
		// Filter the spot-selected sizes by disk controller type compatibility.
		if diskControllerType != "" {
			spotLocation := bsc.HostingPlace
			bsc.ComputeType, err = data.FilterComputeSizesByDiskControllerType(
				mCtx.Context(), &spotLocation, bsc.ComputeType, diskControllerType)
			if err != nil {
				return nil, err
			}
			if len(bsc.ComputeType) == 0 {
				return nil, fmt.Errorf(
					"spot compute sizes in location %q do not support disk controller type %q required by the selected image",
					bsc.HostingPlace, diskControllerType)
			}
		}
		return &AllocationResult{
			ImageRef:     ir,
			Location:     &bsc.HostingPlace,
			Price:        &bsc.Price,
			ComputeSizes: bsc.ComputeType,
		}, nil

	} else {
		location := args.Location
		if location == nil || *location == "" {
			hp := mCtx.TargetHostingPlace()
			if hp == "" {
				return nil, fmt.Errorf("location is required for non-spot allocation: set ARM_LOCATION_NAME or AZURE_DEFAULTS_LOCATION environment variable")
			}
			location = &hp
		}
		// Single SKU enumeration: filter by location and disk controller type together.
		// When diskControllerType is empty the type check is a no-op (any type passes).
		supportedComputeSizes, err := data.FilterComputeSizesByDiskControllerType(
			mCtx.Context(), location, computeSizes, diskControllerType)
		if err != nil {
			return nil, err
		}
		if len(supportedComputeSizes) == 0 {
			if diskControllerType != "" {
				return nil, fmt.Errorf(
					"no compute sizes in location %q support disk controller type %q required by the selected image",
					*location, diskControllerType)
			}
			return nil, fmt.Errorf("no compute sizes available for location %q", *location)
		}
		return &AllocationResult{
			ImageRef:     ir,
			Location:     location,
			ComputeSizes: supportedComputeSizes,
		}, nil
	}
}

// resolveImageRef returns a copy of the image reference, optionally enriched with the
// disk controller type read from the gallery image definition. The gallery Features value
// lists types the image *supports*, not what it *requires*. We only override the caller's
// value when the gallery returns exactly one supported type — that uniquely identifies the
// requirement. When the gallery returns multiple types the image is flexible, so the
// caller's explicit value (if any) is preserved unchanged. On fetch failure the caller's
// value is also preserved.
func resolveImageRef(mCtx *mc.Context, ir *data.ImageReference) *data.ImageReference {
	if ir == nil || ir.SharedImageID == "" {
		return ir
	}
	enriched := *ir
	types, err := data.GetSharedImageDiskControllerTypes(mCtx.Context(), &ir.SharedImageID)
	if err != nil {
		logging.Debugf("could not fetch disk controller types for image %s: %v", ir.SharedImageID, err)
		return &enriched
	}
	if len(types) == 1 && enriched.DiskControllerType == "" {
		enriched.DiskControllerType = types[0]
	}
	return &enriched
}
