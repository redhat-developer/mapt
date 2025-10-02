package data

import (
	"context"

	"github.com/redhat-developer/mapt/pkg/util"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	optInStatusFilter      string = "opt-in-status"
	OptInStatusNotRequired string = "opt-in-not-required"
	OptInStatusOptedIn     string = "opted-in"
)

func GetRegions() ([]string, error) {
	return GetRegionsByOptInStatus([]string{OptInStatusNotRequired, OptInStatusOptedIn})
}

func GetRegionsByOptInStatus(optInStaus []string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	regions, err := client.DescribeRegions(
		context.Background(),
		&ec2.DescribeRegionsInput{
			Filters: []ec2Types.Filter{
				{
					Name:   &optInStatusFilter,
					Values: optInStaus,
				},
			}})
	if err != nil {
		return nil, err
	}
	return util.ArrayConvert(regions.Regions,
			func(item ec2Types.Region) string {
				return *item.RegionName
			}),
		nil
}
