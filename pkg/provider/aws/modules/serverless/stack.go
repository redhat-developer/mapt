package serverless

import (
	"os"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// function to create a mapt servless cmd which will be executed repeatedly
// interval should match expected expression
// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html
func Create(mCtx *mc.Context, command string, scheduleType scheduleType, schedulexpression, logGroupName string) error {
	// Initially manage it by setup, may we need to customize the region
	//
	// THis was initially created for mac, if no FixedLocation we may
	// on a situation where region differs from resources managed..is this working??
	region := os.Getenv("AWS_DEFAULT_REGION")
	r := &serverlessRequest{
		mCtx:               mCtx,
		region:             region,
		command:            command,
		scheduleType:       scheduleType,
		scheduleExpression: schedulexpression,
		logGroupName:       logGroupName,
		// Being isolated stack these values
		// do not care
		prefix:      "mapt",
		componentID: "sf",
	}
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(maptServerlessDefaultPrefix),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	_, err := manager.UpStack(mCtx, stack)
	return err
}

func Destroy(mCtx *mc.Context) (err error) {
	logging.Debug("Destroy serverless resources")
	return aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			Stackname: maptServerlessDefaultPrefix,
		})
}
