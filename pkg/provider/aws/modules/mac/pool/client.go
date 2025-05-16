package pool

import (
	"encoding/json"
	"fmt"

	awsIAM "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/iam"
)

// Create an user and a pair of automation credentials to add on cicd system of choice
// to execute request and release operation with minimum rights
func clientAccount(ctx *pulumi.Context, name, arch, osVersion string, dependsOn []pulumi.Resource) (*awsIAM.User, *awsIAM.AccessKey, error) {
	pc, err := clientPolicy()
	if err != nil {
		return nil, nil, err
	}
	return iam.Deploy(ctx,
		name,
		fmt.Sprintf("%s-%s-%s", name, arch, osVersion),
		pc, dependsOn)
}

// This is only used during create to create a policy content allowing to
// run request and release operations. Helping to reduce the iam rights required
// to make use for the mac pool service from an user point of view
func clientPolicy() (*string, error) {
	pc, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"ecs:RunTask",
					"ecs:DescribeTasks",
					"ecs:DescribeTaskDefinition",
					"ecs:ListTaskDefinitions",
					"tag:GetResources",
				},
				"Resource": []string{
					"*",
				},
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"iam:PassRole",
				},
				"Resource": []string{
					"*",
				},
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"ec2:DescribeSubnets",
					"ec2:DescribeSecurityGroups",
					"ec2:DescribeVpcs",
					"ec2:DescribeRouteTables",
				},
				"Resource": []string{
					"*",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	policy := string(pc)
	return &policy, nil
}
