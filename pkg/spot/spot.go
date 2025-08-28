package spot

import (
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	awsData "github.com/redhat-developer/mapt/pkg/provider/aws/data"
	azureData "github.com/redhat-developer/mapt/pkg/provider/azure/data"
)

const (
	ALL Provider = iota
	AWS
	Azure
)

type Provider int

// GetLowestPrice fetches prices of spot instances for all the supported
// providers and returns the results as a:  map[string]SpotPrice
// where map index key is the cloud provider name
func GetLowestPrice(args *spot.SpotRequestArgs, p Provider) (result map[Provider]*spot.SpotResults, err error) {
	mctx := mc.InitNoState()
	result = make(map[Provider]*spot.SpotResults)
	if p == ALL || p == AWS {
		result[AWS], err = awsData.NewSpotSelector().Select(mctx, args)
		if err != nil {
			return nil, err
		}
	}
	if p == ALL || p == Azure {
		result[Azure], err = azureData.NewSpotSelector().Select(mctx, args)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
