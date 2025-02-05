package iam

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

func (a *iamRequestArgs) deploy(ctx *pulumi.Context) error {
	user, err := iam.NewUser(ctx,
		resourcesUtil.GetResourceName(a.prefix, a.componentID, "user"),
		&iam.UserArgs{
			Name: pulumi.String(a.name),
		})
	if err != nil {
		return err
	}
	_, err = iam.NewUserPolicy(ctx,
		resourcesUtil.GetResourceName(a.prefix, a.componentID, "policy"),
		&iam.UserPolicyArgs{
			User:   user.Name,
			Policy: pulumi.String(*a.policyContent),
		})
	if err != nil {
		return err
	}
	accessKey, err := iam.NewAccessKey(
		ctx,
		resourcesUtil.GetResourceName(a.prefix, a.componentID, "ak"),
		&iam.AccessKeyArgs{
			User: user.Name,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", a.prefix, outputAccessKey), accessKey.ID())
	ctx.Export(fmt.Sprintf("%s-%s", a.prefix, outputSecretKey), accessKey.Secret)
	return nil
}
