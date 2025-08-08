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

	// TODO we need to match tolerance from spot with this value
	// 1-4
	// 4-6
	// 6-8
	// 8-10
	tolerance int32 = 3
)

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
			NewComputeSelector().Select(args.ComputeRequest)
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
	// Max number of results for placement score query
	maxQueryResultsResultsPlacementScore = 10
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
	regions, err := GetRegions()
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
					ImageRequest{
						Name:   args.AMIName,
						Arch:   args.AMIArch,
						Region: &region})
				if err != nil {
					if mCtx.Debug() {
						logging.Warn(err.Error())
					}
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

	placementScores, err := hostingPlaces.RunOnHostingPlaces(regions,
		placementScoreArgs{
			instanceTypes: args.InstaceTypes,
			capacity:      1,
		},
		placementScoresAsync)
	if err != nil {
		return nil, err
	}
	regionsWithPlacementScore := utilMaps.Keys(placementScores)
	spotPricing, err := hostingPlaces.RunOnHostingPlaces(regionsWithPlacementScore,
		spotPricingArgs{
			productDescription: *args.ProductDescription,
			instanceTypes:      args.InstaceTypes,
		},
		spotPricingAsync)
	if err != nil {
		return nil, err
	}
	c, err := selectSpotChoice(
		&spotChoiceArgs{
			placementScores: placementScores,
			spotPricing:     spotPricing,
		})
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
func selectSpotChoice(args *spotChoiceArgs) (*SpotInfoResult, error) {
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
			if len(resultByAZ[ps.AvailabilityZone]) == 5 {
				result[r] = aggregateSpotChoice(resultByAZ[ps.AvailabilityZone])
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
	return spis[0], nil
}

// previously we pick 3 values per Az we aggregate its data
func aggregateSpotChoice(s []*SpotInfoResult) *SpotInfoResult {
	return &SpotInfoResult{
		Region:           s[0].Region,
		AvailabilityZone: s[0].AvailabilityZone,
		Price:            s[4].Price,
		Score:            s[4].Score,
		InstanceType: []string{
			s[0].InstanceType[0],
			s[1].InstanceType[0],
			s[2].InstanceType[0],
			s[3].InstanceType[0],
			s[4].InstanceType[0]},
	}
}

type spotPricingArgs struct {
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
	cfg, err := getConfig(r)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	client := ec2.NewFromConfig(cfg)
	starTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	history, err := client.DescribeSpotPriceHistory(
		context.Background(),
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

	for _, v := range spotPriceGroups {
		for _, sp := range v {
			logging.Debugf("Found InstanceType %s at Availability Zone %s with spot price %s", string(sp.InstanceType), *sp.AvailabilityZone, *sp.SpotPrice)
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
	c <- hostingPlaces.HostingPlaceData[[]spotPrincingResults]{
		Region: r,
		Value:  results,
		Err:    err}
}

type placementScoreArgs struct {
	instanceTypes []string
	capacity      int32
}

type placementScoreResult struct {
	sps ec2Types.SpotPlacementScore
	// We need to match between AzId offered and the naming in our account
	azName string
}

// This will get placement scores grouped on map per region
// only scores over tolerance will be added
func placementScoresAsync(r string, args placementScoreArgs, c chan hostingPlaces.HostingPlaceData[[]placementScoreResult]) {
	azsByRegion := describeAvailabilityZonesByRegions([]string{r})
	cfg, err := getConfig(r)
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	client := ec2.NewFromConfig(cfg)
	sps, err := client.GetSpotPlacementScores(
		context.Background(),
		&ec2.GetSpotPlacementScoresInput{
			SingleAvailabilityZone: aws.Bool(true),
			InstanceTypes:          args.instanceTypes,
			RegionNames:            []string{r},
			TargetCapacity:         aws.Int32(args.capacity),
			MaxResults:             aws.Int32(maxQueryResultsResultsPlacementScore),
		})
	if err != nil {
		hostingPlaces.SendAsyncErr(c, err)
		return
	}
	if len(sps.SpotPlacementScores) == 0 {
		hostingPlaces.SendAsyncErr(c, fmt.Errorf("non available scores"))
		return
	}
	var results []placementScoreResult
	for _, ps := range sps.SpotPlacementScores {
		if *ps.Score >= tolerance {
			azName, err := getZoneName(*ps.AvailabilityZoneId, azsByRegion[*ps.Region])
			if err != nil {
				hostingPlaces.SendAsyncErr(c, err)
				return
			}
			results = append(results, placementScoreResult{
				sps:    ps,
				azName: azName,
			})
		}
	}
	slices.SortFunc(results,
		func(a, b placementScoreResult) int {
			return int(*a.sps.Score - *b.sps.Score)
		})
	c <- hostingPlaces.HostingPlaceData[[]placementScoreResult]{
		Region: r,
		Value:  results}
}
