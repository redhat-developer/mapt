package ami

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/exp/slices"

	"github.com/aws/aws-sdk-go-v2/config"
	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	awsOwnerID       string = "137112412989"
	amazonOwnerAlias string = "amazon"
	redhatOwnerID    string = "309956199498"
)

// AMIResult represents the result of an AMI lookup
type AMIResult struct {
	ImageId      string
	Name         string
	Description  string
	CreationDate string
	OwnerID      string
}

// Looks for the AMI ID on the current Region based on name
// it only allows images from AWS and self
func GetAMIByName(ctx *pulumi.Context,
	imageName string, owner []string, filters map[string]string, region string) (*AMIResult, error) {

	// Build filters
	var ec2Filters []awsEC2Types.Filter
	ec2Filters = append(ec2Filters, awsEC2Types.Filter{
		Name:   pulumi.StringRef("name"),
		Values: []string{imageName},
	})

	for k, v := range filters {
		ec2Filters = append(ec2Filters, awsEC2Types.Filter{
			Name:   pulumi.StringRef(k),
			Values: []string{v},
		})
	}

	// Build owners list
	owners := []string{awsOwnerID, redhatOwnerID, amazonOwnerAlias}
	if len(owner) > 0 {
		owners = append(owners, owner...)
	}

	// Create AWS EC2 client
	var cfgOpts config.LoadOptionsFunc
	if len(region) > 0 {
		cfgOpts = config.WithRegion(region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := awsEC2.NewFromConfig(cfg)

	// Describe images
	resp, err := client.DescribeImages(context.Background(), &awsEC2.DescribeImagesInput{
		Filters: ec2Filters,
		Owners:  owners,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Images) == 0 {
		return nil, fmt.Errorf("no AMI found with name %s", imageName)
	}

	// Sort by creation date to get the most recent
	sort.Slice(resp.Images, func(i, j int) bool {
		return *resp.Images[i].CreationDate > *resp.Images[j].CreationDate
	})

	// Return the most recent image
	img := resp.Images[0]
	return &AMIResult{
		ImageId:      *img.ImageId,
		Name:         *img.Name,
		Description:  getValue(img.Description),
		CreationDate: *img.CreationDate,
		OwnerID:      *img.OwnerId,
	}, nil
}

// getValue safely extracts string value from pointer
func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
