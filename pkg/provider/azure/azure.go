package azure

import (
	"slices"

	"github.com/redhat-developer/mapt/pkg/manager"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
)

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
