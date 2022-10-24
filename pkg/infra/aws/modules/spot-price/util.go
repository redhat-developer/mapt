package spotprice

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func exportSpotPriceGroup(spg *SpotPriceGroup, ctx *pulumi.Context) {
	ctx.Export(StackOutputRegion,
		pulumi.String(spg.Region))
	ctx.Export(StackOutputAvailabilityZone,
		pulumi.String(spg.AvailabilityZone))
	ctx.Export(StackOutputMaxPrice,
		pulumi.Float64(spg.MaxPrice))
	ctx.Export(StackOutputAVGPrice,
		pulumi.Float64(spg.AVGPrice))
	ctx.Export(StackOutputScore,
		pulumi.Float64(spg.Score))
}

func getSpotPriceGroupFromStackResult(stackResult auto.UpResult) (spg *SpotPriceGroup, err error) {
	spg = &SpotPriceGroup{}
	region, ok := stackResult.Outputs[StackOutputRegion].Value.(string)
	if !ok {
		err = fmt.Errorf("error getting region")
	}
	spg.Region = region
	az, ok := stackResult.Outputs[StackOutputAvailabilityZone].Value.(string)
	if !ok {
		err = fmt.Errorf("error getting az")
	}
	spg.AvailabilityZone = az
	maxPrice, ok := stackResult.Outputs[StackOutputMaxPrice].Value.(float64)
	if !ok {
		err = fmt.Errorf("error getting max price")
	}
	spg.MaxPrice = maxPrice
	avgPrice, ok := stackResult.Outputs[StackOutputAVGPrice].Value.(float64)
	if !ok {
		err = fmt.Errorf("error getting avg price")
	}
	spg.AVGPrice = avgPrice
	score, ok := stackResult.Outputs[StackOutputScore].Value.(float64)
	if !ok {
		return nil, fmt.Errorf("error getting score")
	}
	spg.Score = int64(score)
	return
}
