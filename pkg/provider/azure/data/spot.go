package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"text/template"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
	utilSlices "github.com/redhat-developer/mapt/pkg/util/slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	hostingPlaces "github.com/redhat-developer/mapt/pkg/provider/util/hosting-places"
)

const (
	querySpotPrice = "SpotResources | where type =~ 'microsoft.compute/skuspotpricehistory/ostype/location' " +
		"and sku.name in~ ({{range $index, $v := .ComputeSizes}}{{if $index}},{{end}}'{{$v}}'{{end}}) and properties.osType =~ '{{.OSType}}'" +
		"and location =~ '{{.Location}}' " +
		"| project skuName=tostring(sku.name),osType=tostring(properties.osType)," +
		"location,latestSpotPriceUSD=todouble(properties.spotPrices[0].priceUSD)" +
		"| order by latestSpotPriceUSD asc"

	queryEvictionRate = "SpotResources | where type =~ 'microsoft.compute/skuspotevictionrate/location' " +
		"and sku.name in~ ({{range $index, $v := .ComputeSizes}}{{if $index}},{{end}}'{{$v}}'{{end}})" +
		"and location =~ '{{.Location}}' " +
		"and tostring(properties.evictionRate) in~ ({{range $index, $e := .AllowedER}}{{if $index}},{{end}}'{{$e}}'{{end}}) " +
		"| project skuName=tostring(sku.name),location,spotEvictionRate=tostring(properties.evictionRate) "
)

type SpotSelector struct{}

func NewSpotSelector() *SpotSelector { return &SpotSelector{} }

func (c *SpotSelector) Select(mCtx *mc.Context,
	args *spot.SpotRequestArgs) (*spot.SpotResults, error) {
	return getSpotInfo(mCtx, args)
}

