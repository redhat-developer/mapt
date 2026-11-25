package data

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

// BareMetalSelector selects IBM Cloud VPC bare metal server profiles.
//
// IBM Cloud provides nested virtualization (KVM) to the guest on bare metal
// servers only. A virtual server instance cannot expose /dev/kvm, so OpenShift
// Virtualization does not work on it. The selector reads the same
// ComputeRequestArgs fields as the instance selector and returns a metal
// profile.
type BareMetalSelector struct{}

func NewBareMetalSelector() *BareMetalSelector { return &BareMetalSelector{} }

// Select returns the smallest IBM Cloud VPC bare metal profile that satisfies
// the requested core count, memory size, architecture and optional family
// allowlist. The profile names come from ListBareMetalServerProfiles.
func (c *BareMetalSelector) Select(args *computerequest.ComputeRequestArgs) ([]string, error) {
	svc, err := vpcService()
	if err != nil {
		return nil, err
	}

	var all []vpcv1.BareMetalServerProfile
	var start *string
	for {
		opts := &vpcv1.ListBareMetalServerProfilesOptions{}
		if start != nil {
			opts.SetStart(*start)
		}
		result, _, err := svc.ListBareMetalServerProfiles(opts)
		if err != nil {
			return nil, fmt.Errorf("listing bare metal server profiles: %w", err)
		}
		all = append(all, result.Profiles...)
		if result.Next == nil {
			break
		}
		next, err := result.GetNextStart()
		if err != nil {
			return nil, fmt.Errorf("paginating bare metal server profiles: %w", err)
		}
		if next == nil {
			break
		}
		start = next
	}

	matched := filterBareMetalProfiles(all, args)
	if len(matched) == 0 {
		return nil, fmt.Errorf(
			"no IBM Cloud VPC bare metal profile found for CPUs>=%d MemoryGiB>=%d arch %s",
			args.CPUs, args.MemoryGib, bareMetalArch(args.Arch))
	}

	sort.Slice(matched, func(i, j int) bool {
		ci, _ := bareMetalCoreCount(matched[i])
		cj, _ := bareMetalCoreCount(matched[j])
		if ci != cj {
			return ci < cj
		}
		mi, _ := bareMetalMemoryGiB(matched[i])
		mj, _ := bareMetalMemoryGiB(matched[j])
		return mi < mj
	})

	return []string{*matched[0].Name}, nil
}

// CheckCapacity returns an error when IBM Cloud has no free bare metal server
// of the profile in the zone.
//
// IBM Cloud allocates bare metal servers from a limited pool in each zone, so
// the check turns a late create failure into a clear error before the stack
// starts.
func (c *BareMetalSelector) CheckCapacity(profile, zone string) error {
	svc, err := vpcService()
	if err != nil {
		return err
	}
	result, _, err := svc.ListBareMetalServerCapacities(
		&vpcv1.ListBareMetalServerCapacitiesOptions{
			ProfileName: &profile,
			ZoneName:    &zone,
		})
	if err != nil {
		return fmt.Errorf("listing bare metal server capacity for profile %q: %w", profile, err)
	}
	if result == nil || len(result.Capacities) == 0 {
		return fmt.Errorf("no bare metal server capacity for profile %q in zone %q", profile, zone)
	}
	return nil
}

// filterBareMetalProfiles applies the compute request fields to the list of
// bare metal profiles.
//
// The bare metal profile has no GPU field, so a request for GPUs matches no
// profile. The family allowlist matches the profile name prefix (for example
// "cx2") or the family name prefix (for example "compute").
func filterBareMetalProfiles(all []vpcv1.BareMetalServerProfile,
	args *computerequest.ComputeRequestArgs,
) []vpcv1.BareMetalServerProfile {
	if args.GPUs > 0 || args.GPUManufacturer != "" {
		return nil
	}
	wantArch := bareMetalArch(args.Arch)
	var matched []vpcv1.BareMetalServerProfile
	for _, p := range all {
		if !matchesBareMetalArch(p, wantArch) {
			continue
		}
		if args.CPUs > 0 {
			cores, ok := bareMetalCoreCount(p)
			if !ok || cores < int64(args.CPUs) {
				continue
			}
		}
		if args.MaxCPUs > 0 {
			cores, ok := bareMetalCoreCount(p)
			if !ok || cores > int64(args.MaxCPUs) {
				continue
			}
		}
		if args.MemoryGib > 0 {
			mem, ok := bareMetalMemoryGiB(p)
			if !ok || mem < int64(args.MemoryGib) {
				continue
			}
		}
		if len(args.ComputeFamilies) > 0 && !matchesBareMetalFamilies(p, args.ComputeFamilies) {
			continue
		}
		matched = append(matched, p)
	}
	return matched
}

// bareMetalArch maps the request architecture to the IBM Cloud VPC bare metal
// architecture value (amd64 or s390x).
func bareMetalArch(arch computerequest.Arch) string {
	if arch == computerequest.Arm64 {
		return "arm64"
	}
	return vpcv1.BareMetalServerProfileCpuArchitectureValueAmd64Const
}

func matchesBareMetalArch(p vpcv1.BareMetalServerProfile, arch string) bool {
	if p.CpuArchitecture == nil || p.CpuArchitecture.Value == nil {
		return false
	}
	return *p.CpuArchitecture.Value == arch
}

// bareMetalCoreCount returns the minimum core count for the profile.
//
// The SDK unmarshals CpuCoreCount as *BareMetalServerProfileCpuCoreCount (the
// base type) for all field variants. The Type field selects the populated
// fields.
func bareMetalCoreCount(p vpcv1.BareMetalServerProfile) (int64, bool) {
	v, ok := p.CpuCoreCount.(*vpcv1.BareMetalServerProfileCpuCoreCount)
	if !ok || v == nil || v.Type == nil {
		return 0, false
	}
	switch *v.Type {
	case vpcv1.BareMetalServerProfileCpuCoreCountFixedTypeFixedConst:
		if v.Value != nil {
			return *v.Value, true
		}
	case vpcv1.BareMetalServerProfileCpuCoreCountRangeTypeRangeConst:
		if v.Min != nil {
			return *v.Min, true
		}
	case vpcv1.BareMetalServerProfileCpuCoreCountEnumTypeEnumConst:
		if len(v.Values) > 0 {
			return slices.Min(v.Values), true
		}
	}
	return 0, false
}

// bareMetalMemoryGiB returns the minimum memory in GiB for the profile.
func bareMetalMemoryGiB(p vpcv1.BareMetalServerProfile) (int64, bool) {
	v, ok := p.Memory.(*vpcv1.BareMetalServerProfileMemory)
	if !ok || v == nil || v.Type == nil {
		return 0, false
	}
	switch *v.Type {
	case vpcv1.BareMetalServerProfileMemoryFixedTypeFixedConst:
		if v.Value != nil {
			return *v.Value, true
		}
	case vpcv1.BareMetalServerProfileMemoryRangeTypeRangeConst:
		if v.Min != nil {
			return *v.Min, true
		}
	case vpcv1.BareMetalServerProfileMemoryEnumTypeEnumConst:
		if len(v.Values) > 0 {
			return slices.Min(v.Values), true
		}
	}
	return 0, false
}

func matchesBareMetalFamilies(p vpcv1.BareMetalServerProfile, families []string) bool {
	for _, f := range families {
		prefix := strings.ToLower(f)
		if p.Name != nil && strings.HasPrefix(strings.ToLower(*p.Name), prefix) {
			return true
		}
		if p.Family != nil && strings.HasPrefix(strings.ToLower(*p.Family), prefix) {
			return true
		}
	}
	return false
}
