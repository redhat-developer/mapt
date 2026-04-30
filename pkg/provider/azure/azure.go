package azure

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	"github.com/redhat-developer/mapt/pkg/provider/azure/services/storage/blob"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

var azIdentityEnvs = []string{
	"AZURE_TENANT_ID",
	"AZURE_SUBSCRIPTION_ID",
	"AZURE_CLIENT_ID",
	"AZURE_CLIENT_SECRET",
}

type Azure struct{}

func Provider() *Azure {
	return &Azure{}
}

func (a *Azure) Init(ctx context.Context, backedURL string) error {
	setAZIdentityEnvs()
	return nil
}

func (a *Azure) DefaultHostingPlace() (*string, error) {
	hp := os.Getenv("ARM_LOCATION_NAME")
	if len(hp) > 0 {
		return &hp, nil
	}
	hp = os.Getenv("AZURE_DEFAULTS_LOCATION")
	if len(hp) > 0 {
		return &hp, nil
	}
	logging.Infof("missing default value for Azure Location, if needed it should set using ARM_LOCATION_NAME or AZURE_DEFAULTS_LOCATION")
	return nil, nil
}

// Envs required for auth with go sdk
// https://learn.microsoft.com/es-es/azure/developer/go/azure-sdk-authentication?tabs=bash#service-principal-with-a-secret
// do not match standard envs for pulumi envs for auth with native sdk
// https://www.pulumi.com/registry/packages/azure-native/installation-configuration/#set-configuration-using-environment-variables
func setAZIdentityEnvs() {
	for _, e := range azIdentityEnvs {
		if err := os.Setenv(e,
			os.Getenv(strings.ReplaceAll(e, "AZURE", "ARM"))); err != nil {
			logging.Error(err)
		}
	}
}

func GetClouProviderCredentials(fixedCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: nil,
		FixedCredentials:  fixedCredentials}
}

var (
	DefaultCredentials               = GetClouProviderCredentials(nil)
	ResourceGroupLocation            = "eastus"
	locationsSupportingResourceGroup = []string{
		"eastasia",
		"southeastasia",
		"australiaeast",
		"australiasoutheast",
		"brazilsouth",
		"canadacentral",
		"canadaeast",
		"switzerlandnorth",
		"germanywestcentral",
		"eastus2",
		"eastus",
		"centralus",
		"northcentralus",
		"francecentral",
		"uksouth",
		"ukwest",
		"centralindia",
		"southindia",
		"jioindiawest",
		"italynorth",
		"japaneast",
		"japanwest",
		"koreacentral",
		"koreasouth",
		"mexicocentral",
		"northeurope",
		"norwayeast",
		"polandcentral",
		"qatarcentral",
		"spaincentral",
		"swedencentral",
		"uaenorth",
		"westcentralus",
		"westeurope",
		"westus2",
		"westus",
		"southcentralus",
		"westus3",
		"southafricanorth",
		"australiacentral",
		"australiacentral2",
		"israelcentral",
		"westindia",
		"newzealandnorth",
	}
)

const pulumiLocksPath = ".pulumi/locks"

type DestroyStackRequest struct {
	Stackname string
}

func Destroy(mCtx *mc.Context, stackName string) error {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: DefaultCredentials}
	return manager.DestroyStack(mCtx, stack)
}

func DestroyStack(mCtx *mc.Context, s DestroyStackRequest) error {
	logging.Debug("Running destroy operation")
	if len(s.Stackname) == 0 {
		return fmt.Errorf("stackname is required")
	}
	if mCtx.IsForceDestroy() {
		container, path, parseErr := parseAzblobBackedURL(mCtx)
		if parseErr != nil {
			logging.Error(parseErr)
		} else {
			prefix := pulumiLocksPath + "/"
			if path != "" {
				prefix = path + "/" + prefix
			}
			if err := blob.Delete(mCtx.Context(), container, prefix); err != nil {
				logging.Error(err)
			}
		}
	}
	return manager.DestroyStack(
		mCtx,
		manager.Stack{
			StackName:           mCtx.StackNameByProject(s.Stackname),
			ProjectName:         mCtx.ProjectName(),
			BackedURL:           mCtx.BackedURL(),
			ProviderCredentials: DefaultCredentials})
}

func parseAzblobBackedURL(mCtx *mc.Context) (container string, path string, err error) {
	if !strings.HasPrefix(mCtx.BackedURL(), "azblob://") {
		return "", "", fmt.Errorf("invalid azblob URI: must start with azblob://")
	}
	u, err := url.Parse(mCtx.BackedURL())
	if err != nil {
		return "", "", fmt.Errorf("failed to parse azblob URI: %w", err)
	}
	return u.Host, strings.TrimPrefix(u.Path, "/"), nil
}

func CleanupState(mCtx *mc.Context) error {
	if mCtx.IsKeepState() {
		return nil
	}

	container, path, parseErr := parseAzblobBackedURL(mCtx)
	if parseErr != nil {
		logging.Warnf("Failed to parse azblob backend URL, skipping state cleanup: %v", parseErr)
		return nil
	}

	prefix := ".pulumi/"
	if path != "" {
		prefix = path + "/"
	}
	logging.Infof("Cleaning up Pulumi state from azblob://%s/%s", container, path)
	if deleteErr := blob.Delete(mCtx.Context(), container, prefix); deleteErr != nil {
		logging.Warnf("Failed to cleanup Azure blob state: %v", deleteErr)
	} else {
		logging.Info("Successfully cleaned up Pulumi state from Azure Blob Storage")
	}

	return nil
}

func locationSupportsResourceGroup(location string) bool {
	return slices.Contains(locationsSupportingResourceGroup, location)
}

func GetSuitableLocationForResourceGroup(location string) string {
	if locationSupportsResourceGroup(location) {
		return location
	}

	return locationsSupportingResourceGroup[0]
}