func getSpotInfo(mCtx *mc.Context, args *spot.SpotRequestArgs) (*spot.SpotResults, error) {
	var err error
	css := args.ComputeRequest.ComputeSizes
	if len(css) == 0 {
		css, err =
			NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	return SpotInfo(mCtx,
		&SpotInfoArgs{
			ComputeSizes:      css,
			OSType:            osType(args.OS),
			ExcludedLocations: args.SpotParams.ExcludedHostingPlaces,
			SpotTolerance:     &args.SpotParams.Tolerance,
		})
}

type SpotInfoArgs struct {
	ComputeSizes          []string
	ImageRef              *ImageReference
	OSType                string
	ExcludedLocations     []string
	SpotTolerance         *spot.Tolerance
	SpotPriceIncreaseRate *int
}

type SpotInfoResult struct {
	ComputeSize  string  `json:"skuName"`
	Location     string  `json:"location"`
	Price        float64 `json:"latestSpotPriceUSD"`
	EvictionRate string  `json:"evictionRate"`
}

// var ErrEvictionRatesEmtpyData = fmt.Errorf("error eviction rates are returning empty")

// This function will return the best spot option
func SpotInfo(mCtx *mc.Context, args *SpotInfoArgs) (*spot.SpotResults, error) {
	if args.SpotTolerance == nil {
		args.SpotTolerance = &spot.DefaultTolerance
	}
	locations, err := filterLocations(mCtx, args)
	if err != nil {
		return nil, err
	}
	clientFactory, err := getGraphClientFactory()
	if err != nil {
		return nil, err
	}
	evictionRates, err := hostingPlaces.RunOnHostingPlaces(locations,
		evictionRatesArgs{
			computeSizes:  args.ComputeSizes,
			clientFactory: clientFactory,
			allowedER:     allowedER(*args.SpotTolerance),
		},
		evictionRatesAsync)
	if err != nil {
		return nil, err
	}
	// If eviction rates info is available only locations with
	// information will be processed for price
	if len(evictionRates) > 0 {
		locations = utilMaps.Keys(evictionRates)
	}
	// prices
	spotPricings, err := hostingPlaces.RunOnHostingPlaces(locations,
		spotPricingArgs{
			computeSizes:  args.ComputeSizes,
			clientFactory: clientFactory,
			osType:        args.OSType,
		},
		spotPricingAsync)
	if err != nil {
		return nil, err
	}
	c, err := selectSpotChoice(
		&spotChoiceArgs{
			evictionRates: evictionRates,
			spotPricings:  spotPricings,
		})
	if err != nil {
		return nil, err
	}
	if c != nil {
		logging.Debugf("Based on avg prices for instance types %v is az %s, current avg price is %.2f and max price is %.2f with a score of %s",
			args.ComputeSizes, c.Location, c.Price, c.Price, c.EvictionRate)
	} else {
		return nil, fmt.Errorf("couldn't find the best price for instance types %v", args.ComputeSizes)
	}
	sr := spot.SpotResults{
		ComputeType:  []string{c.ComputeSize},
		HostingPlace: c.Location,
		Price: spot.SafePrice(c.Price,
			args.SpotPriceIncreaseRate),
		// ChanceLevel: cl,
	}
	logging.Debugf("Spot data: %v", sr)
	return &sr, nil
}

type evictionRateSpec struct {
	spot.Tolerance
	value string
}

// Ordered list of eviction rates
var evictionRates = []evictionRateSpec{
	{spot.Lowest, "0-5"},
	{spot.Low, "5-10"},
	{spot.Medium, "10-15"},
	{spot.High, "15-20"},
	{spot.Highest, "20+"},
}

var evictionRatesToInt = map[string]int{
	"0-5":   0,
	"5-10":  1,
	"10-15": 2,
	"15-20": 3,
	"20+":   4,
}

// filter locations suitable for running mapt targets on spot instances
func filterLocations(mCtx *mc.Context, args *SpotInfoArgs) ([]string, error) {
	// Get all available locations for subscription allowing PublicIPs
	locations, err := LocationsBySupportedResourceType(RTPublicIPAddresses)
	if err != nil {
		return nil, err
	}
	if len(args.ExcludedLocations) > 0 {
		locations = util.ArrayFilter(locations,
			func(location string) bool {
				return !slices.Contains(args.ExcludedLocations, location)
			})
	}
	if args.ImageRef != nil {
		locations = util.ArrayFilter(locations,
			func(location string) bool {
				return IsImageOffered(mCtx,
					ImageRequest{
						Region:         location,
						ImageReference: *args.ImageRef,
					})
			})
	}
	return locations, err
}

func allowedER(spotTolerance spot.Tolerance) []string {
	idx := slices.IndexFunc(evictionRates,
		func(e evictionRateSpec) bool {
			return e.Tolerance == spotTolerance
		})
	return util.ArrayConvert(
		evictionRates[:idx+1],
		func(e evictionRateSpec) string {
			return e.value
		})
}

type evictionRatesArgs struct {
	clientFactory *armresourcegraph.ClientFactory
	computeSizes  []string
	allowedER     []string
	// capacity     int32
}

type evictionRateResult struct {
	ComputeSize  string `json:"skuName"`
	Location     string `json:"location"`
	EvictionRate string `json:"spotEvictionRate"`
}

type queryERData struct {
	ComputeSizes []string
	Location     string
	AllowedER    []string
}

// This will get evictionrates grouped on map per region
// only scores over tolerance will be added
func evictionRatesAsync(location string, args evictionRatesArgs, c chan hostingPlaces.HostingPlaceData[[]evictionRateResult]) {
	tmpl, err := template.New("graphQuery").Parse(queryEvictionRate)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, queryERData{
		ComputeSizes: args.computeSizes,
		Location:     location,
		AllowedER:    args.allowedER,
	})
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	evrr := buffer.String()
	qr, err := args.clientFactory.NewClient().Resources(context.Background(),
		armresourcegraph.QueryRequest{
			Query: to.Ptr(evrr),
		},
		nil)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	var results []evictionRateResult
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			hostingPlaces.SendAsyncErr(c, err)
			return
		}
		rStruct := evictionRateResult{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			hostingPlaces.SendAsyncErr(c, err)
			return
		}
		results = append(results, rStruct)
	}
	// Order by eviction rate
	slices.SortFunc(results,
		func(a, b evictionRateResult) int {
			return int(
				evictionRatesToInt[a.EvictionRate] - evictionRatesToInt[b.EvictionRate])
		})
	c <- hostingPlaces.HostingPlaceData[[]evictionRateResult]{
		Region: location,
		Value:  results}
}

