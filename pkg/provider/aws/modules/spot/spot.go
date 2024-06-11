package spot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/exp/slices"
)

const (
	maxSpotPlacementScoreResults = 10
	spsMaxScore                  = 10
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription = "product-description"
)

type bestSpotOption struct {
	pulumi.ResourceState
	Option *spotOptionInfo
}

func NewBestSpotOption(ctx *pulumi.Context, name string,
	productDescription string, instaceTypes []string,
	amiName, amiArch string, opts ...pulumi.ResourceOption) (*spotOptionInfo, error) {
	spotOption, err := bestSpotOptionInfo(productDescription, instaceTypes, amiName, amiArch)
	if err != nil {
		return nil, err
	}
	err = ctx.RegisterComponentResource("rh:qe:aws:bso",
		name,
		&bestSpotOption{
			Option: spotOption,
		},
		opts...)
	if err != nil {
		return nil, err
	}
	return spotOption, nil
}

// func GetBestSpotOption(ctx *pulumi.Context, name string, id pulumi.IDInput, opts ...pulumi.ResourceOption) (*spotOptionInfo, error) {
// 	var bso bestSpotOption
// 	err := ctx.ReadResource("bso", name, id, nil, &bso, opts...)
// 	return bso.Option, err
// }

type spotOptionInfo struct {
	Region           string
	AvailabilityZone string
	AVGPrice         float64
	MaxPrice         float64
	Score            int32
}

type spotOptionResult struct {
	Prices []spotOptionInfo
	Err    error
}

// This function checks worlwide which is the best place at any point in time to spin a spot machine
// it basically cross the information for spot prices and placement scores
// the target machine is defined through the inputs for the funtion:
// * productType to be executed within the machine
// * instanceTypes types of machines able to execute the workload
// * amiName ensures the ami is available on the spot option
// the output is the information realted to the best spot option for the target machine
func bestSpotOptionInfo(productDescription string, instaceTypes []string, amiName, amiArch string) (*spotOptionInfo, error) {
	regions, err := data.GetRegions()
	if err != nil {
		return nil, err
	}
	ps, err := placementScores(regions, instaceTypes, 1)
	if err != nil {
		return nil, err
	}
	wPrices := spotOptionWorlwide(productDescription, regions, instaceTypes)
	// check this
	bestPrice := checkBestOption(amiName, amiArch, wPrices, ps, describeAvailabilityZones(regions))
	if bestPrice != nil {
		logging.Debugf("Based on avg prices for instance types %v is az %s, current avg price is %.2f and max price is %.2f with a score of %d",
			instaceTypes, bestPrice.AvailabilityZone, bestPrice.AVGPrice, bestPrice.MaxPrice, bestPrice.Score)
	}
	return bestPrice, nil
}

// This function returns placement scores for the instances types across all the target regions
// the list is ordered from max to min by score
func placementScores(regions, instanceTypes []string,
	capacity int32) ([]ec2Types.SpotPlacementScore, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	sps, err := client.GetSpotPlacementScores(
		context.Background(),
		&ec2.GetSpotPlacementScoresInput{
			SingleAvailabilityZone: aws.Bool(true),
			InstanceTypes:          instanceTypes,
			RegionNames:            regions,
			TargetCapacity:         aws.Int32(capacity),
			MaxResults:             aws.Int32(maxSpotPlacementScoreResults),
		})
	if err != nil {
		return nil, err
	}
	if len(sps.SpotPlacementScores) == 0 {
		return nil, fmt.Errorf("non available scores")
	}
	slices.SortFunc(sps.SpotPlacementScores,
		func(a, b ec2Types.SpotPlacementScore) int {
			return int(*a.Score - *b.Score)
		})
	return sps.SpotPlacementScores, nil
}

func spotOptionWorlwide(productDescription string,
	regions, instanceTypes []string) []spotOptionInfo {
	worldwidePrices := []spotOptionInfo{}
	c := make(chan spotOptionResult)
	for _, region := range regions {
		var lRegion = region
		go spotOptionAsync(
			instanceTypes,
			productDescription,
			lRegion,
			c)
	}
	for i := 0; i < len(regions); i++ {
		spotPriceResult := <-c
		if spotPriceResult.Err == nil {
			worldwidePrices = append(worldwidePrices, spotPriceResult.Prices...)
		}
	}
	close(c)
	return worldwidePrices
}

