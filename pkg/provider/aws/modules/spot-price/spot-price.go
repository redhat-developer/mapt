package spotprice

import (
	"strconv"
	"time"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/meta/azs"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/exp/slices"
)

func (r SpotPriceRequest) deployer(ctx *pulumi.Context) error {
	bsb, err := BestSpotBidOrder(ctx, r.Name, r.TargetHostID)
	if err != nil {
		return err
	}
	exportSpotPriceGroup(bsb.Price, ctx)
	return nil
}

func getPricesPerRegion(productDescription string,
	regions, instanceTypes []string) []SpotPriceGroup {
	worldwidePrices := []SpotPriceGroup{}
	c := make(chan SpotPriceResult)
	for _, region := range regions {
		go getBestSpotPriceAsync(
			instanceTypes,
			productDescription,
			region,
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

func getDescribeAvailabilityZones(regions []string) []*ec2.AvailabilityZone {
	allAvailabilityZones := []*ec2.AvailabilityZone{}
	c := make(chan azs.AvailabilityZonesResult)
	for _, region := range regions {
		go azs.DescribeAvailabilityZonesAsync(region, c)
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

func checkBestOption(source []SpotPriceGroup, sps []*ec2.SpotPlacementScore, availabilityZones []*ec2.AvailabilityZone) *SpotPriceGroup {
	slices.SortFunc(source,
		func(a, b SpotPriceGroup) int {
			return int(a.AVGPrice - b.AVGPrice)
		})
	var score int64 = spsMaxScore
	for score > 3 {
		for _, price := range source {
			idx := slices.IndexFunc(sps, func(item *ec2.SpotPlacementScore) bool {
				// Need transform
				spsZoneName, err := azs.GetZoneName(*item.AvailabilityZoneId, availabilityZones)
				if err != nil {
					return false
				}
				return spsZoneName == price.AvailabilityZone &&
					*item.Score == score
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

func getBestSpotPriceAsync(instanceTypes []string, productDescription, region string, c chan SpotPriceResult) {
	data, err := getBestSpotPrice(instanceTypes, productDescription, region)
	c <- SpotPriceResult{
		Prices: data,
		Err:    err}

}

func getBestSpotPrice(instanceTypes []string, productDescription, region string) (pricesGroup []SpotPriceGroup, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	starTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	history, err := svc.DescribeSpotPriceHistory(
		&awsEC2.DescribeSpotPriceHistoryInput{
			InstanceTypes: aws.StringSlice(instanceTypes),
			Filters: []*awsEC2.Filter{
				{
					Name:   aws.String(spotQueryFilterProductDescription),
					Values: []*string{&productDescription},
				},
			},
			StartTime: &starTime,
			EndTime:   &endTime,
		})
	if err != nil {
		return nil, err
	}
	spotPriceGroups := util.SplitSlice(history.SpotPriceHistory, func(priceData *awsEC2.SpotPrice) SpotPriceGroup {
		return SpotPriceGroup{
			AvailabilityZone: *priceData.AvailabilityZone,
		}
	})
	logging.Debugf("grouped prices %v", spotPriceGroups)
	for groupInfo, pricesHistory := range spotPriceGroups {
		prices := util.ArrayConvert(pricesHistory, func(priceHisotry *awsEC2.SpotPrice) float64 {
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
		pricesGroup = append(pricesGroup, groupInfo)
	}
	return
}
