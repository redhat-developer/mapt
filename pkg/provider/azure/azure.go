package azure

import (
	"os"
	"slices"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
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

func (a *Azure) Custom(ctx *pulumi.Context) (*pulumi.ProviderResource, error) {
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

func Destroy(projectName, backedURL, stackName string) error {
	stack := manager.Stack{
		StackName:           stackName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: DefaultCredentials}
	return manager.DestroyStack(stack)
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
