package serverless

import (
	"fmt"

	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// function to create a mapt servless cmd which will be executed repeatedly
// interval should match expected expression
// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html
func Create(args *ServerlessArgs) error {
	// Initially manage it by setup, may we need to customize the region
	//
	// THis was initially created for mac, if no FixedLocation we may
	// on a situation where region differs from resources managed..is this working??
	r := request(args)
	stack := manager.Stack{
		StackName:           maptContext.StackNameByProject(maptServerlessDefaultPrefix),
		ProjectName:         fmt.Sprintf("%s-%s", r.prefix, maptContext.ProjectName()),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	_, err := manager.UpStack(stack)
	return err
}

func Destroy(prefix *string) (err error) {
	logging.Debug("Destroy serverless resources")
	prefixValue := util.If(prefix != nil,
		*prefix,
		defaultPrefix)
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			ProjectName: fmt.Sprintf("%s-%s",
				prefixValue,
				maptContext.ProjectName()),
			Stackname: maptServerlessDefaultPrefix,
		})
}

func Deploy(ctx *pulumi.Context, args *ServerlessArgs) (*awsxecs.FargateTaskDefinition, error) {
	return request(args).resources(ctx)
}

func OneTimeDelayedTask(ctx *pulumi.Context,
	containerName, region, prefix, componentID string,
	cmd string,
	delay string) error {
	if err := checkBackedURLForServerless(); err != nil {
		return err
	}
	se, err := generateOneTimeScheduleExpression(region, delay)
	if err != nil {
		return err
	}
	r := &serverlessRequestArgs{
		containerName:      containerName,
		region:             region,
		command:            cmd,
		scheduleType:       &OneTime,
		scheduleExpression: se,
		prefix:             prefix,
		componentID:        componentID,
	}

	return r.deploy(ctx)
}
