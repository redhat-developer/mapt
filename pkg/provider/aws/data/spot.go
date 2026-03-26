package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	hostingPlaces "github.com/redhat-developer/mapt/pkg/provider/util/hosting-places"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
	utilSlices "github.com/redhat-developer/mapt/pkg/util/slices"
	"golang.org/x/exp/slices"
)

var (
	arm64  = "aarch64"
	x86_64 = "x86_64"

	defaultOS = "linux"

	amiProductWindows = "Windows"
	amiProductRHEL    = "Red Hat Enterprise Linux"
	amiProductLinux   = "Linux/UNIX"
	amiProducts       = map[string]*string{
		"windows": &amiProductWindows,
		"RHEL":    &amiProductRHEL,
		"fedora":  &amiProductLinux,
		defaultOS: &amiProductLinux,
	}
)

type placementScoreSpec struct {
	spot.Tolerance
	minPlacementScore int32
}

// Ordered list of placement scores rates
var placementScores = []placementScoreSpec{
	{spot.Lowest, 7},
	{spot.Low, 5},
	{spot.Medium, 3},
	{spot.High, 2},
	{spot.Highest, 1},
}

// While calculating the spot price and the types of machines
// to be requested we use the first n number of types per Az
// to control how many types will be used this var is used
var maxNumberOfTypesForSpot = 8

func minPlacementScore(spotTolerance spot.Tolerance) int32 {
	idx := slices.IndexFunc(placementScores,
		func(e placementScoreSpec) bool {
			return e.Tolerance == spotTolerance
		})
	return placementScores[idx].minPlacementScore
}

type SpotSelector struct{}

func NewSpotSelector() *SpotSelector { return &SpotSelector{} }

func (c *SpotSelector) Select(mCtx *mc.Context, args *spot.SpotRequestArgs) (*spot.SpotResults, error) {
	return getSpotInfo(mCtx, args)
}

