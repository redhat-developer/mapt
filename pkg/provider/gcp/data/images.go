package data

import (
	"context"
	"fmt"
	"strings"

	gcpAPI "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const (
	rhelImageProject    = "rhel-cloud"
	windowsImageProject = "windows-cloud"
	fedoraImageProject  = "fedora-cloud"
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

func GetWindowsImageURI(ctx context.Context, version string) (string, error) {
	svc, err := gcpAPI.NewService(ctx, option.WithScopes(gcpAPI.ComputeReadonlyScope))
	if err != nil {
		return "", fmt.Errorf("error creating GCP compute service: %v", err)
	}

	family := fmt.Sprintf("windows-%s", version)

	image, err := svc.Images.GetFromFamily(windowsImageProject, family).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error getting Windows image for family %s: %v", family, err)
	}

	return image.SelfLink, nil
}

func GetFedoraImageURI(ctx context.Context, version string) (string, error) {
	svc, err := gcpAPI.NewService(ctx, option.WithScopes(gcpAPI.ComputeReadonlyScope))
	if err != nil {
		return "", fmt.Errorf("error creating GCP compute service: %v", err)
	}

	family := fmt.Sprintf("fedora-cloud-%s-x86-64", version)

	image, err := svc.Images.GetFromFamily(fedoraImageProject, family).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error getting Fedora image for family %s: %v", family, err)
	}

	return image.SelfLink, nil
}
