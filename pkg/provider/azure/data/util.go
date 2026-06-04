package data

import (
	"os"
	"strings"

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

// splitDiskControllerTypes splits a comma-separated disk controller type string
// (e.g. "SCSI,NVMe") and trims whitespace from each element.
func splitDiskControllerTypes(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		out = append(out, strings.TrimSpace(t))
	}
	return out
}

func getGraphClientFactory() (*armresourcegraph.ClientFactory, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	// ResourceGraph client
	return armresourcegraph.NewClientFactory(cred, nil)
}
