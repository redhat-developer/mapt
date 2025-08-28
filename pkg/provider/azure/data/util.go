package data

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
)

const (
	ENV_AZURE_SUBSCRIPTION_ID = "AZURE_SUBSCRIPTION_ID"
)

func getCredentials() (cred *azidentity.DefaultAzureCredential, subscriptionID *string, err error) {
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return
	}
	azSubsID := os.Getenv(ENV_AZURE_SUBSCRIPTION_ID)
	subscriptionID = &azSubsID
	return
}

func getGraphClientFactory() (*armresourcegraph.ClientFactory, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	// ResourceGraph client
	return armresourcegraph.NewClientFactory(cred, nil)
}
