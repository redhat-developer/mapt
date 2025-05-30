package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/redhat-developer/mapt/pkg/util"
)

const (
	ResourceTypeECS = "ecs:task-definition"
)

var (
	ErrorNoResourcesFound = fmt.Errorf("no resources matching tags")
)

// Return a list of arns for the type of resource which have all tags
func GetResourcesMatchingTags(resourceType string, tags map[string]string) ([]string, error) {
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	tagsClient := resourcegroupstaggingapi.NewFromConfig(cfg)
	ro, err := tagsClient.GetResources(context.Background(),
		&resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: []string{resourceType},
		})
	if err != nil {
		return nil, err
	}
	rFiltered := util.ArrayFilter(ro.ResourceTagMappingList,
		func(r types.ResourceTagMapping) bool {
			ft := util.ArrayFilter(r.Tags,
				func(rt types.Tag) bool {
					for k, v := range tags {
						if *rt.Key == k && *rt.Value == v {
							return true
						}
					}
					return false
				})
			return len(tags) == len(ft)
		})
	if len(rFiltered) == 0 {
		return nil, ErrorNoResourcesFound
	}
	return util.ArrayConvert(
		rFiltered,
		func(rt types.ResourceTagMapping) string {
			return *rt.ResourceARN
		}), nil
}
