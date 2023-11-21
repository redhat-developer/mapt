package ami

import (
	"fmt"
	"sync"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
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
	Image  *awsEC2.Image
}

// IsAMIOffered checks if an ami based on its Name is offered on a specific region
func IsAMIOffered(amiName, amiArch, region string) (bool, *ImageInfo, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return false, nil, err
	}
	svc := awsEC2.New(sess)
	var filterName = "name"
	var filterArch = "architecture"
	result, err := svc.DescribeImages(&awsEC2.DescribeImagesInput{
		Filters: []*awsEC2.Filter{
			{
				Name:   &filterName,
				Values: aws.StringSlice([]string{amiName})},
			{
				Name:   &filterArch,
				Values: aws.StringSlice([]string{amiArch})},
		}})
	if err != nil {
		logging.Debugf("error checking %s in %s error is %v", amiName, region, err)
		return false, nil, err
	}
	if result == nil || len(result.Images) == 0 {
		logging.Debugf("result len 0 checking %s in %s", amiName, region)
		return false, nil, nil
	}
	logging.Debugf("len %d checking %s in %s", len(result.Images), amiName, region)
	if err != nil {
		return false, nil, err
	}
	return len(result.Images) > 0,
		&ImageInfo{
			Region: &region,
			Image:  result.Images[0]},
		nil
}

// This function check all regions to get the AMI filter by its name
// it will return the first region where the AMI is offered
func FindAMI(amiName, amiArch string) (*ImageInfo, error) {
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
				amiName, amiArch, lRegion); isOffered {
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
		return nil, fmt.Errorf("not AMI find with name %s on any region", amiName)
	}
}
