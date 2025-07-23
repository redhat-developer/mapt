package allocation

import (
	"os"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
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
	spSafe := spotPriceBid(mCtx, so.MaxPrice)
	logging.Debugf("Due to the spot increase rate at %d we will request the spot at %f", mCtx.SpotPriceIncreaseRate(), spSafe)
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
func spotPriceBid(mCtx *mc.Context, basePrice float64) float64 {
	return util.If(mCtx.SpotPriceIncreaseRate() > 0,
		basePrice*(1+float64(mCtx.SpotPriceIncreaseRate())/100),
		basePrice)
}
