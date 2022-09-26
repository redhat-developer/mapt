package ec2

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SpotPriceStackRequest struct {
	ProductDescription string
	InstanceType       string
	AvailabilityZones  []string
}

type SpotPriceData struct {
	Price            string
	AvailabilityZone string
	Region           string
	Err              error
}

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription string = "product-description"

	stackGetSpotPriceName                   string = "Get-SpotPrice"
	stackGetSpotPriceOutputSpotPrice        string = "spotPrice"
	stackGetSpotPriceOutputAvailabilityZone string = "availabilityZone"
)

func GetBestSpotPriceAsync(stackSuffix, projectName, backedURL, instanceType, productDescription, region string, c chan SpotPriceData) {
	price, availabilityZone, err := GetBestSpotPrice(stackSuffix, projectName, backedURL, instanceType, productDescription, region)
	c <- SpotPriceData{
		Price:            price,
		AvailabilityZone: availabilityZone,
		Region:           region,
		Err:              err}
}

func GetBestSpotPrice(stackSuffix, projectName, backedURL, instanceType, productDescription, region string) (string, string, error) {
	ctx := context.Background()
	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	stackRequest := SpotPriceStackRequest{
		InstanceType:       instanceType,
		ProductDescription: productDescription}
	stackName := util.If(
		len(stackSuffix) > 0,
		fmt.Sprintf("%s-%s", stackGetSpotPriceName, stackSuffix),
		stackGetSpotPriceName)
	stack := infraUtil.Stack{
		StackName:   stackName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.GetPluginAWS(map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc:  stackRequest.getSpotPrice,
	}
	// Plan stack
	objectStack := infraUtil.GetStack(ctx, stack)
	stackResult, err := objectStack.Up(ctx, stdoutStreamer)
	if err != nil {
		return "", "", err
	}
	bestPrice, ok := stackResult.Outputs[stackGetSpotPriceOutputSpotPrice].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	bestPriceAZ, ok := stackResult.Outputs[stackGetSpotPriceOutputAvailabilityZone].Value.(string)
	if !ok {
		return "", "", fmt.Errorf("error getting best price for spot")
	}
	return bestPrice, bestPriceAZ, nil
}

func (s SpotPriceStackRequest) getSpotPrice(ctx *pulumi.Context) error {
	var spotPrices []*ec2.GetSpotPriceResult

	// If empty azs it will check all non opted-in
	availabilityZones := util.If(len(s.AvailabilityZones) != 0,
		s.AvailabilityZones,
		aws.GetAvailabilityZones(ctx))

	for _, availabilityZone := range availabilityZones {
		if spotPrice, err := s.getSpotPriceByAZ(availabilityZone, ctx); err != nil {
			logging.Debugf("Can not get price for %s on %s due to %v", s.InstanceType, availabilityZone, err)
		} else {
			spotPrices = append(spotPrices, spotPrice)
		}
	}
	minSpotPrice := minSpotPrice(spotPrices)
	if minSpotPrice != nil {
		ctx.Export(stackGetSpotPriceOutputSpotPrice, pulumi.String(minSpotPrice.SpotPrice))
		ctx.Export(stackGetSpotPriceOutputAvailabilityZone, pulumi.String(*minSpotPrice.AvailabilityZone))
	}
	// export the website URL
	return nil
}

func (s SpotPriceStackRequest) getSpotPriceByAZ(az string, ctx *pulumi.Context) (*ec2.GetSpotPriceResult, error) {
	return ec2.GetSpotPrice(ctx, &ec2.GetSpotPriceArgs{
		AvailabilityZone: pulumi.StringRef(az),
		Filters: []ec2.GetSpotPriceFilter{
			{
				Name: spotQueryFilterProductDescription,
				Values: []string{
					s.ProductDescription,
				},
			},
		},
		InstanceType: pulumi.StringRef(s.InstanceType),
	}, nil)
}

func minSpotPrice(source []*ec2.GetSpotPriceResult) *ec2.GetSpotPriceResult {
	if len(source) == 0 {
		return nil
	}
	sort.Slice(source, func(i, j int) bool {
		iPrice, _ := strconv.ParseFloat(source[i].SpotPrice, 64)
		jPrice, _ := strconv.ParseFloat(source[j].SpotPrice, 64)
		return iPrice < jPrice
	})
	return source[0]
}

func MinSpotPricePerRegions(source []SpotPriceData) *SpotPriceData {
	if len(source) == 0 {
		return nil
	}
	sort.Slice(source, func(i, j int) bool {
		iPrice, _ := strconv.ParseFloat(source[i].Price, 64)
		jPrice, _ := strconv.ParseFloat(source[j].Price, 64)
		return iPrice < jPrice
	})
	return &source[0]
}
