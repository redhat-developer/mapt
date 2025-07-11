package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
)

const (
	querySpotPrice = "SpotResources | where type =~ 'microsoft.compute/skuspotpricehistory/ostype/location' " +
		"and sku.name in~ ({{range $index, $v := .VMTypes}}{{if $index}},{{end}}'{{$v}}'{{end}}) and properties.osType =~ '{{.OSType}}'" +
		"| project skuName=tostring(sku.name),osType=tostring(properties.osType)," +
		"location,latestSpotPriceUSD=todouble(properties.spotPrices[0].priceUSD)" +
		"| order by latestSpotPriceUSD asc"

	queryEvictionRate = "SpotResources | where type =~ 'microsoft.compute/skuspotevictionrate/location' " +
		"and sku.name in~ ({{range $index, $v := .VMTypes}}{{if $index}},{{end}}'{{$v}}'{{end}})" +
		"| project skuName=tostring(sku.name),location,spotEvictionRate=tostring(properties.evictionRate) "

	Lowest EvictionRate = iota
	Low
	Medium
	High
	Highest

	DefaultEvictionRate = Lowest
)

type EvictionRate int

type BestSpotChoiceRequest struct {
	VMTypes               []string
	OSType                string
	EvictionRateTolerance EvictionRate
	ImageRef              ImageReference
	ExcludedRegions       []string
}

type BestSpotChoiceResponse struct {
	VMType       string  `json:"skuName"`
	Location     string  `json:"location"`
	Price        float64 `json:"latestSpotPriceUSD"`
	EvictionRate string  `json:"evictionRate"`
}

type priceHistory struct {
	VMType   string  `json:"skuName"`
	OSType   string  `json:"osType"`
	Location string  `json:"location"`
	Price    float64 `json:"latestSpotPriceUSD"`
}

type evictionRateSpec struct {
	id    EvictionRate
	name  string
	order int
	value string
}

type evictionRate struct {
	VMType       string `json:"skuName"`
	Location     string `json:"location"`
	EvictionRate string `json:"spotEvictionRate"`
}

var evictionRates = map[string]evictionRateSpec{
	"lowest":  {Lowest, "lowest", 0, "0-5"},
	"low":     {Low, "low", 1, "5-10"},
	"medium":  {Medium, "medium", 2, "10-15"},
	"high":    {High, "high", 3, "15-20"},
	"highest": {Highest, "highest", 4, "20+"},
}

type SpotSelector struct{}

func NewSpotSelector() *SpotSelector { return &SpotSelector{} }

func (c *SpotSelector) Select(
	args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
	return lowestPrice(args)
}

func lowestPrice(args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
	var err error
	vms := args.ComputeTypes
	if len(vms) == 0 {
		vms, err =
			NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	spr := BestSpotChoiceRequest{
		VMTypes:               vms,
		OSType:                osType(args.OS),
		EvictionRateTolerance: EvictionRate(args.SpotTolerance),
	}
	prices, err := GetBestSpotChoice(spr)
	if err != nil {
		return nil, err
	}
	cl := 0
	evr, ok := parseEvictionRate(prices.EvictionRate)
	if ok {
		cl = int(evr)
	}
	return &spotTypes.SpotResults{
		ComputeType: prices.VMType,
		Region:      prices.Location,
		Price:       float32(prices.Price),
		ChanceLevel: cl,
	}, nil
}

func osType(os *string) string {
	if os == nil {
		return "linux"
	}
	switch *os {
	case "fedora", "RHEL", "rhel", "ubuntu":
		return "linux"
	case "windows", "Windows":
		return "windows"
	default:
		return ""
	}
}

// var ErrEvictionRatesEmtpyData = fmt.Errorf("error eviction rates are returning empty")

// This function will return the best spot option
func GetBestSpotChoice(r BestSpotChoiceRequest) (*BestSpotChoiceResponse, error) {
	// TODO REVIEW THIS
	if r.EvictionRateTolerance == 0 {
		r.EvictionRateTolerance = DefaultEvictionRate
	}
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
	// Run eviction rate request it will get all vm types with each eviction rate
	evrr, err := getEvictionRateInfoByVMTypes(ctx, client, r.VMTypes, r.ExcludedRegions)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	if len(evrr) == 0 {
		logging.Debugf("can not get information about eviction rates, we will continue only based on prices")
		return getSpotChoiceByPrice(phr, r.ImageRef.ID)
	}
	// Compare prices and evictions
	return getBestSpotChoice(phr, evrr, Lowest, r.EvictionRateTolerance, r.ImageRef.ID)
}

func getGraphClient() (*armresourcegraph.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting the best spot price choice: %v", err)
	}
	// ResourceGraph client
	return armresourcegraph.NewClient(cred, nil)
}

// This function will return a slice of values with price ordered from minor prices to major
func getPriceHistory(ctx context.Context, client *armresourcegraph.Client,
	r BestSpotChoiceRequest) ([]priceHistory, error) {
	data := struct {
		VMTypes []string
		OSType  string
	}{
		VMTypes: r.VMTypes,
		OSType:  r.OSType,
	}
	tmpl, err := template.New("graphQuery").Parse(querySpotPrice)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return nil, err
	}
	spr := buffer.String()
	logging.Debug(spr)

	qr, err := client.Resources(ctx,
		armresourcegraph.QueryRequest{
			Query: to.Ptr(spr),
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error getting spot price history: %v", err)
	}
	var pha []priceHistory
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("error getting spot price history: %v", err)
		}
		rStruct := priceHistory{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			return nil, fmt.Errorf("error getting spot price history: %v", err)
		}
		pha = append(pha, rStruct)
	}
	results := pha
	// Exclude results for excluded regions if any
	if len(r.ExcludedRegions) > 0 {
		results = util.ArrayFilter(pha, func(ph priceHistory) bool {
			return !slices.Contains(r.ExcludedRegions, ph.Location)
		})
	}
	logging.Debugf("spot prices history %v", results)
	return results, nil
}

