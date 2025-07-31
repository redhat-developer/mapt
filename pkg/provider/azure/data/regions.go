package data

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

func GetLocations() ([]string, error) {
	cred, subscriptionID, err := getCredentials()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	client, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, err
	}
	var locations []string
	pager := client.NewListLocationsPager(*subscriptionID, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, loc := range page.Value {
			locations = append(locations, *loc.Name)
		}
	}
	return locations, nil
}
