package serverless

import (
	"fmt"
	"os"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
)

// function to create a mapt servless cmd which will be executed repeatedly
// interval should match expected expression
// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html
func CreateRepeatedlyAsStack(command, rateExpression string) error {
	// Initially manage it by setup, may we need to customize the region
	//
	// THis was initially created for mac, if no FixedLocation we may
	// on a situation where region differs from resources managed..is this working??
	region := os.Getenv("AWS_DEFAULT_REGION")
	r := &serverlessRequestArgs{
		region:             region,
		command:            command,
		scheduleExpression: fmt.Sprintf("rate(%s)", rateExpression),
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
