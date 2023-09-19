package spotprice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	querySpotPrice = "SpotResources | where type =~ 'microsoft.compute/skuspotpricehistory/ostype/location' " +
		"and sku.name in~ (%s) and properties.osType =~ '%s'" +
		"| project skuName=tostring(sku.name),osType=tostring(properties.osType)," +
		"location,latestSpotPriceUSD=todouble(properties.spotPrices[0].priceUSD)" +
		"| order by latestSpotPriceUSD asc"

	queryEvictionRate = "SpotResources | where type =~ 'microsoft.compute/skuspotevictionrate/location' " +
		"and sku.name in~ (%s)" +
		"| project skuName=tostring(sku.name),location,spotEvictionRate=tostring(properties.evictionRate) "

	// lowest eviction rate possible
	minEvictionRate = "0-5"
)

type BestSpotChoiceRequest struct {
	VMTypes []string
	OSType  string
}

type BestSpotChoiceResponse struct {
	VMType   string  `json:"skuName"`
	Location string  `json:"location"`
	Price    float64 `json:"latestSpotPriceUSD"`
}

type priceHistory struct {
	VMType   string  `json:"skuName"`
	OSType   string  `json:"osType"`
	Location string  `json:"location"`
	Price    float64 `json:"latestSpotPriceUSD"`
}

type evictionRate struct {
	VMType       string `json:"skuName"`
	Location     string `json:"location"`
	EvictionRate string `json:"spotEvictionRate"`
}

var (
	azIdentityEnvs = []string{
		"AZURE_TENANT_ID",
		"AZURE_SUBSCRIPTION_ID",
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
	}
)

// This function will return the best spot option
func GetBestSpotChoice(r BestSpotChoiceRequest) (*BestSpotChoiceResponse, error) {
	client, err := getGraphClient()
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// Context for requests
	ctx := context.Background()
	// Run spot price history request
	phr, err := getPriceHistory(ctx, client, r)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// Run eviction rate request
	evrr, err := getEvictionRate(ctx, client, r)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// Compare prices and evictions
	return getBestSpotChoice(phr, evrr)
}

func getGraphClient() (*armresourcegraph.Client, error) {
	// Auth identity
	setAZIdentityEnvs()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// ResourceGraph client
	return armresourcegraph.NewClient(cred, nil)
}

// Envs required for auth with go sdk
// https://learn.microsoft.com/es-es/azure/developer/go/azure-sdk-authentication?tabs=bash#service-principal-with-a-secret
// do not match standard envs for pulumi envs for auth with native sdk
// https://www.pulumi.com/registry/packages/azure-native/installation-configuration/#set-configuration-using-environment-variables
func setAZIdentityEnvs() {
	for _, e := range azIdentityEnvs {
		os.Setenv(e,
			os.Getenv(strings.ReplaceAll(e, "AZURE", "ARM")))
	}
}

func getPriceHistory(ctx context.Context, client *armresourcegraph.Client,
	r BestSpotChoiceRequest) ([]priceHistory, error) {
	spr := fmt.Sprintf(querySpotPrice,
		fmt.Sprintf("'%s'", strings.Join(r.VMTypes, ",")),
		r.OSType)
	qr, err := client.Resources(ctx,
		armresourcegraph.QueryRequest{
			Query: to.Ptr(spr),
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error getting spot price history: %v", err)
	}
	var results []priceHistory
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("error getting spot price history: %v", err)
		}
		rStruct := priceHistory{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			return nil, fmt.Errorf("error getting spot price history: %v", err)
		}
		results = append(results, rStruct)
	}
	logging.Debugf("spot prices history %v", results)
	return results, nil
}

func getEvictionRate(ctx context.Context, client *armresourcegraph.Client,
	r BestSpotChoiceRequest) ([]evictionRate, error) {
	evrr := fmt.Sprintf(queryEvictionRate,
		fmt.Sprintf("'%s'", strings.Join(r.VMTypes, ",")))
	qr, err := client.Resources(ctx,
		armresourcegraph.QueryRequest{
			Query: to.Ptr(evrr),
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error getting eviction rate: %v", err)
	}
	var results []evictionRate
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("error getting eviction rate: %v", err)
		}
		rStruct := evictionRate{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			return nil, fmt.Errorf("error getting eviction rate: %v", err)
		}
		results = append(results, rStruct)
	}
	return results, nil
}

func getBestSpotChoice(s []priceHistory, e []evictionRate) (*BestSpotChoiceResponse, error) {
	var evm map[string]string = make(map[string]string)
	for _, ev := range e {
		evm[fmt.Sprintf("%s%s", ev.Location, ev.VMType)] = ev.EvictionRate
	}
	for _, sv := range s {
		er, ok := evm[fmt.Sprintf("%s%s", sv.Location, sv.VMType)]
		if ok && er == minEvictionRate {
			return &BestSpotChoiceResponse{
				VMType:   sv.VMType,
				Location: sv.Location,
				Price:    sv.Price,
			}, nil
		}
	}
	return nil, fmt.Errorf("could not find any spot with minimum eviction rate")
}
