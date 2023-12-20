package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	locatioRegion string = "region"
	locationAZ    string = "availability-zone"

	filternameLocation     string = "location"
	filternameInstanceType string = "instance-type"
)

// Get InstanceTypes offerings on current location
func IsInstanceTypeOfferedByRegion(instanceType, region string) (bool, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(region) > 0 {
		cfgOpts = config.WithRegion(region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return false, err
	}
	client := ec2.NewFromConfig(cfg)
	o, err := client.DescribeInstanceTypeOfferings(
		context.Background(),
		&ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: ec2Types.LocationType(locatioRegion),
			Filters: []ec2Types.Filter{
				{
					Name:   &filternameLocation,
					Values: []string{region}},
				{
					Name:   &filternameInstanceType,
					Values: []string{instanceType}},
			}})
	if err != nil {
		return false, err
	}
	return len(o.InstanceTypeOfferings) == 1, nil
}

func IsInstanceTypeOfferedByAZ(region, instanceType, az string) (bool, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(region) > 0 {
		cfgOpts = config.WithRegion(region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return false, err
	}
	client := ec2.NewFromConfig(cfg)
	o, err := client.DescribeInstanceTypeOfferings(
		context.Background(),
		&ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: ec2Types.LocationType(locationAZ),
			Filters: []ec2Types.Filter{
				{
					Name:   &filternameLocation,
					Values: []string{az}},
				{
					Name:   &filternameInstanceType,
					Values: []string{instanceType}},
			}})
	if err != nil {
		return false, err
	}
	return len(o.InstanceTypeOfferings) == 1, nil
}

// Check on all regions which offers the type of instance got one having it
func LokupRegionOfferingInstanceType(instanceType string) (*string, error) {
	// We need to check on all regions
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	c := make(chan string)
	for _, region := range regions {
		lRegion := region
		go func(c chan string) {
			if is, err := IsInstanceTypeOfferedByRegion(
				instanceType,
				lRegion); err == nil && is {
				c <- lRegion
			}
		}(c)
	}
	// First region with offering is enoguh
	oRegion := <-c
	return &oRegion, nil
}
