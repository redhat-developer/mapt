package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func accountId(ctx context.Context) (*string, error) {
	cfg, err := getGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return identity.Account, nil
}
