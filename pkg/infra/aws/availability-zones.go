package aws

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetAvailabilityZones(ctx *pulumi.Context) []string {
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	}, nil)
	if err != nil {
		return nil
	}
	return azs.Names
}

func GetNotOptedInRegions(ctx *pulumi.Context) ([]string, error) {
	regions, err := aws.GetRegions(ctx, &aws.GetRegionsArgs{
		// AllRegions: pulumi.BoolRef(true),
		Filters: []aws.GetRegionsFilter{{
			Name:   "opt-in-status",
			Values: []string{"opt-in-not-required"}}},
	}, nil)
	if err != nil {
		return nil, err
	}
	return regions.Names, nil
}
