package stacks

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/ec2/spot"
	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util"
)

func GetBestSpotPriceAsync(stackSuffix, projectName, backedURL,
	instanceType, productDescription, region string, c chan spot.SpotPriceResult) {
	price, availabilityZone, err := GetBestSpotPrice(
		stackSuffix, projectName, backedURL,
		instanceType, productDescription, region)
	c <- spot.SpotPriceResult{
		Data: spot.SpotPriceData{
			Price:            price,
			AvailabilityZone: availabilityZone,
			Region:           region,
			InstanceType:     instanceType},
		Err: err}

}

func GetBestSpotPrice(stackSuffix, projectName, backedURL, instanceType,
	productDescription, region string) (string, string, error) {

	request := spot.SpotPriceRequest{
		InstanceType:       instanceType,
		ProductDescription: productDescription}
	stackName := util.If(
		len(stackSuffix) > 0,
		fmt.Sprintf("%s-%s", spot.StackGetSpotPriceName, stackSuffix),
		spot.StackGetSpotPriceName)
	stack := infraUtil.Stack{
		StackName:   stackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.GetPluginAWS(map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc:  request.GetSpotPrice,
	}
	// Exec stack
	stackResult, err := infraUtil.ExecStack(stack)
	if err != nil {
		return "", "", err
	}
	bestPrice, ok := stackResult.Outputs[spot.StackGetSpotPriceOutputSpotPrice].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	bestPriceAZ, ok := stackResult.Outputs[spot.StackGetSpotPriceOutputAvailabilityZone].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	return bestPrice, bestPriceAZ, nil
}
