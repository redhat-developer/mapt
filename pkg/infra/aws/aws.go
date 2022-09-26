package aws

import (
	"context"
	"os"

	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const BACKEND_URL string = "file:///tmp/qenvs"

const StackGetRegionsName string = "Get-Regions"
const StackGetRegionsOutputAWSRegions string = "AWS_REGIONS"

func GetRegions(projectName, backedURL string) ([]string, error) {
	ctx := context.Background()
	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	getRegionsStack := infraUtil.Stack{
		StackName:   StackGetRegionsName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      PluginAWSDefault,
		DeployFunc:  getRegionsStack,
	}
	// Plan stack
	objectStack := infraUtil.GetStack(ctx, getRegionsStack)
	getRegionsStackResult, err := objectStack.Up(ctx, stdoutStreamer)
	if err != nil {
		return nil, err
	}
	regions, ok := getRegionsStackResult.Outputs[StackGetRegionsOutputAWSRegions].Value.([]interface{})
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