func spotOptionAsync(instanceTypes []string, productDescription, region string, c chan spotOptionResult) {
	data, err := spotOption(instanceTypes, productDescription, region)
	c <- spotOptionResult{
		Prices: data,
		Err:    err}
}

func spotOption(instanceTypes []string,
	productDescription, region string) (
	pricesGroup []spotOptionInfo, err error) {
	var cfgOpts config.LoadOptionsFunc
	if len(region) > 0 {
		cfgOpts = config.WithRegion(region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	starTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	history, err := client.DescribeSpotPriceHistory(
		context.Background(),
		&ec2.DescribeSpotPriceHistoryInput{
			InstanceTypes: util.ArrayConvert(instanceTypes,
				func(i string) ec2Types.InstanceType {
					return ec2Types.InstanceType(i)
				}),
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String(spotQueryFilterProductDescription),
					Values: []string{productDescription},
				},
			},
			StartTime: &starTime,
			EndTime:   &endTime,
		})
	if err != nil {
		return nil, err
	}
	spotPriceGroups := util.SplitSlice(history.SpotPriceHistory, func(priceData ec2Types.SpotPrice) spotOptionInfo {
		return spotOptionInfo{
			AvailabilityZone: *priceData.AvailabilityZone,
		}
	})
	logging.Debugf("grouped prices %v", spotPriceGroups)
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
		slices.SortFunc(prices, func(a, b float64) int { return int(a - b) })
		groupInfo.MaxPrice = prices[len(prices)-1]
		groupInfo.Region = region
		pricesGroup = append(pricesGroup, groupInfo)
	}
	return
}

// checkBestOption will cross data from prices (starting at lower prices)
// and cross that information with regions with best scores and also ensuring
// ami is offered on that specific region
//
// # Also function take cares to transfrom from AzID to AZName
//
// first option matching the requirements will be returned
func checkBestOption(amiName, amiArch string, source []spotOptionInfo,
	sps []ec2Types.SpotPlacementScore,
	availabilityZones []ec2Types.AvailabilityZone) *spotOptionInfo {
	slices.SortFunc(source,
		func(a, b spotOptionInfo) int {
			return int(a.AVGPrice - b.AVGPrice)
		})
	var score int32 = spsMaxScore
	for score > 3 {
		for _, price := range source {
			idx := slices.IndexFunc(sps, func(item ec2Types.SpotPlacementScore) bool {
				// Need transform
				spsZoneName, err := data.GetZoneName(*item.AvailabilityZoneId, availabilityZones)
				if err != nil {
					return false
				}
				var result = spsZoneName == price.AvailabilityZone &&
					*item.Score == score
				// Check for AMI is optional, i.e if we will use custom AMIs which can be replicated
				// we want the best option and the we will take care for replicate the AMI
				if result && len(amiName) > 0 {
					result, _, err = data.IsAMIOffered(
						data.ImageRequest{
							Name:   &amiName,
							Arch:   &amiArch,
							Region: &price.Region})
					if err != nil {
						return false
					}
				}
				return result
			})
			if idx != -1 {
				price.Region = *sps[idx].Region
				price.Score = *sps[idx].Score
				return &price
			}
		}
		score--
	}
	return nil
}

// describeAvailabilityZones will get information for each Az on the requested regions
// with information for matching AzID and AzName

// AzName is the general AzName
// AZId is the id for the current user (users are distributed across Azs;
//
//	meaning i.e.
//
// user 1 Name: us-west-1a ID: us-west-11, Name: us-west-1b ID: us-west-12
// user 2 Name: us-west-1a ID: us-west-12, Name: us-west-1b ID: us-west-11
// This allowsa a better distribution among users
func describeAvailabilityZones(regions []string) []ec2Types.AvailabilityZone {
	allAvailabilityZones := []ec2Types.AvailabilityZone{}
	c := make(chan data.AvailabilityZonesResult)
	for _, region := range regions {
		go data.DescribeAvailabilityZonesAsync(region, c)
	}
	for i := 0; i < len(regions); i++ {
		availabilityZonesResult := <-c
		if availabilityZonesResult.Err == nil {
			allAvailabilityZones = append(allAvailabilityZones, availabilityZonesResult.AvailabilityZones...)
		}
	}
	close(c)
	return allAvailabilityZones
}
