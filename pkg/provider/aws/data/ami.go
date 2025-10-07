package data

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ImageInfo struct {
	Region *string
	Image  *ec2Types.Image
}

type ImageRequest struct {
	Name, Arch, Owner *string
	Region            *string
	BlockDeviceType   *string
}

const ERROR_NO_AMI = "no AMI"

// GetAMI based on params defined by request
// In case multiple AMIs it will return the newest
func GetAMI(r ImageRequest) (*ImageInfo, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(*r.Region) > 0 {
		cfgOpts = config.WithRegion(*r.Region)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	var filterName = "name"
	filters := []ec2Types.Filter{
		{
			Name:   &filterName,
			Values: []string{*r.Name}}}
	if r.Arch != nil && len(*r.Arch) > 0 {
		filter := "architecture"
		filters = append(filters, ec2Types.Filter{
			Name:   &filter,
			Values: []string{*r.Arch}})
	}
	if r.BlockDeviceType != nil {
		filter := "block-device-mapping.volume-type"
		filters = append(filters, ec2Types.Filter{
			Name:   &filter,
			Values: []string{*r.BlockDeviceType}})
	}
	input := &ec2.DescribeImagesInput{
		Filters: filters}

	if r.Owner != nil && len(*r.Owner) > 0 {
		input.Owners = []string{*r.Owner}
		aId, err := accountId()
		if err != nil {
			return nil, err
		}
		if *aId != *r.Owner {
			input.ExecutableUsers = []string{"self"}
		}
	}
	result, err := client.DescribeImages(
		context.Background(), input)
	if err != nil {
		logging.Debugf("error checking %s in %s error is %v", *r.Name, *r.Region, err)
		return nil, err
	}
	if result == nil || len(result.Images) == 0 {
		logging.Debugf("result len 0 checking %s in %s", *r.Name, *r.Region)
		return nil, fmt.Errorf("%s %s in %s", ERROR_NO_AMI, *r.Name, *r.Region)
	}
	logging.Debugf("len %d checking %s in %s", len(result.Images), *r.Name, *r.Region)
	if err != nil {
		return nil, err
	}
	if len(result.Images) > 1 {
		slices.SortFunc(result.Images, func(a, b ec2Types.Image) int {
			ac, err := time.Parse("2006-01-02", *a.CreationDate)
			if err != nil {
				return 0
			}
			bc, err := time.Parse("2006-01-02", *b.CreationDate)
			if err != nil {
				return 0
			}
			return bc.Compare(ac)
		})
	}
	return &ImageInfo{
			Region: r.Region,
			Image:  &result.Images[0]},
		nil
}

// IsAMIOffered checks if an ami based on its Name is offered on a specific region
func IsAMIOffered(r ImageRequest) (bool, *ImageInfo, error) {
	ami, err := GetAMI(r)
	if err != nil && strings.Contains(err.Error(), ERROR_NO_AMI) {
		// If there is no AMI for this function this is not considered an error
		return false, nil, nil
	}
	return ami != nil, ami, err
}

// This function check all regions to get the AMI filter by its name
// it will return the first region where the AMI is offered
func FindAMI(amiName, amiArch *string) (*ImageInfo, error) {
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	r := make(chan *ImageInfo, len(regions))
	e := make(chan string, 1)
	closeCh := make(chan string, 1)

	defer close(r)
	defer close(closeCh)

	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		lRegion := region
		go func(r chan *ImageInfo, closeCh chan string) {
			defer wg.Done()
			select {
			case <-closeCh:
				return
			default:
				if isOffered, i, _ := IsAMIOffered(
					ImageRequest{
						Name:   amiName,
						Arch:   amiArch,
						Region: &lRegion,
					}); isOffered {
					r <- i
				}
			}
		}(r, closeCh)
	}
	go func(e chan string) {
		wg.Wait()
		defer close(e)
		e <- "done"
	}(e)
	select {
	case sAMI := <-r:
		closeCh <- "done"
		return sAMI, nil
	case <-e:
		return nil, fmt.Errorf("not AMI find with name %s on any region", *amiName)
	}
}
