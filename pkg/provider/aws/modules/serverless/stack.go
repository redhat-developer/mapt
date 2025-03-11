package serverless

import (
	"os"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
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
	region := os.Getenv("AWS_DEFAULT_REGION")
	r := &serverlessRequestArgs{
		region:             region,
		command:            args.Command,
		scheduleType:       args.ScheduleType,
		scheduleExpression: args.Schedulexpression,
		logGroupName:       args.LogGroupName,
		// Being isolated stack these values
		// do not care
		prefix:      "mapt",
		componentID: "sf",
	}
	stack := manager.Stack{
		StackName:           maptContext.StackNameByProject(maptServerlessDefaultPrefix),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	_, err := manager.UpStack(stack)
	return err
}

func Destroy() (err error) {
	logging.Debug("Destroy serverless resources")
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: maptServerlessDefaultPrefix,
		})
}
