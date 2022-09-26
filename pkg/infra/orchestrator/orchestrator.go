package orchestrator

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/ec2"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

// const BACKEND_URL string = "s3://devopsbox-test-pulumi-backend"
const BACKED_URL string = "file:///tmp/qenvs"
const PROJECT_NAME string = "qenvs"

type SpotPriceData struct {
	price            string
	availabilityZone string
	region           string
}

func GetBestBidForSpot(azs []string, instanceType, productDescription string) error {
	regions, err := aws.GetRegions(PROJECT_NAME, BACKED_URL)
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	logging.Debugf("Got all regions %v", regions)
	// validations
	if len(instanceType) == 0 {
		return fmt.Errorf("instance type is required")
	}
	worldwidePrices := []SpotPriceData{}
	for _, region := range regions {

		price, availabilityZone, err := ec2.GetBestSpotPrice(PROJECT_NAME, BACKED_URL, instanceType, productDescription, region)
		if err == nil {
			worldwidePrices = append(worldwidePrices, SpotPriceData{
				price:            price,
				availabilityZone: availabilityZone,
				region:           region})
		}
	}
	bestPrice := minSpotPrice(worldwidePrices)
	if bestPrice != nil {
		logging.Debugf("Best price for instance %s is %s on %s", instanceType, bestPrice.price, bestPrice.availabilityZone)
	}
	return nil
}

func minSpotPrice(source []SpotPriceData) *SpotPriceData {
	if len(source) == 0 {
		return nil
	}
	sort.Slice(source, func(i, j int) bool {
		iPrice, _ := strconv.ParseFloat(source[i].price, 64)
		jPrice, _ := strconv.ParseFloat(source[j].price, 64)
		return iPrice < jPrice
	})
	return &source[0]
}
