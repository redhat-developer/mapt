package data

import (
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/aws/aws-sdk-go/aws/session"

	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

var (
	optInStatusFilter      string = "opt-in-status"
	optInStatusNorRequired string = "opt-in-not-required"
)

func GetRegions() ([]string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	regions, err := svc.DescribeRegions(
		&awsEC2.DescribeRegionsInput{
			Filters: []*awsEC2.Filter{
				{
					Name:   &optInStatusFilter,
					Values: []*string{&optInStatusNorRequired},
				},
			}})
	if err != nil {
		return nil, err
	}
	return util.ArrayConvert(regions.Regions,
			func(item *awsEC2.Region) string {
				return *item.RegionName
			}),
		nil
}
