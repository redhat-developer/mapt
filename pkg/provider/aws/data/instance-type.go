package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	locationTypeRegion string = "region"
	locationTypeAZ     string = "availability-zone"

	filternameLocation     string = "location"
	filternameInstanceType string = "instance-type"
)

type LocationArgs struct {
	Region, Az *string
}

// Check if InstanceType is offered on current location
// it is valid for Regions or Azs
// if az is nill it will check on region
func IsInstanceTypeOfferedByLocation(instanceType string, args *LocationArgs) (bool, error) {
	offerings, err := FilterInstaceTypesOfferedByLocation([]string{instanceType}, args)
	return len(offerings) == 1, err
}

// Get InstanceTypes offerings on current location
// it is valid for Regions or Azs
func FilterInstaceTypesOfferedByLocation(instanceTypes []string, args *LocationArgs) ([]string, error) {
	cfg, err := getConfig(*args.Region)
	if err != nil {
		return nil, err
	}
	location := *args.Region
	locationType := locationTypeRegion
	if args.Az != nil {
		location = *args.Az
		locationType = locationTypeAZ
	}
	client := ec2.NewFromConfig(cfg)
	o, err := client.DescribeInstanceTypeOfferings(
		context.Background(),
		&ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: ec2Types.LocationType(locationType),
			Filters: []ec2Types.Filter{
				{
					Name:   &filternameLocation,
					Values: []string{location}},
				{
					Name:   &filternameInstanceType,
					Values: instanceTypes},
			}})
	if err != nil {
		return nil, err
	}
	var offerings []string
	for _, o := range o.InstanceTypeOfferings {
		offerings = append(offerings, string(o.InstanceType))
	}
	return offerings, nil
}

// func IsInstanceTypeOfferedByAZ(region, instanceType, az string) (bool, error) {
// 	var cfgOpts config.LoadOptionsFunc
// 	if len(region) > 0 {
// 		cfgOpts = config.WithRegion(region)
// 	}
// 	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
// 	if err != nil {
// 		return false, err
// 	}
// 	client := ec2.NewFromConfig(cfg)
// 	o, err := client.DescribeInstanceTypeOfferings(
// 		context.Background(),
// 		&ec2.DescribeInstanceTypeOfferingsInput{
// 			LocationType: ec2Types.LocationType(locationAZ),
// 			Filters: []ec2Types.Filter{
// 				{
// 					Name:   &filternameLocation,
// 					Values: []string{az}},
// 				{
// 					Name:   &filternameInstanceType,
// 					Values: []string{instanceType}},
// 			}})
// 	if err != nil {
// 		return false, err
// 	}
// 	return len(o.InstanceTypeOfferings) == 1, nil
// }

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
			if is, err := IsInstanceTypeOfferedByLocation(
				instanceType,
				&LocationArgs{
					Region: &lRegion,
				}); err == nil && is {
				c <- lRegion
			}
		}(c)
	}
	// First region with offering is enoguh
	oRegion := <-c
	return &oRegion, nil
}
