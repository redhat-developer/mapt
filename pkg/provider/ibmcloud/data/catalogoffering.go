package data

import (
	"fmt"
	"os"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/catalogmanagementv1"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"golang.org/x/mod/semver"
)

const (
	catalogManagementURL          = "https://cm.globalcatalog.cloud.ibm.com/api/v1-beta"
	consumptionOfferingsPageLimit = 100
	offeringNamePrefix            = "openshift-local-"
)

func GetCatalogOfferingVersionCRN(offeringName, version string) (string, error) {
	client, err := catalogManagementService()
	if err != nil {
		return "", fmt.Errorf("creating catalog management client: %w", err)
	}

	var offset int64
	for {
		opts := client.NewGetConsumptionOfferingsOptions()
		opts.SetLimit(consumptionOfferingsPageLimit)
		opts.SetOffset(offset)

		result, _, err := client.GetConsumptionOfferings(opts)
		if err != nil {
			return "", fmt.Errorf("listing consumption offerings: %w", err)
		}

		for _, offering := range result.Resources {
			if offering.Name == nil || *offering.Name != offeringName {
				continue
			}
			logging.Debugf("found offering %s (id=%s)", *offering.Name, *offering.ID)
			for _, kind := range offering.Kinds {
				for _, v := range kind.Versions {
					if v.Version != nil && *v.Version == version && v.CRN != nil {
						return *v.CRN, nil
					}
				}
			}
			return "", fmt.Errorf("offering %q found but version %q not available", offeringName, version)
		}

		offset += consumptionOfferingsPageLimit
		if result.TotalCount == nil || offset >= *result.TotalCount {
			break
		}
	}
	return "", fmt.Errorf("offering %q not found in any accessible catalog", offeringName)
}

func GetLatestCatalogOfferingVersionCRN(arch string) (string, string, error) {
	client, err := catalogManagementService()
	if err != nil {
		return "", "", fmt.Errorf("creating catalog management client: %w", err)
	}

	type versionEntry struct {
		version string
		crn     string
	}
	var candidates []versionEntry
	archSuffix := "-" + strings.ReplaceAll(arch, "_", "-")

	var offset int64
	for {
		opts := client.NewGetConsumptionOfferingsOptions()
		opts.SetLimit(consumptionOfferingsPageLimit)
		opts.SetOffset(offset)

		result, _, err := client.GetConsumptionOfferings(opts)
		if err != nil {
			return "", "", fmt.Errorf("listing consumption offerings: %w", err)
		}

		for _, offering := range result.Resources {
			if offering.Name == nil {
				continue
			}
			name := *offering.Name
			if !strings.HasPrefix(name, offeringNamePrefix) || !strings.HasSuffix(name, archSuffix) {
				continue
			}
			versionDashed := strings.TrimSuffix(strings.TrimPrefix(name, offeringNamePrefix), archSuffix)
			version := strings.ReplaceAll(versionDashed, "-", ".")

			if !semver.IsValid("v" + version) {
				logging.Debugf("skipping offering %s: %q is not valid semver", name, version)
				continue
			}

			for _, kind := range offering.Kinds {
				for _, v := range kind.Versions {
					if v.Version != nil && *v.Version == version && v.CRN != nil {
						candidates = append(candidates, versionEntry{version: version, crn: *v.CRN})
						break
					}
				}
				if len(candidates) > 0 && candidates[len(candidates)-1].version == version {
					break
				}
			}
		}

		offset += consumptionOfferingsPageLimit
		if result.TotalCount == nil || offset >= *result.TotalCount {
			break
		}
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("no openshift-local offerings found for arch %s", arch)
	}

	best := 0
	for i := 1; i < len(candidates); i++ {
		if semver.Compare("v"+candidates[i].version, "v"+candidates[best].version) > 0 {
			best = i
		}
	}

	logging.Debugf("Found %d openshift-local offerings, latest: %s", len(candidates), candidates[best].version)
	return candidates[best].version, candidates[best].crn, nil
}

func catalogManagementService() (*catalogmanagementv1.CatalogManagementV1, error) {
	return catalogmanagementv1.NewCatalogManagementV1(
		&catalogmanagementv1.CatalogManagementV1Options{
			Authenticator: &core.IamAuthenticator{
				ApiKey: os.Getenv(icConstants.EnvIBMCloudAPIKey),
			},
			URL: catalogManagementURL,
		})
}
