package aws

import (
	"sort"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type StackRequest struct {
	ProductDescription string
	InstanceType       string
	AvailabilityZones  []string
}

// https://github.com/pulumi/automation-api-examples/blob/f5444239378c9891250ee367e9c2a6f26149f375/go/multi_stack_orchestration/main.go#L132

func (s StackRequest) BidSpotPrice(ctx *pulumi.Context) error {
	var spotPrices []*ec2.GetSpotPriceResult

	// If empty azs it will check all non opted-in
	availabilityZones := util.If(len(s.AvailabilityZones) != 0,
		s.AvailabilityZones,
		GetNotOptedInAvailabilityZones(ctx))

	for _, availabilityZone := range availabilityZones {
		if spotPrice, err := s.getSpotPrice(availabilityZone, ctx); err != nil {
			logging.Debugf("Can not get price for %s on %s due to %v", s.InstanceType, availabilityZone, err)
		} else {
			spotPrices = append(spotPrices, spotPrice)
		}
	}
	minSpotPrice := minSpotPrice(spotPrices)
	if minSpotPrice != nil {
		ctx.Export(StackSpotOutputSpotPrice, pulumi.String(minSpotPrice.SpotPrice))
		ctx.Export(StackSpotOutputAvailabilityZone, pulumi.String(*minSpotPrice.AvailabilityZone))
	}
	// export the website URL
	return nil
}

func (s StackRequest) getSpotPrice(az string, ctx *pulumi.Context) (*ec2.GetSpotPriceResult, error) {
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
