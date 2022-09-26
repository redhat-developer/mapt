package orchestrator

import (
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/ec2"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

// const BACKEND_URL string = "s3://devopsbox-test-pulumi-backend"
const BACKED_URL string = "file:///tmp/qenvs"
const PROJECT_NAME string = "qenvs"

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
	worldwidePrices := getBestPricesPerRegion(PROJECT_NAME, BACKED_URL, instanceType, productDescription, regions)
	bestPrice := ec2.MinSpotPricePerRegions(worldwidePrices)
	if bestPrice != nil {
		logging.Debugf("Best price for instance %s is %s on %s", instanceType, bestPrice.Price, bestPrice.AvailabilityZone)
	}
	return nil
}

func getBestPricesPerRegion(projectName, backedURL, instanceType, productDescription string, regions []string) []ec2.SpotPriceData {
	worldwidePrices := []ec2.SpotPriceData{}
	c := make(chan ec2.SpotPriceData)
	for _, region := range regions {
		go ec2.GetBestSpotPriceAsync(region, PROJECT_NAME, BACKED_URL, instanceType, productDescription, region, c)
	}
	for i := 0; i < len(regions); i++ {
		infoPrice := <-c
		if infoPrice.Err == nil {
			worldwidePrices = append(worldwidePrices, ec2.SpotPriceData{
				Price:            infoPrice.Price,
				AvailabilityZone: infoPrice.AvailabilityZone,
				Region:           infoPrice.Region})
		}
	}
	close(c)
	return worldwidePrices
}
