package data

import (
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"golang.org/x/exp/slices"

	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

// Get InstanceTypes offerings on current location
func GetInstanceTypesOfferingsByRegion(region string) ([]string, error) {
	config := aws.Config{}
	if len(region) > 0 {
		config.Region = aws.String(region)
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	o, err := svc.DescribeInstanceTypeOfferings(nil)
	if err != nil {
		return nil, err
	}
	return util.ArrayConvert(o.InstanceTypeOfferings,
			func(item *awsEC2.InstanceTypeOffering) string {
				return *item.InstanceType
			}),
		nil
}

// Check if a instance type is available at the current location
func IsInstaceTypeOffered(instanceType, region string) (bool, error) {
	o, err := GetInstanceTypesOfferingsByRegion(region)
	if err != nil {
		return false, err
	}
	return slices.Contains(o, instanceType), nil
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
			if is, err := IsInstaceTypeOffered(
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
