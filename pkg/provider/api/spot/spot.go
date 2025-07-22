package spot

import (
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	awsData "github.com/redhat-developer/mapt/pkg/provider/aws/data"
	azureData "github.com/redhat-developer/mapt/pkg/provider/azure/data"
	gcpData "github.com/redhat-developer/mapt/pkg/provider/gcp/data"
)

const (
	ALL Provider = iota
	AWS
	Azure
	GCP
)

type Provider int

// GetLowestPrice fetches prices of spot instances for all the supported
// providers and returns the results as a:  map[string]SpotPrice
// where map index key is the cloud provider name
func GetLowestPrice(args *spotTypes.SpotRequestArgs, p Provider) (result map[Provider]*spotTypes.SpotResults, err error) {
	result = make(map[Provider]*spotTypes.SpotResults)
	if p == ALL || p == AWS {
		result[AWS], err = awsData.NewSpotSelector().Select(args)
		if err != nil {
			return nil, err
		}
	}
	if p == ALL || p == Azure {
		result[Azure], err = azureData.NewSpotSelector().Select(args)
		if err != nil {
			return nil, err
		}
	}
	if p == ALL || p == GCP {
		result[GCP], err = gcpData.NewSpotSelector().Select(args)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
