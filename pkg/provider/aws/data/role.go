package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func GetRole(ctx context.Context, roleName string) (*string, error) {
	cfg, err := getGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := iam.NewFromConfig(cfg)
	roleOutput, err := client.GetRole(
		ctx, &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		})
	if err != nil {
		return nil, err
	}
	return roleOutput.Role.Arn, nil
}
