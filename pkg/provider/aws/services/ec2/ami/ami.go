package ami

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/aws/aws-sdk-go-v2/config"
	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/util/logging"
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
