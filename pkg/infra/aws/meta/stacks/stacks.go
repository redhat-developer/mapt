package stacks

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/meta/geo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	StackGetRegionsName string = "Get-Regions"

	StackGetRegionsOutputAWSRegions string = "AWS_REGIONS"
)

func GetRegions(ctx *pulumi.Context) (err error) {
	regions, err := geo.GetNotOptedInRegions(ctx)
	if err == nil {
		ctx.Export(StackGetRegionsOutputAWSRegions, pulumi.ToStringArray(regions))
	}
	return
}
