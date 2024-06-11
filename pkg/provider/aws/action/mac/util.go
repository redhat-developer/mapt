package mac

import (
	"fmt"
	"os"
	"strings"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"golang.org/x/exp/slices"
)

// Compose information around dedicated host
func getHostInformation(h ec2Types.Host) *HostInformation {
	az := *h.AvailabilityZone
	region := az[:len(az)-1]
	archValue := awsArchIDbyArch[*getTagValue(h.Tags, tagKeyArch)]
	return &HostInformation{
		Arch:        &archValue,
		BackedURL:   getTagValue(h.Tags, tagKeyBackedURL),
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

// format for remote backed url when creating the dedicated host
// the backed url from param is used as base and the ID is appended as sub path
func getBackedURL() string {
	if strings.Contains(maptContext.BackedURL(), "file://") {
		return maptContext.BackedURL()
	}
	return fmt.Sprintf("%s/%s", maptContext.BackedURL(), maptContext.RunID())
}

// Get all dedicated hosts matching the tags + arch
// it will return the list ordered by allocation time
func getMatchingHostsInformation(arch string) ([]*HostInformation, error) {
	return getMatchingHostsInStateInformation(arch, nil)
}

// Get all dedicated hosts in available state ordered based on the allocation time
// newer allocations go first
// func getMatchingAvailableHostsInformation(arch string) ([]HostInformation, error) {
// 	as := ec2Types.AllocationStateAvailable
// 	return getMatchingHostsInStateInformation(arch, &as)
// }

// Get all dedicated hosts by tag and state
func getMatchingHostsInStateInformation(arch string, state *ec2Types.AllocationState) ([]*HostInformation, error) {
	matchingTags := maptContext.GetTags()
	matchingTags[tagKeyArch] = arch
	hosts, err := data.GetDedicatedHosts(data.DedicatedHostResquest{
		Tags: matchingTags,
	})
	if err != nil {
		return nil, err
	}
	var r []*HostInformation
	for _, dh := range hosts {
		if state == nil || (state != nil && dh.State == *state) {
			r = append(r, getHostInformation(dh))
		}
	}
	// Order by allocation time, first newest
	if len(r) > 1 {
		slices.SortFunc(r, func(a, b *HostInformation) int {
			return b.Host.AllocationTime.Compare(*a.Host.AllocationTime)
		})
	}
	return r, nil
}

// checks if the machine can be created on the current location (machine type is available on the region)
// if it available it returns the region name
// if not offered and machine should be created on the region it will return an error
// if not offered and machine could be created anywhere it will get a region offering the machine and return its name
func getRegion(r *MacRequest) (*string, error) {
	logging.Debugf("checking if %s is offered at %s",
		r.Architecture,
		os.Getenv("AWS_DEFAULT_REGION"))
	isOffered, err := data.IsInstanceTypeOfferedByRegion(
		macTypesByArch[r.Architecture],
		os.Getenv("AWS_DEFAULT_REGION"))
	if err != nil {
		return nil, err
	}
	if isOffered {
		logging.Debugf("%s offers it", os.Getenv("AWS_DEFAULT_REGION"))
		region := os.Getenv("AWS_DEFAULT_REGION")
		return &region, nil
	}
	if !isOffered && r.FixedLocation {
		return nil, fmt.Errorf("the requested mac %s is not available at the current region %s and the fixed-location flag has been set",
			r.Architecture,
			os.Getenv("AWS_DEFAULT_REGION"))
	}
	// We look for a region offering the type of instance
	logging.Debugf("%s is not offered, a new region offering it will be used instead", os.Getenv("AWS_DEFAULT_REGION"))
	return data.LokupRegionOfferingInstanceType(
		macTypesByArch[r.Architecture])
}

// Get a random AZ from the requested region, it ensures the az offers the instance type
func getAZ(r *MacRequest) (az *string, err error) {
	isOffered := false
	var excludedAZs []string
	for !isOffered {
		az, err = data.GetRandomAvailabilityZone(*r.Region, excludedAZs)
		if err != nil {
			return nil, err
		}
		isOffered, err = data.IsInstanceTypeOfferedByAZ(*r.Region, macTypesByArch[r.Architecture], *az)
		if err != nil {
			return nil, err
		}
		if !isOffered {
			excludedAZs = append(excludedAZs, *az)
		}
	}
	return
}
