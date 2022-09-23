package manager

import (
	"context"
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
)

func GetBestBidForSpot(azs []string, instanceType, productDescription string) error {
	// validations
	if len(instanceType) == 0 {
		return fmt.Errorf("instance type is required")
	}
	// Set information
	request := aws.StackRequest{
		InstanceType:       instanceType,
		ProductDescription: productDescription,
		AvailabilityZones:  azs}
	target := infraUtil.DeployableTarget{
		StackName:   "test",
		ProjectName: "dummy",
		// "s3://devopsbox-test-pulumi-backend",
		BackedURL:  "file:///tmp/qenvs",
		Plugin:     aws.PluginAWS,
		DeployFunc: request.BidSpotPrice,
	}
	ctx := context.Background()
	objectStack := infraUtil.GetStack(ctx, target)

	// wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	spotRes, err := objectStack.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update stack: %v\n\n", err)
	}

	// get the bucketID output that object stack depends on
	bestPrice, ok := spotRes.Outputs[aws.StackSpotOutputSpotPrice].Value.(string)
	if !ok {
		fmt.Println("failed to get spotPrice output")
		os.Exit(1)
	}
	// get the bucketID output that object stack depends on
	bestPriceAZ, ok := spotRes.Outputs[aws.StackSpotOutputAvailabilityZone].Value.(string)
	if !ok {
		fmt.Println("failed to get availability Zone output")
		os.Exit(1)
	}
	fmt.Printf("Best bid price is %s, on az %s\n", bestPrice, bestPriceAZ)
	return nil
}
