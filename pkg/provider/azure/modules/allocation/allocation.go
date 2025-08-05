package allocation

import (
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
)

type AllocationArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	OSType         string
	ImageRef       *data.ImageReference
	Location       *string
	// Need to move this to Context
	Spot                  bool
	SpotExcludedLocations []string
	SpotTolerance         *spot.Tolerance
}

type AllocationResult struct {
	// location and price (if Spot is enable)
	// Region        *string
	// AZ            *string
	// SpotPrice     *float64
	Location *string
	// ComputeType      string
	Price *float64
	// HostingPlace     string
	// AvailabilityZone string
	// ChanceLevel      int
	ComputeSizes []string
	ImageRef     *data.ImageReference
}

func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error) {
	var err error
	computeSizes := args.ComputeRequest.ComputeSizes
	if len(computeSizes) == 0 {
		computeSizes, err =
			data.NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	if args.Spot {
		sArgs := &data.SpotInfoArgs{
			ComputeSizes: computeSizes,
			OSType:       args.OSType,
		}
		if args.ImageRef != nil {
			sArgs.ImageRef = *args.ImageRef
		}
		bsc, err := data.SpotInfo(mCtx, sArgs)
		if err != nil {
			return nil, err
		}
		return &AllocationResult{
			ImageRef:     args.ImageRef,
			Location:     &bsc.HostingPlace,
			Price:        &bsc.Price,
			ComputeSizes: []string{bsc.ComputeType},
		}, nil

	} else {
		// Filter for current location the computesizes
		supportedComputeSizes, err := data.FilterComputeSizesByLocation(
			args.Location, computeSizes)
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
