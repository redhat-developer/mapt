package allocation

import (
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
		// Filter for current location the computesizes
		supportedComputeSizes, err := data.FilterComputeSizesByLocation(
			mCtx.Context(), args.Location, computeSizes)
		if err != nil {
			return nil, err
		}
		return &AllocationResult{
			ImageRef:     args.ImageRef,
			Location:     args.Location,
			ComputeSizes: supportedComputeSizes,
		}, nil

	}
}
