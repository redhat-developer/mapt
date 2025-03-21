package data

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ImageRequest struct {
	Region string
	ImageReference
}

func GetImage(req ImageRequest) (*armcompute.CommunityGalleryImagesClientGetResponse, error) {
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

func IsImageOffered(req ImageRequest) bool {
	if _, err := GetImage(req); err != nil {
		logging.Debugf("error while checking if image available at location: %v", err)
		return false
	}
	return true
}
