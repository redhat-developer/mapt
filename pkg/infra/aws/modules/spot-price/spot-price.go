package spotprice

import (
	"sort"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/meta/azs"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

func (s SpotPriceRequest) GetSpotPrice(ctx *pulumi.Context) error {
	var spotPrices []*ec2.GetSpotPriceResult

	// If empty azs it will check all non opted-in
	availabilityZones := util.If(len(s.AvailabilityZones) != 0,
		s.AvailabilityZones,
		azs.GetAvailabilityZones(ctx))

	for _, availabilityZone := range availabilityZones {
		if spotPrice, err := s.getSpotPriceByAZ(availabilityZone, ctx); err != nil {
			logging.Debugf("Can not get price for %s on %s due to %v", s.InstanceType, availabilityZone, err)
		} else {
			logging.Debugf("Found price for %s on %s, current price is %s", s.InstanceType, availabilityZone, spotPrice.SpotPrice)
			spotPrices = append(spotPrices, spotPrice)
		}
	}
	minSpotPrice := minSpotPrice(spotPrices)
	if minSpotPrice != nil {
		ctx.Export(StackGetSpotPriceOutputSpotPrice, pulumi.String(minSpotPrice.SpotPrice))
		ctx.Export(StackGetSpotPriceOutputAvailabilityZone, pulumi.String(*minSpotPrice.AvailabilityZone))
	}
	return nil
}

func (s SpotPriceRequest) getSpotPriceByAZ(az string, ctx *pulumi.Context) (*ec2.GetSpotPriceResult, error) {
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

func minSpotPricePerRegions(source []SpotPriceData) *SpotPriceData {
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

func (s SpotPriceRequest) AWSGetSpotPrice() (*awsEC2.SpotPrice, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")})
	// Create an EC2 service client.
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	history, err := svc.DescribeSpotPriceHistory(&awsEC2.DescribeSpotPriceHistoryInput{
		InstanceTypes: []*string{&s.InstanceType},
	})
	if err != nil {
		return nil, err
	}

	return min(history.SpotPriceHistory), nil
}

func min(prices []*awsEC2.SpotPrice) *awsEC2.SpotPrice {
	if len(prices) == 0 {
		return nil
	}
	sort.Slice(prices, func(i, j int) bool {
		iPrice, _ := strconv.ParseFloat(*prices[i].SpotPrice, 64)
		jPrice, _ := strconv.ParseFloat(*prices[j].SpotPrice, 64)
		return iPrice < jPrice
	})
	return prices[0]
}
