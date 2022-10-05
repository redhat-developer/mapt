package orchestrator

import (
	"fmt"
	"os"

	ec2Spot "github.com/adrianriobo/qenvs/pkg/infra/aws/ec2/spot"
	awsEC2Stacks "github.com/adrianriobo/qenvs/pkg/infra/aws/ec2/stacks"
	awsMetaStacks "github.com/adrianriobo/qenvs/pkg/infra/aws/meta/stacks"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

// const BACKEND_URL string = "s3://devopsbox-test-pulumi-backend"
const BACKED_URL string = "file:///tmp/qenvs"
const PROJECT_NAME string = "qenvs"

func GetBestBidForSpot(azs, instanceTypes []string, productDescription string) error {
	regions, err := awsMetaStacks.GetRegions(PROJECT_NAME, BACKED_URL)
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	logging.Debugf("Got all regions %v", regions)
	// validations
	if len(instanceTypes) == 0 {
		return fmt.Errorf("instance type is required")
	}
	worldwidePrices := getBestPricesPerRegion(
		PROJECT_NAME, BACKED_URL, productDescription, regions, instanceTypes)
	bestPrice := ec2Spot.MinSpotPricePerRegions(worldwidePrices)
	if bestPrice != nil {
		logging.Debugf("Best price found !!! instance type is %s on %s, current price is %s",
			bestPrice.InstanceType, bestPrice.AvailabilityZone, bestPrice.Price)
	}
	return nil
}

func getBestPricesPerRegion(projectName, backedURL, productDescription string,
	regions, instanceTypes []string) []ec2Spot.SpotPriceData {
	worldwidePrices := []ec2Spot.SpotPriceData{}
	c := make(chan ec2Spot.SpotPriceResult)
	for _, region := range regions {
		for _, instanceType := range instanceTypes {
			go awsEC2Stacks.GetBestSpotPriceAsync(
				fmt.Sprintf("%s-%s", region, instanceType),
				PROJECT_NAME,
				BACKED_URL,
				instanceType,
				productDescription,
				region,
				c)
		}
	}
	for i := 0; i < len(regions)*len(instanceTypes); i++ {
		spotPriceResult := <-c
		if spotPriceResult.Err == nil {
			worldwidePrices = append(worldwidePrices, ec2Spot.SpotPriceData{
				Price:            spotPriceResult.Data.Price,
				AvailabilityZone: spotPriceResult.Data.AvailabilityZone,
				Region:           spotPriceResult.Data.Region,
				InstanceType:     spotPriceResult.Data.InstanceType})
		}
	}
	close(c)
	return worldwidePrices
}
