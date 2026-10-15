package data

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/mapt/pkg/util/logging"
	gcpAPI "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func GetAvailabilityZones(ctx context.Context, region string) ([]string, error) {
	svc, err := gcpAPI.NewService(ctx, option.WithScopes(gcpAPI.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("error creating GCP compute service: %v", err)
	}

	project, err := getProject(ctx)
	if err != nil {
		return nil, err
	}

	zoneList, err := svc.Zones.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing zones: %v", err)
	}

	var zones []string
	for _, z := range zoneList.Items {
		if strings.HasPrefix(z.Name, region+"-") && z.Status == "UP" {
			zones = append(zones, z.Name)
		}
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("no available zones found in region %s", region)
	}

	logging.Debugf("found %d zones in region %s: %v", len(zones), region, zones)
	return zones, nil
}

func getProject(ctx context.Context) (string, error) {
	project := os.Getenv("GOOGLE_PROJECT")
	if len(project) == 0 {
		project = os.Getenv("GCLOUD_PROJECT")
	}
	if len(project) == 0 {
		project = os.Getenv("CLOUDSDK_CORE_PROJECT")
	}
	if len(project) == 0 {
		return "", fmt.Errorf("GCP project not set; set GOOGLE_PROJECT environment variable")
	}
	return project, nil
}
