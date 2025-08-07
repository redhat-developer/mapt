package allocation

import (
	"fmt"
	"os"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
)

type AllocationData struct {
	// location and price (if Spot is enable)
	Region        *string
	AZ            *string
	SpotPrice     *float64
	InstanceTypes []string
}

func AllocationDataOnSpot(mCtx *mc.Context, prefix, amiProductDescription, amiName *string, computeRequest *cr.ComputeRequestArgs) (*AllocationData, error) {
	var err error
	computeTypes := computeRequest.ComputeSizes
	if len(computeTypes) == 0 {
		computeTypes, err =
			data.NewComputeSelector().Select(computeRequest)
		if err != nil {
			return nil, err
		}
	}
	sr := spot.SpotOptionRequest{
		MCtx:               mCtx,
		Prefix:             *prefix,
		ProductDescription: *amiProductDescription,
		InstaceTypes:       computeTypes,
	}
	if amiName != nil {
		sr.AMIName = *amiName
	}
	so, err := sr.Create()
	if err != nil {
		return nil, err
	}
	availableInstaceTypes, err :=
		data.FilterInstaceTypesOfferedByRegion(computeTypes, so.Region)
	if err != nil {
		return nil, err
	}
	return &AllocationData{
		Region:        &so.Region,
		AZ:            &so.AvailabilityZone,
		SpotPrice:     &so.Price,
		InstanceTypes: availableInstaceTypes,
	}, nil
}

func AllocationDataOnDemand(mCtx *mc.Context, prefix, amiProductDescription, amiName *string, computeRequest *cr.ComputeRequestArgs) (*AllocationData, error) {
	var err error
	computeTypes := computeRequest.ComputeSizes
	if len(computeTypes) == 0 {
		computeTypes, err = data.NewComputeSelector().Select(computeRequest)
		if err != nil {
			return nil, err
		}
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}

	ad := &AllocationData{}
	ad.AZ, err = data.GetRandomAvailabilityZone(region, nil)
	if err != nil {
		return nil, err
	}

	availableInstanceTypes, err := data.FilterInstaceTypesOfferedByRegion(computeTypes, region)
	if err != nil {
		return nil, err
	}

	// Ensure we have at least one instance type
	if len(availableInstanceTypes) == 0 {
		return nil, fmt.Errorf("no instance types available in region %s for the specified compute requirements", region)
	}

	// Note: For on-demand, we don't need to handle AMI name like spot does
	// since we're not searching for spot pricing across regions

	return &AllocationData{
		Region:        &region,
		AZ:            ad.AZ,
		SpotPrice:     nil, // No spot pricing for on-demand
		InstanceTypes: availableInstanceTypes,
	}, nil
}
