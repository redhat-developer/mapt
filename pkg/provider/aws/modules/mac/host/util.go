package host

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"golang.org/x/exp/slices"
)

// Get all dedicated hosts matching the tags + arch
// it will return the list ordered by allocation time
func GetMatchingHostsInformation(mCtx *maptContext.Context, arch string) ([]*mac.HostInformation, error) {
	matchingTags := mCtx.GetTags()
	matchingTags[tagKeyArch] = arch
	return GetMatchingHostsInStateInformation(mCtx.Context(), matchingTags, nil)
}

// Get all dedicated hosts matching the tags + arch
// it will return the list ordered by allocation time
func GetPoolDedicatedHostsInformation(ctx context.Context, id *PoolID) ([]*mac.HostInformation, error) {
	return GetMatchingHostsInStateInformation(ctx, id.asTags(), nil)
}

// Get all dedicated hosts in available state ordered based on the allocation time
// newer allocations go first
// func getMatchingAvailableHostsInformation(arch string) ([]HostInformation, error) {
// 	as := ec2Types.AllocationStateAvailable
// 	return getMatchingHostsInStateInformation(arch, &as)
// }

// format for remote backed url when creating the dedicated host
// the backed url from param is used as base and the ID is appended as sub path
func getBackedURL(mCtx *maptContext.Context) string {
	if strings.Contains(mCtx.BackedURL(), "file://") {
		return mCtx.BackedURL()
	}
	return fmt.Sprintf("%s/%s", mCtx.BackedURL(), mCtx.RunID())
}

// Get all dedicated hosts by tag and state
func GetMatchingHostsInStateInformation(ctx context.Context, matchingTags map[string]string, state *ec2Types.AllocationState) ([]*mac.HostInformation, error) {
	hosts, err := data.GetDedicatedHosts(ctx, data.DedicatedHostResquest{
		Tags: matchingTags,
	})
	if err != nil {
		return nil, err
	}
	var r []*mac.HostInformation
	for _, dh := range hosts {
		if state == nil || (state != nil && dh.State == *state) {
			r = append(r, GetHostInformation(dh))
		}
	}
	// Order by allocation time, first newest
	if len(r) > 1 {
		// Sort the slice by time (ascending order)
		sort.Slice(r, func(i, j int) bool {
			return r[i].Host.AllocationTime.Before(*r[j].Host.AllocationTime)
		})
	}
	return r, nil
}

// Compose information around dedicated host
func GetHostInformation(h ec2Types.Host) *mac.HostInformation {
	az := *h.AvailabilityZone
	region := az[:len(az)-1]
	archValue := awsArchIDbyArch[*getTagValue(h.Tags, tagKeyArch)]
	return &mac.HostInformation{
		Arch:        &archValue,
		OSVersion:   getTagValue(h.Tags, tagKeyOSVersion),
		BackedURL:   getTagValue(h.Tags, tagKeyBackedURL),
		Prefix:      getTagValue(h.Tags, tagKeyPrefix),
		ProjectName: getTagValue(h.Tags, maptContext.TagKeyProjectName),
		RunID:       getTagValue(h.Tags, maptContext.TagKeyRunID),
		Region:      &region,
		Host:        &h,
	}
}

func getTagValue(tags []ec2Types.Tag, tagKey string) *string {
	return tags[slices.IndexFunc(
		tags,
		func(t ec2Types.Tag) bool {
			return *t.Key == tagKey
		})].Value
}

// checks if the machine can be created on the current location (machine type is available on the region)
// if it available it returns the region name
// if not offered and machine should be created on the region it will return an error
// if not offered and machine could be created anywhere it will get a region offering the machine and return its name
func getRegion(ctx context.Context, arch string, fixedLocation bool) (*string, error) {
	region := os.Getenv("AWS_DEFAULT_REGION")
	logging.Debugf("checking if %s is offered at %s",
		arch,
		region)
	isOffered, err := data.IsInstanceTypeOfferedByLocation(
		ctx,
		mac.TypesByArch[arch],
		&data.LocationArgs{
			Region: &region,
		})
	if err != nil {
		return nil, err
	}
	if isOffered {
		logging.Debugf("%s offers it",
			os.Getenv("AWS_DEFAULT_REGION"))
		region := os.Getenv("AWS_DEFAULT_REGION")
		return &region, nil
	}
	if !isOffered && fixedLocation {
		return nil, fmt.Errorf("the requested mac %s is not available at the current region %s and the fixed-location flag has been set",
			arch,
			os.Getenv("AWS_DEFAULT_REGION"))
	}
	// We look for a region offering the type of instance
	logging.Debugf("%s is not offered, a new region offering it will be used instead",
		os.Getenv("AWS_DEFAULT_REGION"))
	return data.LokupRegionOfferingInstanceType(
		ctx,
		mac.TypesByArch[arch])
}

// Get a random AZ from the requested region, it ensures the az offers the instance type
func getAZ(ctx context.Context, region, arch string) (az *string, err error) {
	isOffered := false
	var excludedAZs []string
	for !isOffered {
		az, err = data.GetRandomAvailabilityZone(ctx, region, excludedAZs)
		if err != nil {
			return nil, err
		}
		isOffered, err = data.IsInstanceTypeOfferedByLocation(
			ctx,
			mac.TypesByArch[arch],
			&data.LocationArgs{
				Region: &region,
				Az:     az,
			})
		if err != nil {
			return nil, err
		}
		if !isOffered {
			excludedAZs = append(excludedAZs, *az)
		}
	}
	return
}
