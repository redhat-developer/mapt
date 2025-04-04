package iam

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// Create a instance profile based on a list of policies
func InstanceProfile(ctx *pulumi.Context, prefix, id *string, policiesARNs []string) (*iam.InstanceProfile, error) {
	r, err := iam.NewRole(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "ec2-role"),
		&iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": { "Service": "ec2.amazonaws.com" },
					"Action": "sts:AssumeRole"
				}
			]
		}`),
		})
	if err != nil {
		return nil, err
	}
	for i, p := range policiesARNs {
		_, err = iam.NewRolePolicyAttachment(ctx,
			resourcesUtil.GetResourceName(*prefix, *id, fmt.Sprintf("ec2-role-attach-%d", i)),
			&iam.RolePolicyAttachmentArgs{
				Role:      r.Name,
				PolicyArn: pulumi.String(p),
			})
		if err != nil {
			return nil, err
		}
	}
	return iam.NewInstanceProfile(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "instance-profie"),
		&iam.InstanceProfileArgs{
			Role: r})
}

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
