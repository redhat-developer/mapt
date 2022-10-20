package spotprice

import (
	"os"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/regions"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"golang.org/x/exp/slices"
)

func BestSpotPriceInfo(azs []string, supportedHostID string) (*SpotPriceData, error) {
	regions, err := regions.GetRegions()
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	logging.Debugf("Got all regions %v", regions)
	host, err := supportmatrix.GetHost(supportedHostID)
	if err != nil {
		logging.Error(err)
		os.Exit(1)
	}
	// scores (capacity will be calculated by env analyzer)
	err = GetBestPlacementScores(regions, host.Requirements, 1)
	if err != nil {
		logging.Errorf("failed to get spot placement scores")
		// os.Exit(1)
	}
	worldwidePrices := getBestPricesPerRegion(host.ProductDescription, regions, host.InstaceTypes)

	bestPrice := minSpotPricePerRegions(worldwidePrices)
	if bestPrice != nil {
		logging.Debugf("Best price found !!! instance type is %s on %s, current price is %s",
			bestPrice.InstanceType, bestPrice.AvailabilityZone, bestPrice.Price)
	}
	return bestPrice, nil
}

func getBestPricesPerRegion(productDescription string,
	regions, instanceTypes []string) []SpotPriceData {
	worldwidePrices := []SpotPriceData{}
	c := make(chan SpotPriceResult)
	for _, region := range regions {
		for _, instanceType := range instanceTypes {
			go GetBestSpotPriceAsync(
				instanceType,
				productDescription,
				region,
				c)
		}
	}
	for i := 0; i < len(regions)*len(instanceTypes); i++ {
		spotPriceResult := <-c
		if spotPriceResult.Err == nil {
			worldwidePrices = append(worldwidePrices, SpotPriceData{
				Price:            spotPriceResult.Data.Price,
				AvailabilityZone: spotPriceResult.Data.AvailabilityZone,
				Region:           spotPriceResult.Data.Region,
				InstanceType:     spotPriceResult.Data.InstanceType})
		}
	}
	close(c)
	return worldwidePrices
}

func minSpotPricePerRegions(source []SpotPriceData) *SpotPriceData {
	if len(source) == 0 {
		return nil
	}
	slices.SortFunc(source,
		func(a, b SpotPriceData) bool {
			aPrice, err := strconv.ParseFloat(a.Price, 64)
			if err != nil {
				return false
			}
			bPrice, err := strconv.ParseFloat(b.Price, 64)
			if err != nil {
				return false
			}
			return aPrice < bPrice
		})
	return &source[0]
}
