package allocation

import (
	"fmt"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
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
	if args.Spot != nil && args.Spot.Spot {
		sArgs := &data.SpotInfoArgs{
			ComputeSizes:          computeSizes,
			OSType:                args.OSType,
			SpotTolerance:         &args.Spot.Tolerance,
			ExcludedLocations:     args.Spot.ExcludedHostingPlaces,
			SpotPriceIncreaseRate: &args.Spot.IncreaseRate,
		}
		if args.ImageRef != nil {
			sArgs.ImageRef = args.ImageRef
		}
		bsc, err := data.SpotInfo(mCtx, sArgs)
		if err != nil {
			return nil, err
		}
		return &AllocationResult{
			ImageRef:     args.ImageRef,
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
		// Filter for current location the computesizes
		supportedComputeSizes, err := data.FilterComputeSizesByLocation(
			mCtx.Context(), location, computeSizes)
		if err != nil {
			return nil, err
		}
		if len(supportedComputeSizes) == 0 {
			return nil, fmt.Errorf("no compute sizes available for location %q", *location)
		}
		return &AllocationResult{
			ImageRef:     args.ImageRef,
			Location:     location,
			ComputeSizes: supportedComputeSizes,
		}, nil

	}
}
