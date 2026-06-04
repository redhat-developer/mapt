package data

import (
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
)

var azureIdentityEnvs = []string{
	"AZURE_TENANT_ID",
	"AZURE_SUBSCRIPTION_ID",
	"AZURE_CLIENT_ID",
	"AZURE_CLIENT_SECRET",
}

// ensureAzureEnvs maps ARM_* env vars to AZURE_* if the AZURE_* vars are unset.
// Safe to call multiple times — only sets vars that are currently empty.
func ensureAzureEnvs() {
	for _, e := range azureIdentityEnvs {
		if os.Getenv(e) == "" {
			armKey := strings.ReplaceAll(e, "AZURE", "ARM")
			if v := os.Getenv(armKey); v != "" {
				_ = os.Setenv(e, v)
			}
		}
	}
}

// SubscriptionID returns the Azure subscription ID, checking AZURE_SUBSCRIPTION_ID
// first, then falling back to ARM_SUBSCRIPTION_ID (Pulumi/Terraform convention).
func SubscriptionID() string {
	if v := os.Getenv("AZURE_SUBSCRIPTION_ID"); v != "" {
		return v
	}
	return os.Getenv("ARM_SUBSCRIPTION_ID")
}

func getCredentials() (cred *azidentity.DefaultAzureCredential, subscriptionID *string, err error) {
	ensureAzureEnvs()
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return
	}
	azSubsID := SubscriptionID()
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
	ensureAzureEnvs()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	// ResourceGraph client
	return armresourcegraph.NewClientFactory(cred, nil)
}
