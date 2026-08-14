package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilSlices "github.com/redhat-developer/mapt/pkg/util/slices"
)

const (
	azureRetailPricesAPI = "https://prices.azure.com/api/retail/prices"

	// windowsOSPremiumFactor is applied to Linux spot prices when used as a proxy for
	// Windows pricing. Windows spot SKUs include a licensing fee absent from Linux pricing,
	// making Windows spot typically 50-60% more expensive than Linux for the same VM size.
	windowsOSPremiumFactor = 1.6
)

type retailPriceItem struct {
	RetailPrice   float64 `json:"retailPrice"`
	ArmRegionName string  `json:"armRegionName"`
	ArmSkuName    string  `json:"armSkuName"`
	SkuName       string  `json:"skuName"`
}

type retailPricesPage struct {
	Items        []retailPriceItem `json:"Items"`
	NextPageLink *string           `json:"NextPageLink"`
}

// checkSpotPricingRetailAPI is called when the Resource Graph returns no spot pricing data.
// It queries the public Azure Retail Prices API which requires no authentication.
func checkSpotPricingRetailAPI(ctx context.Context, mCtx *mc.Context, locations []string, args checkSpotPricingArgs) (map[string][]spotPricingResult, error) {
	logging.Debugf("Resource Graph returned no spot pricing data, falling back to Azure Retail Prices API")
	var allResults []spotPricingResult
	for _, location := range locations {
		results, err := retailPricesForLocation(ctx, location, args, mCtx.Debug())
		if err != nil {
			logging.Debugf("Retail Prices API error for location %s: %v", location, err)
			continue
		}
		allResults = append(allResults, results...)
	}
	return utilSlices.Split(allResults, func(s spotPricingResult) string {
		return s.Location
	}), nil
}

func retailPricesForLocation(ctx context.Context, location string, args checkSpotPricingArgs, debug bool) ([]spotPricingResult, error) {
	nextURL := azureRetailPricesAPI + "?$filter=" + encodeODataFilter(buildRetailPricesFilter(location))
	var allItems []retailPriceItem
	for nextURL != "" {
		page, err := fetchRetailPricesPage(ctx, nextURL)
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, page.Items...)
		if page.NextPageLink == nil || *page.NextPageLink == "" {
			break
		}
		nextURL = *page.NextPageLink
	}

	results := convertRetailItems(allItems, location, args, debug)
	if len(results) == 0 && len(allItems) > 0 {
		// No OS-specific spot SKUs found (e.g. Windows spot not yet listed for newer VM series).
		// Fall back to any spot SKU for the requested compute sizes as a price estimate.
		if debug {
			logging.Debugf("Retail Prices API: no %s-specific spot SKUs for location %s, using OS-agnostic pricing", args.osType, location)
		}
		results = convertRetailItemsAnyOS(allItems, location, args)
	}
	return results, nil
}

// buildRetailPricesFilter returns an OData filter for the Azure Retail Prices API.
// The API only supports eq/and/contains — or is not accepted, so SKU filtering
// is done client-side in convertRetailItems rather than here.
func buildRetailPricesFilter(location string) string {
	return fmt.Sprintf(
		"serviceName eq 'Virtual Machines' and priceType eq 'Consumption' and contains(skuName, 'Spot') and armRegionName eq '%s'",
		location,
	)
}

func fetchRetailPricesPage(ctx context.Context, apiURL string) (*retailPricesPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating retail prices request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying retail prices API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading retail prices response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("retail prices API status %d: %s", resp.StatusCode, string(body))
	}
	var page retailPricesPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parsing retail prices response: %w", err)
	}
	return &page, nil
}

// encodeODataFilter encodes an OData filter value for a URL query string.
// url.QueryEscape encodes spaces as '+' and parens/commas as %28/%29/%2C, which
// breaks the Azure Retail Prices API OData parser. Azure expects %20 for spaces
// and literal parens/commas as structural characters in the filter expression.
func encodeODataFilter(filter string) string {
	encoded := url.QueryEscape(filter)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "%28", "(")
	encoded = strings.ReplaceAll(encoded, "%29", ")")
	encoded = strings.ReplaceAll(encoded, "%2C", ",")
	return encoded
}

// convertRetailItems filters items by OS type and compute size, keeping the lowest
// price per SKU when duplicates exist.
func convertRetailItems(items []retailPriceItem, location string, args checkSpotPricingArgs, debug bool) []spotPricingResult {
	bestPrice := make(map[string]spotPricingResult)
	for _, item := range items {
		if !slices.ContainsFunc(args.computeSizes, func(cs string) bool {
			return strings.EqualFold(cs, item.ArmSkuName)
		}) {
			continue
		}
		if !retailItemMatchesOSType(item.SkuName, args.osType) {
			continue
		}
		if debug {
			logging.Debugf("Retail Prices API: found %s at %s price %.4f", item.ArmSkuName, location, item.RetailPrice)
		}
		if existing, ok := bestPrice[item.ArmSkuName]; !ok || item.RetailPrice < existing.Price {
			bestPrice[item.ArmSkuName] = spotPricingResult{
				ComputeSize: item.ArmSkuName,
				OSType:      args.osType,
				Location:    location,
				Price:       item.RetailPrice,
			}
		}
	}
	results := make([]spotPricingResult, 0, len(bestPrice))
	for _, r := range bestPrice {
		results = append(results, r)
	}
	return results
}

// convertRetailItemsAnyOS is a secondary pass that ignores OS type, used when
// the Retail API has no OS-specific spot SKUs for the requested compute sizes
// (e.g. Windows spot not yet listed for newer VM series). When the target OS is
// Windows, prices are scaled by windowsOSPremiumFactor to account for the
// Windows licensing fee that Linux spot prices do not include.
func convertRetailItemsAnyOS(items []retailPriceItem, location string, args checkSpotPricingArgs) []spotPricingResult {
	bestPrice := make(map[string]spotPricingResult)
	for _, item := range items {
		if !slices.ContainsFunc(args.computeSizes, func(cs string) bool {
			return strings.EqualFold(cs, item.ArmSkuName)
		}) {
			continue
		}
		price := item.RetailPrice
		if strings.EqualFold(args.osType, "windows") {
			price *= windowsOSPremiumFactor
		}
		if existing, ok := bestPrice[item.ArmSkuName]; !ok || price < existing.Price {
			bestPrice[item.ArmSkuName] = spotPricingResult{
				ComputeSize: item.ArmSkuName,
				OSType:      args.osType,
				Location:    location,
				Price:       price,
			}
		}
	}
	results := make([]spotPricingResult, 0, len(bestPrice))
	for _, r := range bestPrice {
		results = append(results, r)
	}
	return results
}

// retailItemMatchesOSType checks whether a SKU name matches the requested OS type.
// Windows SKUs contain "windows" in the name; Linux SKUs do not.
func retailItemMatchesOSType(skuName, osType string) bool {
	isWindows := strings.Contains(strings.ToLower(skuName), "windows")
	return strings.EqualFold(osType, "windows") == isWindows
}
