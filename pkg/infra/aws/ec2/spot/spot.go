package spot

import (
	"sort"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/meta/geo"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SpotPriceRequest struct {
	ProductDescription string
	InstanceType       string
	AvailabilityZones  []string
}

type SpotPriceData struct {
	Price            string
	AvailabilityZone string
	Region           string
	InstanceType     string
}

type SpotPriceResult struct {
	Data SpotPriceData
	Err  error
}

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription string = "product-description"

	StackGetSpotPriceName                   string = "Get-SpotPrice"
	StackGetSpotPriceOutputSpotPrice        string = "spotPrice"
	StackGetSpotPriceOutputAvailabilityZone string = "availabilityZone"
)

func (s SpotPriceRequest) GetSpotPrice(ctx *pulumi.Context) error {
	var spotPrices []*ec2.GetSpotPriceResult

	// If empty azs it will check all non opted-in
	availabilityZones := util.If(len(s.AvailabilityZones) != 0,
		s.AvailabilityZones,
		geo.GetAvailabilityZones(ctx))

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
