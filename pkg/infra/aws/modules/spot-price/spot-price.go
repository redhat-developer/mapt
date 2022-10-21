package spotprice

import (
	"strconv"
	"time"

	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/exp/slices"
)

func GetBestSpotPriceAsync(instanceTypes []string, productDescription, region string, c chan SpotPriceResult) {
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
		slices.SortFunc(prices, func(a, b float64) bool { return a < b })
		groupInfo.MaxPrice = prices[len(prices)-1]
		pricesGroup = append(pricesGroup, groupInfo)
	}
	return
}
