package regions

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type regionsRequest struct {
	optINStatus string
}

func getDefaultRegionsRequest() regionsRequest {
	return regionsRequest{
		optINStatus: "opt-in-not-required"}
}

func (r regionsRequest) GetRegions(ctx *pulumi.Context) (err error) {
	regions, err := aws.GetRegions(ctx, &aws.GetRegionsArgs{
		// AllRegions: pulumi.BoolRef(true),
		Filters: []aws.GetRegionsFilter{{
			Name:   "opt-in-status",
			Values: []string{r.optINStatus}}},
	}, nil)
	if err != nil {
		return err
	}
	if err == nil {
		ctx.Export(StackGetRegionsOutputAWSRegions, pulumi.ToStringArray(regions.Names))
	}
	return
}
