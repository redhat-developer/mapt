package iam

import (
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// Create a instance profile based on a list of policies
func InstanceProfile(ctx *pulumi.Context, prefix, id *string, policiesARNs []string) (*iam.InstanceProfile, error) {
	role, err := iam.NewRole(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "ec2-role"),
		&iam.RoleArgs{
			AssumeRolePolicyDocument: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": { "Service": "ec2.amazonaws.com" },
					"Action": "sts:AssumeRole"
				}
			]
		}`),
			ManagedPolicyArns: pulumi.ToStringArray(policiesARNs),
		})
	if err != nil {
		return nil, err
	}
	// Use the role's RoleName property to reference it in the instance profile
	return iam.NewInstanceProfile(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "instance-profie"),
		&iam.InstanceProfileArgs{
			Roles: pulumi.StringArray{
				role.RoleName.Elem(),
			},
		})
}

func (r *iamRequestArgs) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	_, err := iam.NewUser(ctx,
		resourcesUtil.GetResourceName(r.prefix, r.componentID, "user"),
		&iam.UserArgs{
			UserName: pulumi.String(r.name),
		})
	if err != nil {
		return err
	}
	_, err = iam.NewUserPolicy(ctx,
		resourcesUtil.GetResourceName(r.prefix, r.componentID, "policy"),
		&iam.UserPolicyArgs{
			UserName:       pulumi.String(r.name),
			PolicyDocument: pulumi.String(*r.policyContent),
		})
	if err != nil {
		return err
	}
	// Note: AccessKey is not available in AWS Native provider
	// You would need to use the AWS Classic provider for AccessKey resources
	// or handle key creation through other means (AWS CLI, Console, etc.)
	return nil
}
