package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func accountId() (*string, error) {
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return identity.Account, nil
}
