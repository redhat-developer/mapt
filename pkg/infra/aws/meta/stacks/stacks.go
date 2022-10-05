package stacks

import (
	awsCommon "github.com/adrianriobo/qenvs/pkg/infra/aws"
	meta "github.com/adrianriobo/qenvs/pkg/infra/aws/meta"
	util "github.com/adrianriobo/qenvs/pkg/util"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	StackGetRegionsName string = "Get-Regions"

	StackGetRegionsOutputAWSRegions string = "AWS_REGIONS"
)

func GetRegions(projectName, backedURL string) ([]string, error) {
	stack := utilInfra.Stack{
		StackName:   StackGetRegionsName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      awsCommon.PluginAWSDefault,
		DeployFunc:  getRegionsStack,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
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
	regions, err := meta.GetNotOptedInRegions(ctx)
	if err == nil {
		ctx.Export(StackGetRegionsOutputAWSRegions, pulumi.ToStringArray(regions))
	}
	return
}