func getSpotInfo(mCtx *mc.Context, args *spot.SpotRequestArgs) (*spot.SpotResults, error) {
	var err error
	computeTypes := args.ComputeRequest.ComputeSizes
	if len(computeTypes) == 0 {
		computeTypes, err =
			NewComputeSelector().Select(mCtx.Context(), args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	siArgs := &SpotInfoArgs{
		InstaceTypes:       computeTypes,
		AMIName:            args.ImageName,
		ProductDescription: amiProducts[defaultOS],
	}
	if args.ComputeRequest != nil {
		siArgs.AMIArch = util.If(
			args.ComputeRequest.Arch == cr.Arm64,
			&arm64,
			&x86_64)
	}
	if args.OS != nil {
		siArgs.ProductDescription = amiProducts[*args.OS]
	}
	return SpotInfo(mCtx, siArgs)
}

const (
	// Page size for the batched GetSpotPlacementScores call (API max is 1000)
	maxPageSizePlacementScore = 1000
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription = "product-description"
)

type SpotInfoArgs struct {
	InstaceTypes []string
	// AMI information
	ProductDescription *string
	AMIName, AMIArch   *string

	ExcludedRegions       []string
	SpotTolerance         *spot.Tolerance
	SpotPriceIncreaseRate *int
}

type SpotInfoResult struct {
	Region           string
	AvailabilityZone string
	Price            float64
	Score            int32
	// In aws we need at least 3 types
	InstanceType []string
}

// filter regions suitable for running mapt targets on spot instances
func filterRegions(mCtx *mc.Context, args *SpotInfoArgs) ([]string, error) {
	regions, err := GetRegions(mCtx.Context())
	if err != nil {
		return nil, err
	}
	if len(args.ExcludedRegions) > 0 {
		regions = util.ArrayFilter(regions,
			func(region string) bool {
				return !slices.Contains(args.ExcludedRegions, region)
			})
	}
	if args.AMIName != nil && len(*args.AMIName) > 0 {
		regions = util.ArrayFilter(regions,
			func(region string) bool {
				validAMI, _, err := IsAMIOffered(
					mCtx.Context(),
					ImageRequest{
						Name:   args.AMIName,
						Arch:   args.AMIArch,
						Region: &region})
				if err != nil {
					if mCtx.Debug() {
						logging.Warn(err.Error())
					}
				}
				if !validAMI {
					logging.Debugf("AMI %s is not available in region %s", *args.AMIName, region)
				}
				return validAMI
			})
	}
	return regions, err
}

// This function checks worlwide which is the best place at any point in time to spin a spot machine
// it basically cross the information for spot prices and placement scores
// the target machine is defined through the inputs for the funtion:
// * productType to be executed within the machine
// * instanceTypes types of machines able to execute the workload
// * amiName ensures the ami is available on the spot option
// the output is the information realted to the best spot option for the target machine
func SpotInfo(mCtx *mc.Context, args *SpotInfoArgs) (*spot.SpotResults, error) {
	if args.SpotTolerance == nil {
		args.SpotTolerance = &spot.DefaultTolerance
	}
	regions, err := filterRegions(mCtx, args)
	if err != nil {
		return nil, err
	}
	// Placement Scores — one batched API call for all regions to avoid
	// exceeding the account quota on concurrent placement configurations.
	nOpInRegions, err := GetRegionsByOptInStatus(mCtx.Context(), []string{OptInStatusNotRequired})
	if err != nil {
		return nil, err
	}
	placementScores, err := getPlacementScores(
		placementScoreArgs{
			ctx:               mCtx.Context(),
			apiRegions:        util.RandomizeArrayContent(nOpInRegions),
			minPlacementScore: minPlacementScore(*args.SpotTolerance),
			instanceTypes:     args.InstaceTypes,
			capacity:          1,
		},
		regions)
	if err != nil {
		return nil, err
	}
	regionsWithPlacementScore := utilMaps.Keys(placementScores)
	spotPricing, err := hostingPlaces.RunOnHostingPlaces(regionsWithPlacementScore,
		spotPricingArgs{
			ctx:                mCtx.Context(),
			mCtx:               mCtx,
			productDescription: *args.ProductDescription,
			instanceTypes:      args.InstaceTypes,
		},
		spotPricingAsync)
	if err != nil {
		return nil, err
	}
	numberOfTypesForSpot := util.If(
		len(args.InstaceTypes) < maxNumberOfTypesForSpot,
		len(args.InstaceTypes),
		maxNumberOfTypesForSpot)
	c, err := selectSpotChoice(
		&spotChoiceArgs{
			placementScores: placementScores,
			spotPricing:     spotPricing,
		},
		numberOfTypesForSpot)
	if err != nil {
		return nil, err
	}
	if c != nil {
		logging.Debugf("Based on prices for instance types %v is az %s, current price is %.2f with a score of %d",
			c.InstanceType, c.AvailabilityZone, c.Price, c.Score)
	} else {
		return nil, fmt.Errorf("couldn't find the best price for instance types %v", args.InstaceTypes)
	}
	// TODO
	// translate Score
	sr := spot.SpotResults{
		ComputeType: c.InstanceType,
		Price: spot.SafePrice(c.Price,
			args.SpotPriceIncreaseRate),
		HostingPlace:     c.Region,
		AvailabilityZone: c.AvailabilityZone,
		// ChanceLevel:      int(sp.Score),
	}
	logging.Debugf("Spot data: %v", sr)
	return &sr, nil
}

type spotChoiceArgs struct {
	placementScores map[string][]placementScoreResult
	spotPricing     map[string][]spotPrincingResults
}

// checkBestOption will cross data from prices (starting at lower prices)
// and cross that information with regions with best scores and also ensuring
// ami is offered on that specific region
//
// # Also function take cares to transfrom from AzID to AZName
//
// first option matching the requirements will be returned
func selectSpotChoice(args *spotChoiceArgs, numberOfTypesForSpot int) (*SpotInfoResult, error) {
	result := make(map[string]*SpotInfoResult)
	// This can bexecuted async
	for r, pss := range args.spotPricing {
		resultByAZ := make(map[string][]*SpotInfoResult)
		for _, ps := range pss {
			idx := slices.IndexFunc(args.placementScores[r],
				func(psr placementScoreResult) bool {
					return psr.azName == ps.AvailabilityZone
				})
			if idx != -1 {
				resultByAZ[ps.AvailabilityZone] = append(
					resultByAZ[ps.AvailabilityZone], &SpotInfoResult{
						Region:           *args.placementScores[r][idx].sps.Region,
						AvailabilityZone: ps.AvailabilityZone,
						Price:            ps.Price,
						Score:            *args.placementScores[r][idx].sps.Score,
						InstanceType:     []string{ps.InstanceType},
					})
			}
			if len(resultByAZ[ps.AvailabilityZone]) == numberOfTypesForSpot {
				result[r] = aggregateSpotChoice(resultByAZ[ps.AvailabilityZone], numberOfTypesForSpot)
				break
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
	logging.Debugf("Sorted %d spot options by price", len(spis))
	for i, spi := range spis {
		logging.Debugf("  Option %d: Region %s, AZ %s, Price $%.4f, Score %d, Instance types: %v",
			i+1, spi.Region, spi.AvailabilityZone, spi.Price, spi.Score, spi.InstanceType)
	}
	logging.Debugf("Selected cheapest option - Region %s, AZ %s, Price $%.4f, Score %d, Instance types: %v",
		spis[0].Region, spis[0].AvailabilityZone, spis[0].Price, spis[0].Score, spis[0].InstanceType)
	return spis[0], nil
}

// previously we pick 3 values per Az we aggregate its data
func aggregateSpotChoice(s []*SpotInfoResult, numberOfTypesForSpot int) *SpotInfoResult {
	return &SpotInfoResult{
		Region:           s[0].Region,
		AvailabilityZone: s[0].AvailabilityZone,
		Price:            s[numberOfTypesForSpot-1].Price,
		Score:            s[numberOfTypesForSpot-1].Score,
		InstanceType: util.ArrayConvert(s[:numberOfTypesForSpot],
			func(s *SpotInfoResult) string {
				return s.InstanceType[0]
			}),
	}
}

type spotPricingArgs struct {
	ctx                context.Context
	mCtx               *mc.Context
	productDescription string
	instanceTypes      []string
}

type spotPrincingResults struct {
	Region           string
	AvailabilityZone string
	Price            float64
	Score            int32
	InstanceType     string
}

func spotPricingAsync(r string, args spotPricingArgs, c chan hostingPlaces.HostingPlaceData[[]spotPrincingResults]) {
	cfg, err := getConfig(args.ctx, r)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	client := ec2.NewFromConfig(cfg)
	starTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	history, err := client.DescribeSpotPriceHistory(
		args.ctx,
		&ec2.DescribeSpotPriceHistoryInput{
			InstanceTypes: util.ArrayConvert(args.instanceTypes,
				func(i string) ec2Types.InstanceType {
					return ec2Types.InstanceType(i)
				}),
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String(spotQueryFilterProductDescription),
					Values: []string{args.productDescription},
				},
			},
			StartTime: &starTime,
			EndTime:   &endTime,
		})
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	spotPriceGroups := utilSlices.Split(
		history.SpotPriceHistory,
		func(priceData ec2Types.SpotPrice) spotPrincingResults {
			return spotPrincingResults{
				AvailabilityZone: *priceData.AvailabilityZone,
				InstanceType:     string(priceData.InstanceType),
			}
		})
	if args.mCtx.Debug() {
		for _, v := range spotPriceGroups {
			for _, sp := range v {
				logging.Debugf("Found InstanceType %s at Availability Zone %s with spot price %s", string(sp.InstanceType), *sp.AvailabilityZone, *sp.SpotPrice)
			}
		}
	}
	var results []spotPrincingResults
	for groupInfo, pricesHistory := range spotPriceGroups {
		prices := util.ArrayConvert(pricesHistory, func(priceHisotry ec2Types.SpotPrice) float64 {
			price, err := strconv.ParseFloat(*priceHisotry.SpotPrice, 64)
			if err != nil {
				// Overcost
				return 100
			}
			return price
		})
		groupInfo.Price = prices[len(prices)-1]
		groupInfo.Region = r
		results = append(results, groupInfo)
	}
	utilSlices.SortbyFloat(results,
		func(s spotPrincingResults) float64 {
			return s.Price
		})
	c <- hostingPlaces.HostingPlaceData[[]spotPrincingResults]{
		Region: r,
		Value:  results,
		Err:    err}
}

type placementScoreArgs struct {
	ctx context.Context
	// Not all regions offer the GetSpotPlacementScores API.
	// Regions are tried in order; the first successful one is used.
	// Pass a shuffled list to distribute load and recover from unsupported regions.
	apiRegions        []string
	minPlacementScore int32
	instanceTypes     []string
	capacity          int32
}

type placementScoreResult struct {
	sps ec2Types.SpotPlacementScore
	// We need to match between AzId offered and the naming in our account
	azName string
}

// getPlacementScores makes a single paginated GetSpotPlacementScores call for all
// regions instead of N concurrent per-region calls. Concurrent calls were hitting
// the account quota on simultaneous placement configurations.
// apiRegions is tried in order (pass a shuffled list to distribute load); the first
// region that responds successfully is used. Regions that don't support the API are
// skipped. Returns a map of region → AZ scores filtered to those meeting minPlacementScore.
func getPlacementScores(args placementScoreArgs, regions []string) (map[string][]placementScoreResult, error) {
	azsByRegion := describeAvailabilityZonesByRegions(args.ctx, regions)
	var lastErr error
	for _, apiRegion := range args.apiRegions {
		result, err := placementScoresViaRegion(apiRegion, args, regions, azsByRegion)
		if err != nil {
			logging.Debugf("placement score API unavailable in region %s: %v, trying next", apiRegion, err)
			lastErr = err
			continue
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("no placement scores above minimum threshold found across regions")
		}
		for r := range result {
			slices.SortFunc(result[r], func(a, b placementScoreResult) int {
				return int(*b.sps.Score - *a.sps.Score)
			})
		}
		return result, nil
	}
	return nil, fmt.Errorf("placement score API failed across all candidate regions: %w", lastErr)
}

// placementScoresViaRegion calls GetSpotPlacementScores using apiRegion as the endpoint
// and returns unsorted results grouped by region. Returns error only on API failure,
// not when results are empty (empty means no AZ met the score threshold).
func placementScoresViaRegion(apiRegion string, args placementScoreArgs, regions []string, azsByRegion map[string][]ec2Types.AvailabilityZone) (map[string][]placementScoreResult, error) {
	cfg, err := getConfig(args.ctx, apiRegion)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	result := make(map[string][]placementScoreResult)
	var nextToken *string
	for {
		sps, err := client.GetSpotPlacementScores(
			args.ctx,
			&ec2.GetSpotPlacementScoresInput{
				SingleAvailabilityZone: aws.Bool(true),
				InstanceTypes:          args.instanceTypes,
				RegionNames:            regions,
				TargetCapacity:         aws.Int32(args.capacity),
				MaxResults:             aws.Int32(maxPageSizePlacementScore),
				NextToken:              nextToken,
			})
		if err != nil {
			return nil, err
		}
		for _, ps := range sps.SpotPlacementScores {
			if *ps.Score >= args.minPlacementScore {
				azName, err := getZoneName(*ps.AvailabilityZoneId, azsByRegion[*ps.Region])
				if err != nil {
					// AZ may not be visible to this account; skip rather than abort
					logging.Debugf("skipping AZ %s in region %s: %v", *ps.AvailabilityZoneId, *ps.Region, err)
					continue
				}
				result[*ps.Region] = append(result[*ps.Region], placementScoreResult{
					sps:    ps,
					azName: azName,
				})
			} else {
				logging.Debugf("Availability zone %s in region %s filtered out (score %d < minimum %d)",
					*ps.AvailabilityZoneId, *ps.Region, *ps.Score, args.minPlacementScore)
			}
		}
		if sps.NextToken == nil {
			break
		}
		nextToken = sps.NextToken
	}
	return result, nil
}
