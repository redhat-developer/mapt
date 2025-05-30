package iam

import (
	"fmt"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const defaultPrefix = "mapt"

func Create(prefix, accountName string, policyContent *string) error {
	r := &iamRequestArgs{
		name:          accountName,
		policyContent: policyContent,
		prefix: util.If(len(prefix) > 0,
			prefix,
			string(defaultPrefix)),
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
	return Results(sr, r.prefix)
}

func Destroy() (err error) {
	logging.Debug("Destroy iam resources")
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		})
}

func Deploy(ctx *pulumi.Context, prefix, accountName string, policyContent *string, dependsOn []pulumi.Resource) (*iam.User, *iam.AccessKey, error) {
	r := &iamRequestArgs{
		name:          accountName,
		policyContent: policyContent,
		prefix: util.If(len(prefix) > 0,
			prefix,
			string(defaultPrefix)),
		componentID: "",
		dependsOn:   dependsOn,
	}
	return r.resources(ctx)
}

func Results(stackResult auto.UpResult, prefix string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputAccessKey): "accessKey",
		fmt.Sprintf("%s-%s", prefix, outputSecretKey): "secretKey",
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}
