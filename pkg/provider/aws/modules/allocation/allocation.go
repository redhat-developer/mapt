package allocation

import (
	"fmt"
	"os"

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
	return allocationOnDemand(instancesTypes)
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

func allocationOnDemand(instancesTypes []string) (*AllocationResult, error) {
	region := os.Getenv("AWS_DEFAULT_REGION")
	supportedInstancesType, err :=
		data.FilterInstaceTypesOfferedByRegion(instancesTypes, region)
	if err != nil {
		return nil, err
	}
	if len(supportedInstancesType) == 0 {
		return nil, ErrNoSupportedInstaceTypes
	}
	az, err := data.GetRandomAvailabilityZone(region, nil)
	if err != nil {
		return nil, err
	}
	return &AllocationResult{
		Region:        &region,
		AZ:            az,
		InstanceTypes: supportedInstancesType,
	}, nil
}
