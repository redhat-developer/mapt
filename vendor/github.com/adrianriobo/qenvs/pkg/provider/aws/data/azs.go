package data

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/exp/slices"
)

func GetRandomAvailabilityZone(region string) (*string, error) {
	azs, err := DescribeAvailabilityZones(region)
	if err != nil {
		return nil, err
	}
	return azs[util.Random(len(azs)-1, 0)].ZoneName, nil
}

func GetAvailabilityZones() []string {
	azs, err := describeAvailabilityZones("")
	if err != nil {
		logging.Error(err)
		return nil
	}
	return util.ArrayConvert(azs, func(source *ec2.AvailabilityZone) string {
		return *source.ZoneName
	})
}

type AvailabilityZonesResult struct {
	AvailabilityZones []*ec2.AvailabilityZone
	Err               error
}

func DescribeAvailabilityZonesAsync(regionName string, c chan AvailabilityZonesResult) {
	data, err := DescribeAvailabilityZones(regionName)
	c <- AvailabilityZonesResult{
		AvailabilityZones: data,
		Err:               err}

}

func DescribeAvailabilityZones(regionName string) ([]*ec2.AvailabilityZone, error) {
	return describeAvailabilityZones(regionName)
}

func describeAvailabilityZones(regionName string) ([]*ec2.AvailabilityZone, error) {
	config := aws.Config{}
	if len(regionName) > 0 {
		config.Region = aws.String(regionName)
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}
	svc := ec2.New(sess)
	// TODO check what happen when true and region name
	input := &ec2.DescribeAvailabilityZonesInput{
		// AllAvailabilityZones: aws.Bool(true),
	}
	input.Filters = []*ec2.Filter{
		{
			Name:   aws.String("zone-type"),
			Values: aws.StringSlice([]string{"availability-zone"}),
		},
	}
	resultAZs, err := svc.DescribeAvailabilityZones(input)
	if err != nil {
		return nil, err
	}
	return resultAZs.AvailabilityZones, nil
}

func GetZoneName(azID string, azDescriptions []*ec2.AvailabilityZone) (string, error) {
	idx := slices.IndexFunc(azDescriptions,
		func(azDescription *ec2.AvailabilityZone) bool {
			return azID == *azDescription.ZoneId
		})
	if idx == -1 {
		return "", fmt.Errorf("az id not found")
	}
	return *azDescriptions[idx].ZoneName, nil
}
