package data

import (
	"fmt"
	"os"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/catalogmanagementv1"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	catalogManagementURL          = "https://cm.globalcatalog.cloud.ibm.com/api/v1-beta"
	consumptionOfferingsPageLimit = 100
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

func catalogManagementService() (*catalogmanagementv1.CatalogManagementV1, error) {
	return catalogmanagementv1.NewCatalogManagementV1(
		&catalogmanagementv1.CatalogManagementV1Options{
			Authenticator: &core.IamAuthenticator{
				ApiKey: os.Getenv(icConstants.EnvIBMCloudAPIKey),
			},
			URL: catalogManagementURL,
		})
}
