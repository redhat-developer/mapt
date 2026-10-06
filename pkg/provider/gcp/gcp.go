package gcp

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	gcpConstants "github.com/redhat-developer/mapt/pkg/provider/gcp/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type GCP struct{}

func Provider() *GCP {
	return &GCP{}
}

func (g *GCP) Init(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (g *GCP) DefaultHostingPlace() (*string, error) {
	hp := os.Getenv("GOOGLE_REGION")
	if len(hp) > 0 {
		return &hp, nil
	}
	hp = os.Getenv("GCP_REGION")
	if len(hp) > 0 {
		return &hp, nil
	}
	hp = os.Getenv("CLOUDSDK_COMPUTE_REGION")
	if len(hp) > 0 {
		return &hp, nil
	}
	return nil, fmt.Errorf("missing default value for GCP Region, set GOOGLE_REGION environment variable")
}

var envCredentials = map[string]string{
	gcpConstants.CONFIG_GCP_PROJECT: "GOOGLE_PROJECT",
	gcpConstants.CONFIG_GCP_REGION:  "GOOGLE_REGION",
}

var DefaultCredentials = GetClouProviderCredentials(nil)

func GetClouProviderCredentials(customCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: SetGCPCredentials,
		FixedCredentials:  customCredentials,
	}
}

func SetGCPCredentials(ctx context.Context, mCtx *mc.Context, stack auto.Stack, customCredentials map[string]string) error {
	for configKey, envKey := range envCredentials {
		if value, ok := customCredentials[configKey]; ok {
			if err := stack.SetConfig(ctx, configKey,
				auto.ConfigValue{Value: value}); err != nil {
				logging.Errorf("Failed setting credential: %v", err)
				return err
			}
		} else {
			if err := stack.SetConfig(ctx, configKey,
				auto.ConfigValue{Value: os.Getenv(envKey)}); err != nil {
				logging.Errorf("Failed setting credential: %v", err)
				return err
			}
		}
	}
	return nil
}

func Destroy(mCtx *mc.Context, stackName string) error {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: DefaultCredentials,
	}
	return manager.DestroyStack(mCtx, stack)
}
