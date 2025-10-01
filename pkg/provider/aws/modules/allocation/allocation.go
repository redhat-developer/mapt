package allocation

import (
	"fmt"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
)

var ErrNoSupportedInstaceTypes = fmt.Errorf("the current region does not support any of the requested instance types")

type AllocationArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	Prefix,
	AMIProductDescription,
	AMIName *string
	Spot *spotTypes.SpotArgs
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
			data.NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	if args.Spot != nil && args.Spot.Spot {
		sr := &spot.SpotStackArgs{
			Prefix:       *args.Prefix,
			InstaceTypes: instancesTypes,
			Spot:         args.Spot,
		}
		if args.AMIName != nil {
			sr.AMIName = *args.AMIName
		}
		if args.AMIProductDescription != nil {
			sr.ProductDescription = *args.AMIProductDescription
		}
		return allocationSpot(mCtx, sr)
	}
	return allocationOnDemand(mCtx, instancesTypes)
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

func allocationOnDemand(mCtx *mc.Context, instancesTypes []string) (*AllocationResult, error) {
	region := mCtx.TargetHostingPlace()
	excludedAZs := []string{}
	var err error
	var az *string
	var supportedInstancesType []string
	azs := data.GetAvailabilityZones(region, nil)
	for {
		az, err = data.GetRandomAvailabilityZone(region, excludedAZs)
		if err != nil {
			return nil, err
		}
		supportedInstancesType, err =
			data.FilterInstaceTypesOfferedByLocation(instancesTypes, &data.LocationArgs{
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
		if len(excludedAZs) == len(azs) {
			return nil, ErrNoSupportedInstaceTypes
		}
	}
	return &AllocationResult{
		Region:        &region,
		AZ:            az,
		InstanceTypes: supportedInstancesType,
	}, nil
}
