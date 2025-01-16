package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"golang.org/x/exp/slices"
)

// Check if a host contais exactly all tags defined by tags param
func allTagsMatches(tags map[string]string, tTags []ec2Types.Tag) bool {
	count := 0
	for k, v := range tags {
		if slices.ContainsFunc(tTags, func(t ec2Types.Tag) bool {
			return *t.Key == k && *t.Value == v
		}) {
			count++
		}
	}
	return count == len(tags)
}

func getGlobalConfig() (aws.Config, error) {
	return getConfig("")
}

func getConfig(region string) (aws.Config, error) {
	if len(region) > 0 {
		return config.LoadDefaultConfig(
			context.Background(),
			config.WithRegion(region))
	}
	return config.LoadDefaultConfig(context.Background())
}
