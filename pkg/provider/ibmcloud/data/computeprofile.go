package data

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector { return &ComputeSelector{} }

// Select returns the smallest IBM Cloud VPC instance profile name satisfying the
// requested CPUs, memory, and optional family restrictions. The profile must
// support the amd64 architecture.
func (c *ComputeSelector) Select(args *computerequest.ComputeRequestArgs) ([]string, error) {
	svc, err := vpcService()
	if err != nil {
		return nil, err
	}

	var all []vpcv1.InstanceProfile
	var start *string
	for {
		opts := &vpcv1.ListInstanceProfilesOptions{}
		if start != nil {
			opts.SetStart(*start)
		}
		result, _, err := svc.ListInstanceProfiles(opts)
		if err != nil {
			return nil, fmt.Errorf("listing instance profiles: %w", err)
		}
		all = append(all, result.Profiles...)
		if result.Next == nil {
			break
		}
		next, err := result.GetNextStart()
		if err != nil {
			return nil, fmt.Errorf("paginating instance profiles: %w", err)
		}
		if next == nil {
			break
		}
		start = next
	}

	var matched []vpcv1.InstanceProfile
	for _, p := range all {
		if !matchesArch(p, "amd64") {
			continue
		}
		if args.CPUs > 0 {
			vcpus, ok := vcpuCount(p)
			if !ok || vcpus < int64(args.CPUs) {
				continue
			}
		}
		if args.MemoryGib > 0 {
			mem, ok := memoryGiB(p)
			if !ok || mem < int64(args.MemoryGib) {
				continue
			}
		}
		if len(args.ComputeFamilies) > 0 && !matchesFamilies(p, args.ComputeFamilies) {
			continue
		}
		matched = append(matched, p)
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no IBM Cloud VPC profile found for CPUs>=%d MemoryGiB>=%d", args.CPUs, args.MemoryGib)
	}

	sort.Slice(matched, func(i, j int) bool {
		ci, _ := vcpuCount(matched[i])
		cj, _ := vcpuCount(matched[j])
		if ci != cj {
			return ci < cj
		}
		mi, _ := memoryGiB(matched[i])
		mj, _ := memoryGiB(matched[j])
		return mi < mj
	})

	return []string{*matched[0].Name}, nil
}

// vcpuCount returns the minimum vCPU count for the profile.
//
// The IBM Cloud VPC SDK unmarshals VcpuCount as *InstanceProfileVcpu (the base
// type) regardless of the profile field variant (fixed/range/enum). The Type
// field discriminates which fields are populated.
func vcpuCount(p vpcv1.InstanceProfile) (int64, bool) {
	v, ok := p.VcpuCount.(*vpcv1.InstanceProfileVcpu)
	if !ok || v == nil || v.Type == nil {
		return 0, false
	}
	switch *v.Type {
	case vpcv1.InstanceProfileVcpuTypeFixedConst:
		if v.Value != nil {
			return *v.Value, true
		}
	case vpcv1.InstanceProfileVcpuRangeTypeRangeConst:
		if v.Min != nil {
			return *v.Min, true
		}
	case vpcv1.InstanceProfileVcpuEnumTypeEnumConst:
		if len(v.Values) > 0 {
			return slices.Min(v.Values), true
		}
	}
	return 0, false
}

// memoryGiB returns the minimum memory in GiB for the profile.
//
// Same pattern as vcpuCount: the SDK always returns *InstanceProfileMemory.
func memoryGiB(p vpcv1.InstanceProfile) (int64, bool) {
	v, ok := p.Memory.(*vpcv1.InstanceProfileMemory)
	if !ok || v == nil || v.Type == nil {
		return 0, false
	}
	switch *v.Type {
	case vpcv1.InstanceProfileMemoryTypeFixedConst:
		if v.Value != nil {
			return *v.Value, true
		}
	case vpcv1.InstanceProfileMemoryRangeTypeRangeConst:
		if v.Min != nil {
			return *v.Min, true
		}
	case vpcv1.InstanceProfileMemoryEnumTypeEnumConst:
		if len(v.Values) > 0 {
			return slices.Min(v.Values), true
		}
	}
	return 0, false
}

// matchesArch checks VcpuArchitecture.Value which is "amd64" or "s390x".
func matchesArch(p vpcv1.InstanceProfile, arch string) bool {
	if p.VcpuArchitecture == nil || p.VcpuArchitecture.Value == nil {
		return false
	}
	return *p.VcpuArchitecture.Value == arch
}

// ProfileSupportsSpot returns true when the named profile permits spot
// (preemptible) instances. Profiles with a fixed availability class of
// "standard" only cannot be used with spot.
func ProfileSupportsSpot(name string) (bool, error) {
	svc, err := vpcService()
	if err != nil {
		return false, err
	}
	result, _, err := svc.GetInstanceProfile(&vpcv1.GetInstanceProfileOptions{Name: &name})
	if err != nil {
		return false, fmt.Errorf("getting instance profile %q: %w", name, err)
	}
	ac, ok := result.AvailabilityClass.(*vpcv1.InstanceProfileAvailabilityClass)
	if !ok || ac == nil || ac.Type == nil {
		return false, nil
	}
	switch *ac.Type {
	case vpcv1.InstanceProfileAvailabilityClassTypeEnumConst:
		for _, v := range ac.Values {
			if v == vpcv1.InstanceProfileAvailabilityClassValuesSpotConst {
				return true, nil
			}
		}
	default: // fixed
		return ac.Value != nil && *ac.Value == vpcv1.InstanceProfileAvailabilityClassValueSpotConst, nil
	}
	return false, nil
}

func matchesFamilies(p vpcv1.InstanceProfile, families []string) bool {
	if p.Family == nil {
		return false
	}
	for _, f := range families {
		if strings.HasPrefix(*p.Family, f) {
			return true
		}
	}
	return false
}
