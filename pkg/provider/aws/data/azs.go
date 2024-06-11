package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"golang.org/x/exp/slices"
)

func GetRandomAvailabilityZone(region string, excludedAZs []string) (*string, error) {
	azs, err := DescribeAvailabilityZones(region)
	if err != nil {
		return nil, err
	}
	if len(excludedAZs) > 0 {
		azs = slices.DeleteFunc(azs, func(a ec2types.AvailabilityZone) bool {
			return slices.Contains(excludedAZs, *a.ZoneName)
		})
	}
	return azs[util.Random(len(azs)-1, 0)].ZoneName, nil
}

func GetAvailabilityZones() []string {
	azs, err := describeAvailabilityZones("")
	if err != nil {
		logging.Error(err)
		return nil
	}
	return util.ArrayConvert(azs, func(source ec2types.AvailabilityZone) string {
		return *source.ZoneName
	})
}

type AvailabilityZonesResult struct {
	AvailabilityZones []ec2types.AvailabilityZone
	Err               error
}

func DescribeAvailabilityZonesAsync(regionName string, c chan AvailabilityZonesResult) {
	data, err := DescribeAvailabilityZones(regionName)
	c <- AvailabilityZonesResult{
		AvailabilityZones: data,
		Err:               err}

}

func DescribeAvailabilityZones(regionName string) ([]ec2types.AvailabilityZone, error) {
	return describeAvailabilityZones(regionName)
}

func describeAvailabilityZones(regionName string) ([]ec2types.AvailabilityZone, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(regionName) > 0 {
		cfgOpts = config.WithRegion(regionName)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	// TODO check what happen when true and region name
	input := ec2.DescribeAvailabilityZonesInput{
		// AllAvailabilityZones: aws.Bool(true),
	}
	input.Filters = []ec2types.Filter{
		{
			Name:   aws.String("zone-type"),
			Values: []string{"availability-zone"},
		},
	}
	resultAZs, err := client.DescribeAvailabilityZones(context.Background(), &input)
	if err != nil {
		return nil, err
	}
	return resultAZs.AvailabilityZones, nil
}

func GetZoneName(azID string, azDescriptions []ec2types.AvailabilityZone) (string, error) {
	idx := slices.IndexFunc(azDescriptions,
		func(azDescription ec2types.AvailabilityZone) bool {
			return azID == *azDescription.ZoneId
		})
	if idx == -1 {
		return "", fmt.Errorf("az id not found")
	}
	return *azDescriptions[idx].ZoneName, nil
}
