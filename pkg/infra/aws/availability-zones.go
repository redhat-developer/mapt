package aws

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetNotOptedInAvailabilityZones(ctx *pulumi.Context) []string {
	// azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
	// 	State: pulumi.StringRef("available"),
	// 	Filters: []aws.GetAvailabilityZonesFilter{{
	// 		Name:   "opt-in-status",
	// 		Values: []string{"not-opted-in"}}},
	// }, nil)
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	}, nil)
	if err != nil {
		return nil
	}
	return azs.Names
}
