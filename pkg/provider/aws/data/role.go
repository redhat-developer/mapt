package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func GetRole(roleName string) (*string, error) {
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	client := iam.NewFromConfig(cfg)
	roleOutput, err := client.GetRole(
		context.TODO(), &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		})
	if err != nil {
		return nil, err
	}
	return roleOutput.Role.Arn, nil
}
