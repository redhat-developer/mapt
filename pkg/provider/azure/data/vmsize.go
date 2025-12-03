package data

import (
	"context"
	"os"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/redhat-developer/mapt/pkg/util"
)

func IsVMSizeOfferedByLocation(ctx context.Context, vmSize, location string) (bool, error) {
	offerings, err := FilterVMSizeOfferedByLocation(ctx, []string{vmSize}, location)
	return len(offerings) == 1, err
}

// Get InstanceTypes offerings on current location
func FilterVMSizeOfferedByLocation(ctx context.Context, vmSizes []string, location string) ([]string, error) {
	// Create a new Azure credential (uses environment variables or managed identity)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	client := clientFactory.NewResourceSKUsClient()
	p := client.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: &location,
	})
	var offerings []string
	for p.More() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sku := range page.Value {
			if slices.Contains(vmSizes, *sku.Name) &&
				!slices.Contains(offerings, *sku.Name) {
				if slices.Contains(
					util.ArrayConvert(
						sku.Locations,
						func(l *string) string {
							return *l
						}),
					location) {
					offerings = append(offerings, *sku.Name)
				}
			}
		}
	}
	return offerings, nil
}
