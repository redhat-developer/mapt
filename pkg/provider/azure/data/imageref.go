package data

import (
	"fmt"
	"strings"
)

type OSType int

const (
	Ubuntu OSType = iota + 1
	RHEL
	Fedora
)

var (
	FedoraDefaultVersion string = "43"
)

const fedoraImageGalleryBase = "/CommunityGalleries/Fedora-5e266ba4-2250-406d-adad-5d73860d958f/Images/"

// /subscriptions/02db6bd4-035c-4074-b699-468f3d914744/resourceGroups/RHEL-AI-CUDA-AZURE-3.0.0/providers/Microsoft.Compute/galleries/rhel_ai_cuda_azure_3.0.0/images/rhel-ai-cuda-azure-3.0.0/versions/1.0.0

type ImageReference struct {
	// Market Place
	Publisher string
	Offer     string
	Sku       string
	// Community
	CommunityImageID string
	// // Private Shared
	SharedImageID string
}

var (
	defaultImageRefs = map[OSType]map[string]ImageReference{
		RHEL: {
			"x86_64": {
				Publisher: "RedHat",
				Offer:     "RHEL",
				Sku:       "%s_%s",
			},
			"arm64": {
				Publisher: "RedHat",
				Offer:     "rhel-arm64",
				Sku:       "%s_%s-arm64",
			},
		},
		Ubuntu: {
			"x86_64": {
				Publisher: "Canonical",
				Offer:     "ubuntu-%s_%s-lts-daily",
				Sku:       "server",
			},
		},
		Fedora: {
			"x86_64": {
				CommunityImageID: fedoraImageGalleryBase + "Fedora-Cloud-%s-x64/Versions/latest",
			},
			"arm64": {
				CommunityImageID: fedoraImageGalleryBase + "Fedora-Cloud-%s-Arm64/Versions/latest",
			},
		},
	}
)

// version should came in format X.Y (major.minor)
func GetImageRef(osTarget OSType, arch string, version string) (*ImageReference, error) {
	ir := defaultImageRefs[osTarget][arch]
	versions := strings.Split(version, ".")
	switch osTarget {
	case Ubuntu:
		return &ImageReference{
			Publisher: ir.Publisher,
			Offer:     fmt.Sprintf(ir.Offer, versions[0], versions[1]),
			Sku:       ir.Sku,
		}, nil
	case RHEL:
		return &ImageReference{
			Publisher: ir.Publisher,
			Offer:     ir.Offer,
			Sku:       fmt.Sprintf(ir.Sku, versions[0], versions[1]),
		}, nil
	case Fedora:
		return &ImageReference{
			CommunityImageID: fmt.Sprintf(ir.CommunityImageID, versions[0]),
		}, nil
	}
	return nil, fmt.Errorf("os type not supported")
}
