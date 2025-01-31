package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
)

const (
	s3Prefix        = "s3://"
	s3PathSeparator = "/"
)

func ValidateS3Path(p string) bool { return strings.HasPrefix(p, s3Prefix) }

func GetBucketLocationFromS3Path(p string) (*string, error) {
	bucket, err := getBucketFromS3Path(p)
	if err != nil {
		return nil, err
	}
	return GetBucketLocation(*bucket)
}

func GetBucketLocation(bucketName string) (*string, error) {
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	// b, err := client.HeadBucket(
	// 	context.Background(),
	// 	&s3.HeadBucketInput{
	// 		Bucket: &bucketName,
	// 	})
	b, err := client.GetBucketLocation(
		context.Background(),
		&s3.GetBucketLocationInput{
			Bucket: &bucketName,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bucket region: %w", err)
	}
	// AWS returns nil for "us-east-1", so we set it manually
	region := string(b.LocationConstraint)
	if region == "" {
		region = awsConstants.DefaultAWSRegion
	}

	return &region, nil
}

func getBucketFromS3Path(p string) (*string, error) {
	if !ValidateS3Path(p) {
		return nil, fmt.Errorf("this is not a valid s3 path")
	}
	pWithoutProtocol := strings.TrimPrefix(p, s3Prefix)
	pParts := strings.Split(pWithoutProtocol, s3PathSeparator)
	return &pParts[0], nil
}
