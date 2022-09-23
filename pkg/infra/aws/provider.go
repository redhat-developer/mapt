package aws

import (
	"context"
	"os"

	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// pulumi config key : aws env credential
var credentials = map[string]string{
	"aws:region":    "AWS_DEFAULT_REGION",
	"aws:accessKey": "AWS_ACCESS_KEY_ID",
	"aws:secretKey": "AWS_SECRET_ACCESS_KEY"}

func SetAWSCredentials(ctx context.Context, stack auto.Stack) error {
	for configKey, envKey := range credentials {
		if err := stack.SetConfig(ctx, configKey,
			auto.ConfigValue{Value: os.Getenv(envKey)}); err != nil {
			logging.Errorf("Failed setting credential: %v", err)
			return err
		}
	}
	return nil
}

var PluginAWS = infraUtil.PluginInfo{
	Name:              "aws",
	Version:           "v4.0.0",
	SetCredentialFunc: SetAWSCredentials}
