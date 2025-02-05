package iam

import (
	"fmt"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(accountName string, policyContent *string) error {
	r := &iamRequestArgs{
		name:          accountName,
		policyContent: policyContent,
		// Being isolated stack these values
		// do not care
		prefix:      "mapt",
		componentID: "",
	}
	stack := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackName),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, err := manager.UpStack(stack)
	if err != nil {
		return err
	}
	return manageResultsMachine(sr, r.prefix)
}

func Destroy() (err error) {
	logging.Debug("Destroy iam resources")
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		})
}

func manageResultsMachine(stackResult auto.UpResult, prefix string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputAccessKey): "accessKey",
		fmt.Sprintf("%s-%s", prefix, outputSecretKey): "secretKey",
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}
