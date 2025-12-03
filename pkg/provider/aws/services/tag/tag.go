package tag

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func Update(ctx context.Context, tagKey, tagValue, region, resourceID string) error {
	var cfgOpts config.LoadOptionsFunc
	if len(region) > 0 {
		cfgOpts = config.WithRegion(region)
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts)
	if err != nil {
		return err
	}
	client := ec2.NewFromConfig(cfg)
	// Update tag
	tags := []ec2Types.Tag{
		{
			Key:   &tagKey,
			Value: &tagValue,
		},
	}
	_, err = client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{resourceID},
		Tags:      tags,
	})
	return err
}
