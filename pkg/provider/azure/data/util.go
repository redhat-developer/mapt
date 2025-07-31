package data

import (
	"fmt"
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

func getGraphClient() (*armresourcegraph.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// ResourceGraph client
	return armresourcegraph.NewClient(cred, nil)
}
