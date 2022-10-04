package meta

import (
	awsCommon "github.com/adrianriobo/qenvs/pkg/infra/aws"
	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const StackGetRegionsName string = "Get-Regions"
const StackGetRegionsOutputAWSRegions string = "AWS_REGIONS"

func GetRegions(projectName, backedURL string) ([]string, error) {
	stack := infraUtil.Stack{
		StackName:   StackGetRegionsName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      awsCommon.PluginAWSDefault,
		DeployFunc:  getRegionsStack,
	}
	// Exec stack
	stackResult, err := infraUtil.ExecStack(stack)
	if err != nil {
		return nil, err
	}
	regions, ok := stackResult.Outputs[StackGetRegionsOutputAWSRegions].Value.([]interface{})
	if !ok {
		return nil, err
	}
	return util.ArrayConvert[string](regions), nil
}

func getRegionsStack(ctx *pulumi.Context) (err error) {
	regions, err := GetNotOptedInRegions(ctx)
	if err == nil {
		ctx.Export(StackGetRegionsOutputAWSRegions, pulumi.ToStringArray(regions))
	}
	return
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
