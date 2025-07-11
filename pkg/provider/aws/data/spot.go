package data

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
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

func (c *SpotSelector) Select(
	args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
	return getSpotInfo(args)
}

func getSpotInfo(args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
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
		AMIName:            args.AMIName,
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
	sp, err := SpotInfo(siArgs)
	if err != nil {
		return nil, err
	}
	return &spotTypes.SpotResults{
		ComputeType:      sp.InstanceType,
		Price:            float32(sp.AVGPrice),
		Region:           sp.Region,
		AvailabilityZone: sp.AvailabilityZone,
		ChanceLevel:      int(sp.Score),
	}, nil
}

const (
	// Max number of results for placement score query
	maxQueryResultsResultsPlacementScore = 10
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription = "product-description"
)

type SpotInfoArgs struct {
	ProductDescription *string
	InstaceTypes       []string
	AMIName, AMIArch   *string
}

type SpotInfoResult struct {
	Region           string
	AvailabilityZone string
	AVGPrice         float64
	MaxPrice         float64
	Score            int32
	InstanceType     string
}

// This function checks worlwide which is the best place at any point in time to spin a spot machine
// it basically cross the information for spot prices and placement scores
// the target machine is defined through the inputs for the funtion:
// * productType to be executed within the machine
// * instanceTypes types of machines able to execute the workload
// * amiName ensures the ami is available on the spot option
// the output is the information realted to the best spot option for the target machine
func SpotInfo(args *SpotInfoArgs) (*SpotInfoResult, error) {
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	placementScores := runByRegion(regions,
		placementScoreArgs{
			instanceTypes: args.InstaceTypes,
			capacity:      1,
		},
		placementScoresAsync)
	regionsWithPlacementScore := utilMaps.Keys(placementScores)
	spotPricing := runByRegion(regionsWithPlacementScore,
		spotPricingArgs{
			productDescription: *args.ProductDescription,
			instanceTypes:      args.InstaceTypes,
		},
		spotPricingAsync)
	c, err := selectSpotChoice(
		&spotChoiceArgs{
			placementScores: placementScores,
			spotPricing:     spotPricing,
			amiName:         args.AMIName,
			amiArch:         args.AMIArch,
		})
	if err != nil {
		return nil, err
	}
	if c != nil {
		logging.Debugf("Based on avg prices for instance types %v is az %s, current avg price is %.2f and max price is %.2f with a score of %d",
			args.InstaceTypes, c.AvailabilityZone, c.AVGPrice, c.MaxPrice, c.Score)
	} else {
		return nil, fmt.Errorf("couldn't find the best price for instance types %v", args.InstaceTypes)
	}
	return c, nil
}

type spotChoiceArgs struct {
	placementScores map[string][]placementScoreResult
	spotPricing     map[string][]spotPrincingResults
	amiName         *string
	amiArch         *string
}

// checkBestOption will cross data from prices (starting at lower prices)
// and cross that information with regions with best scores and also ensuring
// ami is offered on that specific region
//
// # Also function take cares to transfrom from AzID to AZName
//
// first option matching the requirements will be returned
func selectSpotChoice(args *spotChoiceArgs) (*SpotInfoResult, error) {
	var err error
	result := make(map[string]*SpotInfoResult)
	// This can bexecuted async
	for r, pss := range args.spotPricing {
		validAMI := true
		if args.amiName != nil && len(*args.amiName) > 0 {
			validAMI, _, err = IsAMIOffered(
				ImageRequest{
					Name:   args.amiName,
					Arch:   args.amiArch,
					Region: &r})
			if err != nil {
				return nil, err
			}
		}
		for _, ps := range pss {
			idx := slices.IndexFunc(args.placementScores[r],
				func(psr placementScoreResult) bool {
					return psr.azName == ps.AvailabilityZone && validAMI
				})
			if idx != -1 {
				result[r] = &SpotInfoResult{
					Region:           *args.placementScores[r][idx].sps.Region,
					AvailabilityZone: ps.AvailabilityZone,
					AVGPrice:         ps.AVGPrice,
					MaxPrice:         ps.MaxPrice,
					Score:            *args.placementScores[r][idx].sps.Score,
					InstanceType:     ps.InstanceType,
				}
			}
		}
	}
	spis := utilMaps.Values(result)
	utilSlices.SortbyFloat(spis,
		func(s *SpotInfoResult) float64 {
			return s.MaxPrice
		})
	if len(spis) == 0 {
		return nil, fmt.Errorf("no good choice was found")
	}
	return spis[0], nil
}

// Struct to communicate data tied region
// when running some aggregation data func async on a number of regions
type regionData[Y any] struct {
	Region string
	Err    error
	Value  Y
}

// Generic function to run specific function on each region
// and then aggregate the results into a struct
func runByRegion[X, Y any](regions []string, data X,
	run func(string, X, chan regionData[Y])) map[string]Y {
	result := make(map[string]Y)
	c := make(chan regionData[Y], len(regions))
	var wg sync.WaitGroup
	for _, r := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			run(region, data, c)
		}(r)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	for rr := range c {
		if rr.Err == nil {
			result[rr.Region] = rr.Value
		}
	}
	return result
}

type spotPricingArgs struct {
	productDescription string
	instanceTypes      []string
}

type spotPrincingResults struct {
	Region           string
	AvailabilityZone string
	AVGPrice         float64
	MaxPrice         float64
	Score            int32
	InstanceType     string
}

func spotPricingAsync(r string, args spotPricingArgs, c chan regionData[[]spotPrincingResults]) {
	cfg, err := getConfig(r)
	if err != nil {
		c <- regionData[[]spotPrincingResults]{
			Err: err}
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
		c <- regionData[[]spotPrincingResults]{
			Err: err}
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
		groupInfo.AVGPrice = util.Average(prices)
		if len(prices) > 0 {
			utilSlices.SortbyFloat(prices,
				func(s float64) float64 {
					return s
				})
		}
		groupInfo.MaxPrice = prices[len(prices)-1]
		groupInfo.Region = r
		results = append(results, groupInfo)
	}
	c <- regionData[[]spotPrincingResults]{
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
func placementScoresAsync(r string, args placementScoreArgs, c chan regionData[[]placementScoreResult]) {
	azsByRegion := describeAvailabilityZonesByRegions([]string{r})
	cfg, err := getConfig(r)
	if err != nil {
		c <- regionData[[]placementScoreResult]{
			Err: err}
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
		c <- regionData[[]placementScoreResult]{
			Err: err}
		return
	}
	if len(sps.SpotPlacementScores) == 0 {
		c <- regionData[[]placementScoreResult]{
			Err: fmt.Errorf("non available scores")}
		return
	}
	var results []placementScoreResult
	for _, ps := range sps.SpotPlacementScores {
		if *ps.Score >= tolerance {
			azName, err := getZoneName(*ps.AvailabilityZoneId, azsByRegion[*ps.Region])
			if err != nil {
				c <- regionData[[]placementScoreResult]{
					Err: err}
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
	c <- regionData[[]placementScoreResult]{
		Region: r,
		Value:  results}
}