func getEvictionRateInfoByVMTypes(ctx context.Context, client *armresourcegraph.Client,
	vmTypes, excludedRegions []string) ([]evictionRate, error) {
	data := struct {
		VMTypes []string
	}{
		VMTypes: vmTypes,
	}
	tmpl, err := template.New("graphQuery").Parse(queryEvictionRate)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return nil, err
	}
	evrr := buffer.String()
	logging.Debug(evrr)

	qr, err := client.Resources(ctx,
		armresourcegraph.QueryRequest{
			Query: to.Ptr(evrr),
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error getting eviction rate: %v", err)
	}
	var era []evictionRate
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("error getting eviction rate: %v", err)
		}
		rStruct := evictionRate{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			return nil, fmt.Errorf("error getting eviction rate: %v", err)
		}
		era = append(era, rStruct)
	}
	results := era
	// Exclude results for excluded regions if any
	if len(excludedRegions) > 0 {
		results = util.ArrayFilter(era, func(er evictionRate) bool {
			return !slices.Contains(excludedRegions, er.Location)
		})
	}
	return results, nil
}

func getBestSpotChoice(s []priceHistory, e []evictionRate, currentERT EvictionRate, maxERT EvictionRate, imageID string) (*BestSpotChoiceResponse, error) {
	var evm = make(map[string]string)
	for _, ev := range e {
		evm[fmt.Sprintf("%s%s", ev.Location, ev.VMType)] = ev.EvictionRate
	}
	var spotChoices []*BestSpotChoiceResponse
	for _, sv := range s {
		er, ok := evm[fmt.Sprintf("%s%s", sv.Location, sv.VMType)]
		// If there are multiple choices we added them to a slice
		// and pick one randomly to improve distribution of instances
		// across locations
		if ok && er == getEvictionRateValue(currentERT) {
			ir := ImageRequest{
				Region: sv.Location,
				ImageReference: ImageReference{
					ID: imageID,
				},
			}
			if IsImageOffered(ir) {
				spotChoices = append(spotChoices,
					&BestSpotChoiceResponse{
						VMType:       sv.VMType,
						Location:     sv.Location,
						Price:        sv.Price,
						EvictionRate: er,
					})

			}
		}
	}
	if len(spotChoices) > 0 {
		return util.RandomItemFromArray(spotChoices), nil
	}
	// If current is equal to max tolerance we can not give any spot
	if currentERT == maxERT {
		return nil, fmt.Errorf("could not find any spot with minimum eviction rate")
	}
	// We will run getBestSpotChoice recursively based on ordered list of tolerances
	// when we reach the lowest if no machine is available it will return err
	higherERT, ok := getHigherEvictionRate(currentERT)
	if !ok {
		return nil, fmt.Errorf("could not find any spot")
	}
	return getBestSpotChoice(s, e, *higherERT, maxERT, imageID)
}

// Get previous higher evicition rate for a giving eviction rate
// if there is a higher rate it returns its value and true
// if the current is the highest it returns nil and false
func getHigherEvictionRate(current EvictionRate) (*EvictionRate, bool) {
	var ers []evictionRateSpec
	for _, er := range evictionRates {
		ers = append(ers, er)
	}
	sort.Slice(ers, func(i, j int) bool { return ers[i].order < ers[j].order })
	i := slices.IndexFunc(ers, func(e evictionRateSpec) bool {
		return e.id == current
	})
	if i == 0 {
		return nil, false
	}
	return &ers[i-1].id, true
}

// Translate eviction rate to value
func getEvictionRateValue(er EvictionRate) string {
	var ers []evictionRateSpec
	for _, er := range evictionRates {
		ers = append(ers, er)
	}
	i := slices.IndexFunc(
		ers,
		func(e evictionRateSpec) bool {
			return e.id == er
		})
	return ers[i].value
}

// Get eviction rate parsing its name
func parseEvictionRate(str string) (EvictionRate, bool) {
	c, ok := evictionRates[strings.ToLower(str)]
	return c.id, ok
}

// This is a fallback function in case we need to get an option only based in price
// In order to add some type of distribution across the information we will 1/3 at beguining
// 1/3 at the end and then randomly we will pick one of the remaining
func getSpotChoiceByPrice(s []priceHistory, imageID string) (*BestSpotChoiceResponse, error) {
	var spotChoices []*BestSpotChoiceResponse
	for _, sv := range s {
		ir := ImageRequest{
			Region: sv.Location,
			ImageReference: ImageReference{
				ID: imageID,
			},
		}
		if IsImageOffered(ir) {
			spotChoices = append(spotChoices,
				&BestSpotChoiceResponse{
					VMType:   sv.VMType,
					Location: sv.Location,
					Price:    sv.Price,
				})

		}
	}
	if len(spotChoices) > 3 {
		return util.RandomItemFromArray(
				spotChoices[len(spotChoices)/3 : len(spotChoices)-len(spotChoices)/3]),
			nil
	}
	return util.RandomItemFromArray(spotChoices), nil
}
