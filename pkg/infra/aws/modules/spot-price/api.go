package spotprice

import (
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/regions"
	"github.com/adrianriobo/qenvs/pkg/util"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

func GetBestBidForSpot(projectName, backedURL string, azs, instanceTypes []string, productDescription string) error {
	regions, err := regions.GetRegions(projectName, backedURL)
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
		projectName, backedURL, productDescription, regions, instanceTypes)
	bestPrice := minSpotPricePerRegions(worldwidePrices)
	if bestPrice != nil {
		logging.Debugf("Best price found !!! instance type is %s on %s, current price is %s",
			bestPrice.InstanceType, bestPrice.AvailabilityZone, bestPrice.Price)
	}
	return nil
}

func getBestPricesPerRegion(projectName, backedURL, productDescription string,
	regions, instanceTypes []string) []SpotPriceData {
	worldwidePrices := []SpotPriceData{}
	c := make(chan SpotPriceResult)
	for _, region := range regions {
		for _, instanceType := range instanceTypes {
			go GetBestSpotPriceAsync(
				fmt.Sprintf("%s-%s", region, instanceType),
				projectName,
				backedURL,
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

func GetBestSpotPrice(stackSuffix, projectName, backedURL, instanceType,
	productDescription, region string) (string, string, error) {

	request := SpotPriceRequest{
		InstanceType:       instanceType,
		ProductDescription: productDescription}
	stackName := util.If(
		len(stackSuffix) > 0,
		fmt.Sprintf("%s-%s", StackGetSpotPriceName, stackSuffix),
		StackGetSpotPriceName)
	stack := utilInfra.Stack{
		StackName:   stackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.GetPluginAWS(map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc:  request.GetSpotPrice,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return "", "", err
	}
	bestPrice, ok := stackResult.Outputs[StackGetSpotPriceOutputSpotPrice].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	bestPriceAZ, ok := stackResult.Outputs[StackGetSpotPriceOutputAvailabilityZone].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	return bestPrice, bestPriceAZ, nil
}

func GetBestSpotPriceAsync(stackSuffix, projectName, backedURL,
	instanceType, productDescription, region string, c chan SpotPriceResult) {
	price, availabilityZone, err := GetBestSpotPrice(
		stackSuffix, projectName, backedURL,
		instanceType, productDescription, region)
	c <- SpotPriceResult{
		Data: SpotPriceData{
			Price:            price,
			AvailabilityZone: availabilityZone,
			Region:           region,
			InstanceType:     instanceType},
		Err: err}

}
