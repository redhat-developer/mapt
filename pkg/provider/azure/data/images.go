package data

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

type ImageRequest struct {
	Region string
	ImageReference
}

func IsImageOffered(ctx context.Context, req ImageRequest) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return err
	}
	if len(req.CommunityImageID) > 0 {
		_, err := getCommunityImage(ctx, clientFactory, &req.CommunityImageID, &req.Region)
		return err
	}
	if len(req.SharedImageID) > 0 {
		_, err := getSharedImage(ctx, clientFactory, &req.SharedImageID)
		return err
	}
	// for azure offered VM images: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machine-images/get
	// there's a different API to check but currently we only check availability of community images
	return fmt.Errorf("no valid image to check")
}

func getCommunityImage(ctx context.Context, c *armcompute.ClientFactory, id, region *string) (*armcompute.CommunityGalleryImagesClientGetResponse, error) {
	// extract gallary ID and image name from ID url which looks like:
	// /CommunityGalleries/Fedora-5e266ba4-2250-406d-adad-5d73860d958f/Images/Fedora-Cloud-40-Arm64/Versions/latest
	parts := strings.Split(*id, "/")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid community gallary image ID: %s", *id)
	}
	res, err := c.NewCommunityGalleryImagesClient().Get(ctx, *region, parts[2], parts[4], nil)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func GetSharedImage(ctx context.Context, id *string) (*armcompute.GalleryImageVersionsClientGetResponse, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	c, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(*id, "/")
	if len(parts) != 13 {
		return nil, fmt.Errorf("invalid shared image ID: %s", *id)
	}
	res, err := c.NewGalleryImageVersionsClient().Get(ctx, parts[4], parts[8], parts[10], parts[12], nil)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func getSharedImage(ctx context.Context, c *armcompute.ClientFactory, id *string) (*armcompute.GalleryImageVersionsClientGetResponse, error) {
	parts := strings.Split(*id, "/")
	if len(parts) != 13 {
		return nil, fmt.Errorf("invalid shared image ID: %s", *id)
	}
	res, err := c.NewGalleryImageVersionsClient().Get(ctx, parts[4], parts[8], parts[10], parts[12], nil)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func SkuG2Support(ctx context.Context, location string, publisher string, offer string, sku string) (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")

	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return "", err
	}
	imagesClient := clientFactory.NewVirtualMachineImagesClient()
	if !verify_g2(ctx, imagesClient, location, publisher, offer, sku) {
		finalSKU, err := get_g2_sku(ctx, imagesClient, location, publisher, offer, sku)
		if err == nil && finalSKU != "" {
			if verify_g2(ctx, imagesClient, location, publisher, offer, finalSKU) {
				fmt.Printf("%s is g1, Using SKU %s\n", sku, finalSKU)
				return finalSKU, nil
			}
		}
	} else {
		return sku, nil
	}
	return "", fmt.Errorf("the SKU %s is not support for G2", sku)
}

func verify_g2(ctx context.Context, imagesClient *armcompute.VirtualMachineImagesClient, location string, publisher string, offer string, sku string) bool {
	// List available image versions
	resp, err := imagesClient.List(ctx, location, publisher, offer, sku, nil)
	if err != nil {
		return false
	}

	image := resp.VirtualMachineImageResourceArray[0]
	version := *image.Name
	resps, _ := imagesClient.Get(ctx, location, publisher, offer, sku, version, nil)
	info := resps.VirtualMachineImage
	generation := *info.Properties.HyperVGeneration
	return generation == "V2"
}

func get_g2_sku(ctx context.Context, imagesClient *armcompute.VirtualMachineImagesClient, location string, publisher string, offer string, originSKU string) (string, error) {
	resp, err := imagesClient.ListSKUs(ctx, location, publisher, offer, nil)
	if err != nil {
		return "", err
	}
	for _, sku := range resp.VirtualMachineImageResourceArray {
		if strings.HasPrefix(*sku.Name, originSKU+"-") {
			return *sku.Name, nil
		}
	}
	return "", nil
}
