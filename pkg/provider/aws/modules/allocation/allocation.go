package allocation

import (
	"os"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	"github.com/redhat-developer/mapt/pkg/util"
)

type AllocationData struct {
	// location and price (if Spot is enable)
	Region        *string
	AZ            *string
	SpotPrice     *float64
	InstanceTypes []string
}

func AllocationDataOnSpot(prefix, amiProductDescription, amiName *string, instanceTypes []string) (*AllocationData, error) {
	sr := spot.SpotOptionRequest{
		// do not need to filter the AMI as if it does not exist on the target region
		// mapt will copy it
		Prefix:             *prefix,
		ProductDescription: *amiProductDescription,
		InstaceTypes:       instanceTypes,
	}
	if amiName != nil {
		sr.AMIName = *amiName
	}
	so, err := sr.Create()
	if err != nil {
		return nil, err
	}
	availableInstaceTypes, err :=
		data.FilterInstaceTypesOfferedByRegion(instanceTypes, so.Region)
	if err != nil {
		return nil, err
	}
	spSafe := spotPriceBid(so.MaxPrice)
	return &AllocationData{
		Region:        &so.Region,
		AZ:            &so.AvailabilityZone,
		SpotPrice:     &spSafe,
		InstanceTypes: availableInstaceTypes,
	}, nil
}

func AllocationDataOnDemand() (ad *AllocationData, err error) {
	ad = &AllocationData{}
	region := os.Getenv("AWS_DEFAULT_REGION")
	ad.Region = &region
	ad.AZ, err = data.GetRandomAvailabilityZone(region, nil)
	return
}

// Calculate a bid price for spot using a increased rate set by user
func spotPriceBid(basePrice float64) float64 {
	return util.If(maptContext.SpotPriceIncreaseRate() > 0,
		basePrice*(1+float64(maptContext.SpotPriceIncreaseRate())/100),
		basePrice)
}
