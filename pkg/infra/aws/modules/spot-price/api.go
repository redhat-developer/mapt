package spotprice

import (
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/regions"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"

	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
)

func Create(projectName, backedURL, targetHostID string) (*SpotPriceGroup, error) {
	stack, err := utilInfra.CheckStack(utilInfra.Stack{
		StackName:   StackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
	})
	if err != nil {
		return createStack(projectName, backedURL, targetHostID)
	} else {
		return getOutputs(stack)
	}
}

func Destroy(projectName, backedURL string) (err error) {
	stack := utilInfra.Stack{
		StackName:   StackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	err = utilInfra.DestroyStack(stack)
	if err == nil {
		logging.Debugf("%s has been destroyed", StackName)
	}
	return
}

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
	sps, err := GetBestPlacementScores(regions, host.InstaceTypes, 1)
	if err != nil {
		logging.Errorf("failed to get spot placement scores")
		// os.Exit(1)
	}
	worldwidePrices := getPricesPerRegion(host.ProductDescription, regions, host.InstaceTypes)
	// check this
	bestPrice := checkBestOption(worldwidePrices, sps, getDescribeAvailabilityZones(regions))
	if bestPrice != nil {
		logging.Debugf("Based on avg prices for instance types %v is az %s, current avg price is %.2f and max price is %.2f with a score of %d",
			host.InstaceTypes, bestPrice.AvailabilityZone, bestPrice.AVGPrice, bestPrice.MaxPrice, bestPrice.Score)
	}
	return bestPrice, nil
}

func createStack(projectName, backedURL, targetHostID string) (*SpotPriceGroup, error) {
	request := SpotPriceRequest{
		TargetHostID: targetHostID,
		Name:         projectName,
	}
	stack := utilInfra.Stack{
		StackName:   StackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
		DeployFunc:  request.deployer,
	}
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return nil, err
	}
	return getSpotPriceGroupFromStackResult(stackResult)
}

func getOutputs(stack *auto.Stack) (*SpotPriceGroup, error) {
	outputs, err := utilInfra.GetOutputs(*stack)
	if err != nil {
		return nil, err
	}
	return getSpotPriceGroupFromStackOutputs(outputs), nil
}
