package iam

import (
	"fmt"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(mCtx *mc.Context, accountName string, policyContent *string) error {
	r := &iamRequestArgs{
		mCtx:          mCtx,
		name:          accountName,
		policyContent: policyContent,
		// Being isolated stack these values
		// do not care
		prefix:      "mapt",
		componentID: "",
	}
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, err := manager.UpStack(mCtx, stack)
	if err != nil {
		return err
	}
	return manageResultsMachine(mCtx, sr, r.prefix)
}

func Destroy(mCtx *mc.Context) (err error) {
	logging.Debug("Destroy iam resources")
	return aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			Stackname: stackName,
		})
}

func manageResultsMachine(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputAccessKey): "accessKey",
		fmt.Sprintf("%s-%s", prefix, outputSecretKey): "secretKey",
	}
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}
