package blob

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func newClient() (*azblob.Client, error) {
	storageAccount := os.Getenv("AZURE_STORAGE_ACCOUNT")
	if len(storageAccount) == 0 {
		return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT environment variable is not set")
	}
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", storageAccount)

	if key := os.Getenv("AZURE_STORAGE_KEY"); key != "" {
		cred, err := azblob.NewSharedKeyCredential(storageAccount, key)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared key credential: %w", err)
		}
		return azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	}

	if sasToken := os.Getenv("AZURE_STORAGE_SAS_TOKEN"); sasToken != "" {
		return azblob.NewClientWithNoCredential(serviceURL+"?"+sasToken, nil)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}
	return azblob.NewClient(serviceURL, cred, nil)
}

func Delete(ctx context.Context, containerName, prefix string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	return deleteBlobs(ctx, client, containerName, prefix)
}

func deleteBlobs(ctx context.Context, client *azblob.Client, containerName, prefix string) error {
	blobNames, err := listBlobNames(ctx, client, containerName, prefix)
	if err != nil {
		return err
	}
	for _, name := range blobNames {
		if _, err := client.DeleteBlob(ctx, containerName, name, nil); err != nil {
			logging.Errorf("failed to delete blob %s: %v", name, err)
		}
	}
	return nil
}

func listBlobNames(ctx context.Context, client *azblob.Client, containerName, prefix string) ([]string, error) {
	var names []string
	pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing blobs: %w", err)
		}
		for _, blob := range page.Segment.BlobItems {
			names = append(names, *blob.Name)
		}
	}
	return names, nil
}
