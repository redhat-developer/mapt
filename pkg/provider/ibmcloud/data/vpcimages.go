package data

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/IBM/go-sdk-core/v5/core"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
)

const (
	VPC_ARCH_X86_64 vpcArch = "amd64"
	VPC_ARCH_IBMZ   vpcArch = "s390x"
)

type vpcArch string

type VPCImageArgs struct {
	Name string
	Arch vpcArch
}

func GetVPCImage(args *VPCImageArgs) (*string, error) {
	return getVPCImage(nil, args)
}

func getVPCImage(next *string, args *VPCImageArgs) (*string, error) {
	vpcService, err := vpcService()
	if err != nil {
		return nil, err
	}
	images, _, err := vpcService.ListImages(
		&vpcv1.ListImagesOptions{
			Start: next,
		})
	if err != nil {
		return nil, err
	}
	idx := slices.IndexFunc(images.Images,
		func(i vpcv1.Image) bool {
			return *i.OperatingSystem.Architecture == string(args.Arch) &&
				strings.Contains(*i.Name, args.Name)
		})
	if idx != -1 {
		return images.Images[idx].ID, nil
	}
	if images.Next != nil {
		next, err := images.GetNextStart()
		if err != nil {
			return nil, err
		}
		return getVPCImage(next, args)
	}
	return nil, fmt.Errorf("no image %s available", args.Name)
}

// GetImageIDByCatalogOfferingCRN returns the ID of the VPC image that belongs
// to the catalog offering version.
//
// A bare metal server needs an image ID: BareMetalServerPrototype has no
// catalog offering field, unlike the virtual server instance prototype. The
// image is the one published for the offering, so the bare metal server runs
// the same SNC bundle as the instance.
func GetImageIDByCatalogOfferingCRN(catalogOfferingCRN string) (string, error) {
	svc, err := vpcService()
	if err != nil {
		return "", err
	}
	var start *string
	for {
		images, _, err := svc.ListImages(
			&vpcv1.ListImagesOptions{
				Start: start,
			})
		if err != nil {
			return "", fmt.Errorf("listing VPC images: %w", err)
		}
		for _, image := range images.Images {
			if matchesCatalogOffering(image, catalogOfferingCRN) {
				return *image.ID, nil
			}
		}
		if images.Next == nil {
			break
		}
		next, err := images.GetNextStart()
		if err != nil {
			return "", fmt.Errorf("paginating VPC images: %w", err)
		}
		if next == nil {
			break
		}
		start = next
	}
	return "", fmt.Errorf("no VPC image found for catalog offering %s", catalogOfferingCRN)
}

// matchesCatalogOffering reports whether the image belongs to the catalog
// offering version and is available for use.
func matchesCatalogOffering(image vpcv1.Image, catalogOfferingCRN string) bool {
	if image.ID == nil {
		return false
	}
	if image.Status != nil && *image.Status != vpcv1.ImageStatusAvailableConst {
		return false
	}
	if image.CatalogOffering == nil || image.CatalogOffering.Version == nil {
		return false
	}
	if image.CatalogOffering.Version.CRN == nil {
		return false
	}
	return *image.CatalogOffering.Version.CRN == catalogOfferingCRN
}

func vpcService() (*vpcv1.VpcV1, error) {
	serviceURL, err := vpcv1.GetServiceURLForRegion(os.Getenv("IC_REGION"))
	if err != nil {
		return nil, err
	}
	return vpcv1.NewVpcV1(&vpcv1.VpcV1Options{
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv(icConstants.EnvIBMCloudAPIKey),
		},
		URL: serviceURL,
	})
}
