package data

import (
	"testing"

	"github.com/IBM/vpc-go-sdk/vpcv1"
)

func TestMatchesCatalogOffering(t *testing.T) {
	crn := "crn:v1:bluemix:public:globally-unique-id"
	available := vpcv1.ImageStatusAvailableConst
	pending := "pending"
	id := "image-id"
	otherCRN := "crn:v1:bluemix:public:other"

	tests := []struct {
		name     string
		image    vpcv1.Image
		expected bool
	}{
		{
			name: "available image with the offering CRN",
			image: vpcv1.Image{
				ID:     &id,
				Status: &available,
				CatalogOffering: &vpcv1.ImageCatalogOffering{
					Version: &vpcv1.CatalogOfferingVersionReference{CRN: &crn},
				},
			},
			expected: true,
		},
		{
			name: "image with another offering CRN",
			image: vpcv1.Image{
				ID:     &id,
				Status: &available,
				CatalogOffering: &vpcv1.ImageCatalogOffering{
					Version: &vpcv1.CatalogOfferingVersionReference{CRN: &otherCRN},
				},
			},
			expected: false,
		},
		{
			name: "image is not available",
			image: vpcv1.Image{
				ID:     &id,
				Status: &pending,
				CatalogOffering: &vpcv1.ImageCatalogOffering{
					Version: &vpcv1.CatalogOfferingVersionReference{CRN: &crn},
				},
			},
			expected: false,
		},
		{
			name: "image has no catalog offering",
			image: vpcv1.Image{
				ID:     &id,
				Status: &available,
			},
			expected: false,
		},
		{
			name: "offering has no version",
			image: vpcv1.Image{
				ID:              &id,
				Status:          &available,
				CatalogOffering: &vpcv1.ImageCatalogOffering{},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := matchesCatalogOffering(test.image, crn)
			if got != test.expected {
				t.Errorf("expected %v, got %v", test.expected, got)
			}
		})
	}
}
