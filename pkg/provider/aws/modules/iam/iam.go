package iam

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
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

func (r *iamRequestArgs) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	user, err := iam.NewUser(ctx,
		resourcesUtil.GetResourceName(r.prefix, r.componentID, "user"),
		&iam.UserArgs{
			Name: pulumi.String(r.name),
		})
	if err != nil {
		return err
	}
	_, err = iam.NewUserPolicy(ctx,
		resourcesUtil.GetResourceName(r.prefix, r.componentID, "policy"),
		&iam.UserPolicyArgs{
			User:   user.Name,
			Policy: pulumi.String(*r.policyContent),
		})
	if err != nil {
		return err
	}
	accessKey, err := iam.NewAccessKey(
		ctx,
		resourcesUtil.GetResourceName(r.prefix, r.componentID, "ak"),
		&iam.AccessKeyArgs{
			User: user.Name,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputAccessKey), accessKey.ID())
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputSecretKey), accessKey.Secret)
	return nil
}
