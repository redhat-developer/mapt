package data

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ImageRequest struct {
	Region string
	ImageReference
}

func GetCommunityGalleryImage(req ImageRequest) (*armcompute.CommunityGalleryImagesClientGetResponse, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")

	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	// for community gallary images
	if len(req.ID) > 0 {
		// extract gallary ID and image name from ID url which looks like:
		// /CommunityGalleries/Fedora-5e266ba4-2250-406d-adad-5d73860d958f/Images/Fedora-Cloud-40-Arm64/Versions/latest
		parts := strings.Split(req.ID, "/")
		if len(parts) != 7 {
			return nil, fmt.Errorf("invalid community gallary image ID: %s", req.ID)
		}
		res, err := clientFactory.NewCommunityGalleryImagesClient().Get(ctx, req.Region, parts[2], parts[4], nil)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
	// for azure offered VM images: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machine-images/get
	// there's a different API to check but currently we only check availability of community images
	return nil, nil
}

func GetCustomImage(req ImageRequest) (*armcompute.ImagesClientGetResponse, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")

	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	if len(req.ID) > 0 {
		// extract resource group and image name from ID url which looks like:
		// /subscriptions/b0ad4737-8299-4c0a-9dd5-959cbcf8d81c/resourceGroups/cloud-importer-resourceGroup-a558d7c1/providers/Microsoft.Compute/images/openshift-local-%s-%s
		parts := strings.Split(req.ID, "/")
		if len(parts) != 9 {
			return nil, fmt.Errorf("invalid custom image ID: %s", req.ID)
		}

		res, err := clientFactory.NewImagesClient().Get(ctx, parts[4], parts[8], nil)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
	return nil, nil
}

func IsImageOffered(req ImageRequest) bool {
	if strings.Contains(req.ID, "CommunityGalleries") {
		if _, err := GetCommunityGalleryImage(req); err != nil {
			logging.Debugf("error while checking if image available at location: %v", err)
			return false
		}
		return true
	}
	if _, err := GetCustomImage(req); err != nil {
		logging.Debugf("error while checking if image available at location: %v", err)
		return false
	}
	return true
}

func SkuG2Support(location string, publisher string, offer string, sku string) (string, error) {
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
	if !verify_g2(imagesClient, location, publisher, offer, sku) {
		finalSKU, err := get_g2_sku(imagesClient, location, publisher, offer, sku)
		if err == nil && finalSKU != "" {
			if verify_g2(imagesClient, location, publisher, offer, finalSKU) {
				fmt.Printf("%s is g1, Using SKU %s\n", sku, finalSKU)
				return finalSKU, nil
			}
		}
	} else {
		return sku, nil
	}
	return "", fmt.Errorf("the SKU %s is not support for G2", sku)
}

func verify_g2(imagesClient *armcompute.VirtualMachineImagesClient, location string, publisher string, offer string, sku string) bool {
	// List available image versions
	resp, err := imagesClient.List(context.Background(), location, publisher, offer, sku, nil)
	if err != nil {
		return false
	}

	image := resp.VirtualMachineImageResourceArray[0]
	version := *image.Name
	resps, _ := imagesClient.Get(context.Background(), location, publisher, offer, sku, version, nil)
	info := resps.VirtualMachineImage
	generation := *info.Properties.HyperVGeneration
	return generation == "V2"
}

func get_g2_sku(imagesClient *armcompute.VirtualMachineImagesClient, location string, publisher string, offer string, originSKU string) (string, error) {
	resp, err := imagesClient.ListSKUs(context.Background(), location, publisher, offer, nil)
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
