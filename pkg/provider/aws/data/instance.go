package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"golang.org/x/exp/maps"
)

type InstanceResquest struct {
	Tags map[string]string
}

func GetInstanceByRegion(ctx context.Context, r InstanceResquest, regionName string) ([]ec2Types.Instance, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(regionName) > 0 {
		cfgOpts = config.WithRegion(regionName)
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	tagKey := "tag-key"
	i, err := client.DescribeInstances(
		ctx,
		&ec2.DescribeInstancesInput{
			Filters: []ec2Types.Filter{
				{
					Name:   &tagKey,
					Values: maps.Keys(r.Tags),
				},
			},
		})
	if err != nil {
		return nil, err
	}
	if len(i.Reservations) != 1 || len(i.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("dedicated host was not found on current region")
	}
	if len(r.Tags) > 0 {
		return util.ArrayFilter(i.Reservations[0].Instances,
			func(h ec2Types.Instance) bool {
				return allTagsMatches(r.Tags, h.Tags)
			}), nil
	}
	return i.Reservations[0].Instances, nil
}
