package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

var (
	locatioRegion string = "region"
	locationAZ    string = "availability-zone"

	filternameLocation     string = "location"
	filternameInstanceType string = "instance-type"
)

// Get InstanceTypes offerings on current location
func IsInstanceTypeOfferedByRegion(instanceType, region string) (bool, error) {
	config := aws.Config{}
	if len(region) > 0 {
		config.Region = aws.String(region)
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return false, err
	}
	svc := awsEC2.New(sess)
	o, err := svc.DescribeInstanceTypeOfferings(
		&awsEC2.DescribeInstanceTypeOfferingsInput{
			LocationType: &locatioRegion,
			Filters: []*awsEC2.Filter{
				{
					Name:   &filternameLocation,
					Values: []*string{&region}},
				{
					Name:   &filternameInstanceType,
					Values: []*string{&instanceType}},
			}})
	if err != nil {
		return false, err
	}
	return len(o.InstanceTypeOfferings) == 1, nil
}

func IsInstanceTypeOfferedByAZ(region, instanceType, az string) (bool, error) {
	config := aws.Config{}
	if len(region) > 0 {
		config.Region = aws.String(region)
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return false, err
	}
	svc := awsEC2.New(sess)
	o, err := svc.DescribeInstanceTypeOfferings(
		&awsEC2.DescribeInstanceTypeOfferingsInput{
			LocationType: &locationAZ,
			Filters: []*awsEC2.Filter{
				{
					Name:   &filternameLocation,
					Values: []*string{&az}},
				{
					Name:   &filternameInstanceType,
					Values: []*string{&instanceType}},
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
