package serverless

import (
	"os"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Create(command string, schedulexpression string) error {
	region := os.Getenv("AZURE_DEFAULT_REGION")
	r := &serverlessRequestArgs{
		region:             region,
		command:            command,
		scheduleExpression: schedulexpression,
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
