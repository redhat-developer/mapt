package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"golang.org/x/exp/slices"
)

func GetRandomAvailabilityZone(ctx context.Context, region string, excludedAZs []string) (*string, error) {
	azs, err := DescribeAvailabilityZones(ctx, region)
	if err != nil {
		return nil, err
	}
	if len(excludedAZs) > 0 {
		azs = slices.DeleteFunc(azs, func(a ec2Types.AvailabilityZone) bool {
			return slices.Contains(excludedAZs, *a.ZoneName)
		})
	}
	return azs[util.Random(len(azs)-1, 0)].ZoneName, nil
}

func GetAvailabilityZones(ctx context.Context, region string, excludedZoneIDs []string) []string {
	azs, err := describeAvailabilityZones(ctx, region, excludedZoneIDs)
	if err != nil {
		logging.Error(err)
		return nil
	}
	return util.ArrayConvert(azs, func(source ec2Types.AvailabilityZone) string {
		return *source.ZoneName
	})
}

type AvailabilityZonesResult struct {
	AvailabilityZones []ec2Types.AvailabilityZone
	Err               error
}

func describeAvailabilityZonesAsync(ctx context.Context, regionName string, c chan AvailabilityZonesResult) {
	data, err := DescribeAvailabilityZones(ctx, regionName)
	c <- AvailabilityZonesResult{
		AvailabilityZones: data,
		Err:               err}

}

func DescribeAvailabilityZones(ctx context.Context, regionName string) ([]ec2Types.AvailabilityZone, error) {
	return describeAvailabilityZones(ctx, regionName, nil)
}

func describeAvailabilityZones(ctx context.Context, regionName string, excludedZoneIDs []string) ([]ec2Types.AvailabilityZone, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(regionName) > 0 {
		cfgOpts = config.WithRegion(regionName)
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	// TODO check what happen when true and region name
	input := ec2.DescribeAvailabilityZonesInput{
		// AllAvailabilityZones: aws.Bool(true),
	}
	input.Filters = []ec2Types.Filter{
		{
			Name:   aws.String("zone-type"),
			Values: []string{"availability-zone"},
		},
	}
	resultAZs, err := client.DescribeAvailabilityZones(ctx, &input)
	if err != nil {
		return nil, err
	}
	
	// Filter out excluded zone-ids if provided
	if len(excludedZoneIDs) > 0 {
		filteredAZs := slices.DeleteFunc(resultAZs.AvailabilityZones, func(az ec2Types.AvailabilityZone) bool {
			return slices.Contains(excludedZoneIDs, *az.ZoneId)
		})
		return filteredAZs, nil
	}
	
	return resultAZs.AvailabilityZones, nil
}

func getZoneName(azID string, azDescriptions []ec2Types.AvailabilityZone) (string, error) {
	idx := slices.IndexFunc(azDescriptions,
		func(azDescription ec2Types.AvailabilityZone) bool {
			return azID == *azDescription.ZoneId
		})
	if idx == -1 {
		return "", fmt.Errorf("az id not found")
	}
	return *azDescriptions[idx].ZoneName, nil
}

// describeAvailabilityZones will get information for each Az on the requested regions
// with information for matching AzID and AzName

// AzName is the general AzName
// AZId is the id for the current user (users are distributed across Azs;
//
//	meaning i.e.
//
// user 1 Name: us-west-1a ID: us-west-11, Name: us-west-1b ID: us-west-12
// user 2 Name: us-west-1a ID: us-west-12, Name: us-west-1b ID: us-west-11
// This allowsa a better distribution among users
func describeAvailabilityZonesByRegions(ctx context.Context, regions []string) map[string][]ec2Types.AvailabilityZone {
	result := make(map[string][]ec2Types.AvailabilityZone)
	c := make(chan AvailabilityZonesResult)
	for _, region := range regions {
		lRegion := region
		go describeAvailabilityZonesAsync(ctx, lRegion, c)
	}
	for i := 0; i < len(regions); i++ {
		availabilityZonesResult := <-c
		if availabilityZonesResult.Err == nil {
			region := availabilityZonesResult.AvailabilityZones[0].RegionName
			result[*region] = append(result[*region], availabilityZonesResult.AvailabilityZones...)
		}
	}
	close(c)
	return result
}
