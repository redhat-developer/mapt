package allocation

import (
	"fmt"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	"github.com/redhat-developer/mapt/pkg/util"
	"golang.org/x/exp/slices"
)

var ErrNoSupportedInstaceTypes = fmt.Errorf("the current region does not support any of the requested instance types")

type AllocationArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	Prefix,
	AMIProductDescription,
	AMIName *string
	Spot  *spotTypes.SpotArgs
	VpcID *string
}

type AllocationResult struct {
	Region        *string
	AZ            *string
	SpotPrice     *float64
	InstanceTypes []string
}

func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error) {
	var err error
	instancesTypes := args.ComputeRequest.ComputeSizes
	if len(instancesTypes) == 0 {
		instancesTypes, err =
			data.NewComputeSelector().Select(mCtx.Context(), args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	var allowedAZs []string
	if args.VpcID != nil {
		region := mCtx.TargetHostingPlace()
		allowedAZs, err = data.GetSubnetAZsForVPC(mCtx.Context(), region, *args.VpcID)
		if err != nil {
			return nil, err
		}
	}
	if args.Spot != nil && args.Spot.Spot {
		sr := &spot.SpotStackArgs{
			Prefix:       *args.Prefix,
			InstaceTypes: instancesTypes,
			Spot:         args.Spot,
			AllowedAZs:   allowedAZs,
		}
		if args.AMIName != nil {
			sr.AMIName = *args.AMIName
		}
		if args.AMIProductDescription != nil {
			sr.ProductDescription = *args.AMIProductDescription
		}
		return allocationSpot(mCtx, sr)
	}
	return allocationOnDemand(mCtx, instancesTypes, allowedAZs)
}

func allocationSpot(mCtx *mc.Context,
	args *spot.SpotStackArgs) (*AllocationResult, error) {
	so, err := spot.Create(mCtx, args)
	if err != nil {
		return nil, err
	}
	return &AllocationResult{
		Region:        &so.Region,
		AZ:            &so.AvailabilityZone,
		SpotPrice:     &so.Price,
		InstanceTypes: so.InstanceType,
	}, nil
}

func allocationOnDemand(mCtx *mc.Context, instancesTypes []string, allowedAZs []string) (*AllocationResult, error) {
	region := mCtx.TargetHostingPlace()
	candidateAZs := allowedAZs
	if len(candidateAZs) == 0 {
		candidateAZs = data.GetAvailabilityZones(mCtx.Context(), region, nil)
	}
	excludedAZs := []string{}
	var err error
	var az *string
	var supportedInstancesType []string
	for {
		remaining := util.ArrayFilter(candidateAZs, func(a string) bool {
			return !slices.Contains(excludedAZs, a)
		})
		if len(remaining) == 0 {
			return nil, ErrNoSupportedInstaceTypes
		}
		azName := remaining[util.Random(len(remaining)-1, 0)]
		az = &azName
		supportedInstancesType, err =
			data.FilterInstaceTypesOfferedByLocation(mCtx.Context(), instancesTypes, &data.LocationArgs{
				Region: &region,
				Az:     az,
			})
		if err != nil {
			return nil, err
		}
		if len(supportedInstancesType) > 0 {
			break
		}
		excludedAZs = append(excludedAZs, *az)
	}
	return &AllocationResult{
		Region:        &region,
		AZ:            az,
		InstanceTypes: supportedInstancesType,
	}, nil
}
