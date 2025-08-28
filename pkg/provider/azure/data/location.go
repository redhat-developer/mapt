package data

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/redhat-developer/mapt/pkg/util"
)

func Locations() ([]string, error) {
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

func LocationsBySupportedResourceType(rt ResourceType) ([]string, error) {
	cred, subscriptionID, err := getCredentials()
	if err != nil {
		return nil, err
	}
	client, err := armresources.NewProvidersClient(*subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Get(context.Background(), "Microsoft.Network", nil)
	if err != nil {
		return nil, err
	}
	var locationsDN []string
	for _, sku := range resp.ResourceTypes {
		if strings.EqualFold(*sku.ResourceType, string(rt)) {
			for _, loc := range sku.Locations {
				locationsDN = append(locationsDN, *loc)
			}
		}
	}
	return translate(cred, subscriptionID, locationsDN)
}

func translate(cred *azidentity.DefaultAzureCredential,
	subscriptionID *string, lDisplayName []string) ([]string, error) {
	locationsClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, err
	}
	locationPager := locationsClient.NewListLocationsPager(*subscriptionID, nil)
	var locationsMap = make(map[string]string)
	for locationPager.More() {
		page, err := locationPager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, loc := range page.Value {
			if loc.Name != nil && loc.DisplayName != nil {
				locationsMap[*loc.DisplayName] = *loc.Name
			}
		}
	}
	return util.ArrayConvert(lDisplayName,
		func(dn string) string {
			return locationsMap[dn]
		}), nil
}
