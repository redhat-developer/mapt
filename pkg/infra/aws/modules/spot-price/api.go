package spotprice

import (
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/azs"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/regions"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/exp/slices"
)

func BestSpotPriceInfo(targetHostID string) (*SpotPriceGroup, error) {
	regions, err := regions.GetRegions()
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	logging.Debugf("Got all regions %v", regions)
	host, err := supportmatrix.GetHost(targetHostID)
	if err != nil {
		logging.Error(err)
		os.Exit(1)
	}
	// scores (capacity will be calculated by env analyzer)
	sps, err := GetBestPlacementScores(regions, host.Requirements, 1)
	if err != nil {
		logging.Errorf("failed to get spot placement scores")
		// os.Exit(1)
	}
	worldwidePrices := GetPricesPerRegion(host.ProductDescription, regions, host.InstaceTypes)
	// check this
	bestPrice := checkBestOption(worldwidePrices, sps, GetDescribeAvailabilityZones(regions))
	if bestPrice != nil {
		logging.Debugf("Based on avg prices for instance types %v is az %s, current avg price is %.2f and max price is %.2f with a score of %d",
			host.InstaceTypes, bestPrice.AvailabilityZone, bestPrice.AVGPrice, bestPrice.MaxPrice, bestPrice.Score)
	}
	return bestPrice, nil
}

func GetPricesPerRegion(productDescription string,
	regions, instanceTypes []string) []SpotPriceGroup {
	worldwidePrices := []SpotPriceGroup{}
	c := make(chan SpotPriceResult)
	for _, region := range regions {
		go GetBestSpotPriceAsync(
			instanceTypes,
			productDescription,
			region,
			c)
	}
	for i := 0; i < len(regions); i++ {
		spotPriceResult := <-c
		if spotPriceResult.Err == nil {
			worldwidePrices = append(worldwidePrices, spotPriceResult.Prices...)
		}
	}
	close(c)
	return worldwidePrices
}

func GetDescribeAvailabilityZones(regions []string) []*ec2.AvailabilityZone {
	allAvailabilityZones := []*ec2.AvailabilityZone{}
	c := make(chan azs.AvailabilityZonesResult)
	for _, region := range regions {
		go azs.DescribeAvailabilityZonesAsync(region, c)
	}
	for i := 0; i < len(regions); i++ {
		availabilityZonesResult := <-c
		if availabilityZonesResult.Err == nil {
			allAvailabilityZones = append(allAvailabilityZones, availabilityZonesResult.AvailabilityZones...)
		}
	}
	close(c)
	return allAvailabilityZones
}

func checkBestOption(source []SpotPriceGroup, sps []*ec2.SpotPlacementScore, availabilityZones []*ec2.AvailabilityZone) *SpotPriceGroup {
	slices.SortFunc(source,
		func(a, b SpotPriceGroup) bool {
			return a.AVGPrice < b.AVGPrice
		})
	var score int64 = spsMaxScore
	for score > 3 {
		for _, price := range source {
			idx := slices.IndexFunc(sps, func(item *ec2.SpotPlacementScore) bool {
				// Need transform
				spsZoneName, err := azs.GetZoneName(*item.AvailabilityZoneId, availabilityZones)
				if err != nil {
					return false
				}
				return spsZoneName == price.AvailabilityZone &&
					*item.Score == score
			})
			if idx != -1 {
				price.Region = *sps[idx].Region
				price.Score = *sps[idx].Score
				return &price
			}
		}
		score--
	}
	return nil
}

// func minSpotPricePerRegions(source []SpotPriceGroup) *SpotPriceGroup {
// 	if len(source) == 0 {
// 		return nil
// 	}
// 	slices.SortFunc(source,
// 		func(a, b SpotPriceGroup) bool {
// 			return a.AVGPrice < b.AVGPrice
// 		})
// 	return &source[0]
// }
