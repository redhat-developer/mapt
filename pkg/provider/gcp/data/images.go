package data

import (
	"context"
	"fmt"
	"strings"

	gcpAPI "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const (
	rhelImageProject = "rhel-cloud"
)

func GetRHELImageURI(ctx context.Context, version string) (string, error) {
	svc, err := gcpAPI.NewService(ctx, option.WithScopes(gcpAPI.ComputeReadonlyScope))
	if err != nil {
		return "", fmt.Errorf("error creating GCP compute service: %v", err)
	}

	// GCP image families use the major version only (e.g., "rhel-9")
	majorVersion := strings.Split(version, ".")[0]
	family := fmt.Sprintf("rhel-%s", majorVersion)

	image, err := svc.Images.GetFromFamily(rhelImageProject, family).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error getting RHEL image for family %s: %v", family, err)
	}

	return image.SelfLink, nil
}
