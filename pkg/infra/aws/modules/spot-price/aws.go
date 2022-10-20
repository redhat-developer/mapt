package spotprice

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"

	"golang.org/x/exp/slices"
)

func GetBestSpotPrice(instanceType, productDescription, region string) (string, string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)})
	if err != nil {
		return "", "", err
	}
	svc := awsEC2.New(sess)
	starTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	history, err := svc.DescribeSpotPriceHistory(
		&awsEC2.DescribeSpotPriceHistoryInput{
			InstanceTypes: []*string{&instanceType},
			Filters: []*awsEC2.Filter{
				{
					Name:   &spotQueryFilterProductDescription,
					Values: []*string{&productDescription},
				},
			},
			StartTime: &starTime,
			EndTime:   &endTime,
		})
	if err != nil {
		return "", "", err
	}
	if len(history.SpotPriceHistory) == 0 {
		return "", "", fmt.Errorf("non available prices for the search criteria at %s", region)
	}
	// Check if history is empty?
	slices.SortFunc(history.SpotPriceHistory,
		func(a, b *awsEC2.SpotPrice) bool {
			aPrice, err := strconv.ParseFloat(*a.SpotPrice, 64)
			if err != nil {
				return false
			}
			bPrice, err := strconv.ParseFloat(*b.SpotPrice, 64)
			if err != nil {
				return false
			}
			return aPrice < bPrice
		})
	return *history.SpotPriceHistory[0].SpotPrice, *history.SpotPriceHistory[0].AvailabilityZone, nil
}

func GetBestSpotPriceAsync(instanceType, productDescription, region string, c chan SpotPriceResult) {
	price, availabilityZone, err := GetBestSpotPrice(
		instanceType, productDescription, region)
	c <- SpotPriceResult{
		Data: SpotPriceData{
			Price:            price,
			AvailabilityZone: availabilityZone,
			Region:           region,
			InstanceType:     instanceType},
		Err: err}

}
