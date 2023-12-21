package ami

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/exp/slices"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go-v2/config"
	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	awsOwnerID       string = "137112412989"
	amazonOwnerAlias string = "amazon"
	redhatOwnerID    string = "309956199498"
)

// Looks for the AMI ID on the current Region based on name
// it only allows images from AWS and self
func GetAMIByName(ctx *pulumi.Context,
	imageName, owner string, filters map[string]string) (*ec2.LookupAmiResult, error) {
	mostRecent := true
	lookupfilters := []ec2.GetAmiFilter{
		{
			Name:   "name",
			Values: []string{imageName},
		},
	}
	for k, v := range filters {
		lookupfilters = append(lookupfilters, ec2.GetAmiFilter{
			Name:   k,
			Values: []string{v},
		})
	}
	owners := []string{awsOwnerID, redhatOwnerID, amazonOwnerAlias}
	if len(owner) > 0 {
		owners = append(owners, owner)
	}
	return ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		Filters:    lookupfilters,
		Owners:     owners,
		MostRecent: &mostRecent,
	})
}

type ImageInfo struct {
	Region *string
	Image  *awsEC2Types.Image
}

// IsAMIOffered checks if an ami based on its Name is offered on a specific region
func IsAMIOffered(amiName, amiArch, region *string) (bool, *ImageInfo, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(*region) > 0 {
		cfgOpts = config.WithRegion(*region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return false, nil, err
	}
	client := awsEC2.NewFromConfig(cfg)
	var filterName = "name"
	filters := []awsEC2Types.Filter{
		{
			Name:   &filterName,
			Values: []string{*amiName}}}
	if amiArch != nil {
		var filterArch = "architecture"
		filters = append(filters, awsEC2Types.Filter{
			Name:   &filterArch,
			Values: []string{*amiArch}})
	}
	result, err := client.DescribeImages(
		context.Background(),
		&awsEC2.DescribeImagesInput{
			Filters: filters})
	if err != nil {
		logging.Debugf("error checking %s in %s error is %v", *amiName, *region, err)
		return false, nil, err
	}
	if result == nil || len(result.Images) == 0 {
		logging.Debugf("result len 0 checking %s in %s", *amiName, *region)
		return false, nil, nil
	}
	logging.Debugf("len %d checking %s in %s", len(result.Images), *amiName, *region)
	if err != nil {
		return false, nil, err
	}
	return len(result.Images) > 0,
		&ImageInfo{
			Region: region,
			Image:  &result.Images[0]},
		nil
}

// This function check all regions to get the AMI filter by its name
// it will return the first region where the AMI is offered
func FindAMI(amiName, amiArch *string) (*ImageInfo, error) {
	regions, err := data.GetRegions()
	if err != nil {
		return nil, err
	}
	r := make(chan *ImageInfo, len(regions))
	e := make(chan string, 1)
	defer close(r)
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		lRegion := region
		go func(r chan *ImageInfo) {
			defer wg.Done()
			if isOffered, i, _ := IsAMIOffered(
				amiName, amiArch, &lRegion); isOffered {
				r <- i
			}
		}(r)
	}
	go func(e chan string) {
		wg.Wait()
		defer close(e)
		e <- "done"
	}(e)
	select {
	case sAMI := <-r:
		return sAMI, nil
	case <-e:
		return nil, fmt.Errorf("not AMI find with name %s on any region", *amiName)
	}
}

// Enable Fast Launchon AMI, it will only work with Windows instances
func EnableFastLaunch(region *string, amiID *string, maxParallel *int32) error {
	logging.Debugf("Enabling fast launch for ami %s", *amiID)
	var cfgOpts config.LoadOptionsFunc
	if len(*region) > 0 {
		cfgOpts = config.WithRegion(*region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return err
	}
	client := awsEC2.NewFromConfig(cfg)
	o, err := client.EnableFastLaunch(context.Background(),
		&awsEC2.EnableFastLaunchInput{
			ImageId:             amiID,
			MaxParallelLaunches: maxParallel,
		})
	if err != nil {
		return nil
	}
	fastLaunchState := o.State
	for fastLaunchState != awsEC2Types.FastLaunchStateCodeEnabled {
		dfl, err := client.DescribeFastLaunchImages(
			context.Background(),
			&awsEC2.DescribeFastLaunchImagesInput{
				ImageIds: []string{*amiID},
			})
		if err != nil {
			return nil
		}
		if len(dfl.FastLaunchImages) != 1 {
			return fmt.Errorf("unexpected result enabling fast launch for AMI %s on region %s", *amiID, *region)
		}
		fastLaunchState = dfl.FastLaunchImages[0].State
		if fastLaunchState == awsEC2Types.FastLaunchStateCodeEnabledFailed {
			return fmt.Errorf("error enabling fast launch on AMI %s", *amiID)
		}
		if !slices.Contains([]awsEC2Types.FastLaunchStateCode{
			awsEC2Types.FastLaunchStateCodeEnabling,
			awsEC2Types.FastLaunchStateCodeEnabled}, fastLaunchState) {
			return fmt.Errorf("unexpected state while enabling fast launch on AMI %s", *amiID)
		}
	}
	return nil
}