type spotPricingArgs struct {
	clientFactory *armresourcegraph.ClientFactory
	computeSizes  []string
	osType        string
	// capacity     int32
}

type spotPricingResult struct {
	ComputeSize string  `json:"skuName"`
	OSType      string  `json:"osType"`
	Location    string  `json:"location"`
	Price       float64 `json:"latestSpotPriceUSD"`
}

type querySpotPriceData struct {
	ComputeSizes []string
	Location     string
	OSType       string
}

// This function will return a slice of values with price ordered from minor prices to major
func spotPricingAsync(location string, args spotPricingArgs, c chan hostingPlaces.HostingPlaceData[[]spotPricingResult]) {
	tmpl, err := template.New("graphQuery").Parse(querySpotPrice)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, querySpotPriceData{
		ComputeSizes: args.computeSizes,
		Location:     location,
		OSType:       args.osType,
	})
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	spr := buffer.String()
	qr, err := args.clientFactory.NewClient().Resources(context.Background(),
		armresourcegraph.QueryRequest{
			Query: to.Ptr(spr),
		},
		nil)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	var results []spotPricingResult
	for _, r := range qr.Data.([]interface{}) {
		rJSON, err := json.Marshal(r)
		if err != nil {
			hostingPlaces.SendAsyncErr(c, err)
			return
		}
		rStruct := spotPricingResult{}
		if err := json.Unmarshal(rJSON, &rStruct); err != nil {
			hostingPlaces.SendAsyncErr(c, err)
			return
		}
		logging.Debugf("Found ComputeSize %s at Location %s with spot price %.2f",
			string(rStruct.ComputeSize), rStruct.Location, rStruct.Price)
		results = append(results, rStruct)
	}
	// Order by price
	if len(results) > 0 {
		utilSlices.SortbyFloat(results,
			func(s spotPricingResult) float64 {
				return s.Price
			})
	}
	c <- hostingPlaces.HostingPlaceData[[]spotPricingResult]{
		Region: location,
		Value:  results}

}

type spotChoiceArgs struct {
	evictionRates map[string][]evictionRateResult
	spotPricings  map[string][]spotPricingResult
}

// checkBestOption will cross data from prices (starting at lower prices)
// and cross that information with regions with best scores and also ensuring
// ami is offered on that specific region
//
// # Also function take cares to transfrom from AzID to AZName
//
// first option matching the requirements will be returned
func selectSpotChoice(args *spotChoiceArgs) (*SpotInfoResult, error) {
	result := make(map[string]*SpotInfoResult)
	// Fix random error with graphql query not giving information for eviction rates
	if len(args.evictionRates) == 0 {
		return spotOnlyByPrices(args.spotPricings)
	}
	// This can bexecuted async
	for l, pss := range args.spotPricings {
		for _, ps := range pss {
			idx := slices.IndexFunc(args.evictionRates[l],
				func(evr evictionRateResult) bool {
					return evr.Location == ps.Location
				})
			if idx != -1 {
				result[l] = &SpotInfoResult{
					Location:     args.evictionRates[l][idx].Location,
					Price:        ps.Price,
					ComputeSize:  ps.ComputeSize,
					EvictionRate: args.evictionRates[l][idx].EvictionRate,
				}
			}
		}
	}
	spis := utilMaps.Values(result)
	utilSlices.SortbyFloat(spis,
		func(s *SpotInfoResult) float64 {
			return s.Price
		})
	if len(spis) == 0 {
		return nil, fmt.Errorf("no good choice was found")
	}
	return spis[0], nil
}

// // This is a fallback function in case we need to get an option only based in price
// // In order to add some type of distribution across the information we will 1/3 at beguining
// // 1/3 at the end and then randomly we will pick one of the remaining
func spotOnlyByPrices(s map[string][]spotPricingResult) (*SpotInfoResult, error) {
	var bsp []*SpotInfoResult
	for location, prices := range s {
		bsp = append(bsp,
			&SpotInfoResult{
				ComputeSize: prices[0].ComputeSize,
				Location:    location,
				Price:       prices[0].Price,
			})
	}
	return util.RandomItemFromArray(bsp), nil
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
