package azure

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
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

func (a *Azure) Init(backedURL string) error {
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
	return nil, fmt.Errorf("missing default value for Azure Location")
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

func Destroy(mCtx *mc.Context, stackName string) error {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: DefaultCredentials}
	return manager.DestroyStack(mCtx, stack)
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
