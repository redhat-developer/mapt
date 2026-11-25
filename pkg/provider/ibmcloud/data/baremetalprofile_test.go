package data

import (
	"testing"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

func bareMetalProfile(name, family, arch string, cores, memory int64) vpcv1.BareMetalServerProfile {
	coreType := vpcv1.BareMetalServerProfileCpuCoreCountFixedTypeFixedConst
	memoryType := vpcv1.BareMetalServerProfileMemoryFixedTypeFixedConst
	return vpcv1.BareMetalServerProfile{
		Name:   &name,
		Family: &family,
		CpuArchitecture: &vpcv1.BareMetalServerProfileCpuArchitecture{
			Type:  strPtr(vpcv1.BareMetalServerProfileCpuArchitectureTypeFixedConst),
			Value: &arch,
		},
		CpuCoreCount: &vpcv1.BareMetalServerProfileCpuCoreCount{
			Type:  &coreType,
			Value: &cores,
		},
		Memory: &vpcv1.BareMetalServerProfileMemory{
			Type:  &memoryType,
			Value: &memory,
		},
	}
}

func strPtr(s string) *string { return &s }

func TestFilterBareMetalProfilesSelectsBySize(t *testing.T) {
	profiles := []vpcv1.BareMetalServerProfile{
		bareMetalProfile("bx2-metal-96x384", "balanced", "amd64", 96, 384),
		bareMetalProfile("mx2-metal-96x768", "memory", "amd64", 96, 768),
	}
	matched := filterBareMetalProfiles(profiles, &computerequest.ComputeRequestArgs{
		CPUs:      96,
		MemoryGib: 512,
		Arch:      computerequest.Amd64,
	})
	if len(matched) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(matched))
	}
	if *matched[0].Name != "mx2-metal-96x768" {
		t.Errorf("unexpected profile: %s", *matched[0].Name)
	}
}

func TestFilterBareMetalProfilesRejectsVirtualServerArchitecture(t *testing.T) {
	profiles := []vpcv1.BareMetalServerProfile{
		bareMetalProfile("bx2-metal-96x384", "balanced", "amd64", 96, 384),
	}
	matched := filterBareMetalProfiles(profiles, &computerequest.ComputeRequestArgs{
		CPUs:      4,
		MemoryGib: 16,
		Arch:      computerequest.Arm64,
	})
	if len(matched) != 0 {
		t.Fatalf("expected no profile for arm64, got %d", len(matched))
	}
}

func TestFilterBareMetalProfilesMatchesNameAndFamilyPrefix(t *testing.T) {
	profiles := []vpcv1.BareMetalServerProfile{
		bareMetalProfile("bx2-metal-96x384", "balanced", "amd64", 96, 384),
		bareMetalProfile("cx2-metal-96x192", "compute", "amd64", 96, 192),
	}
	for _, families := range [][]string{{"cx2"}, {"compute"}, {"CX2"}} {
		matched := filterBareMetalProfiles(profiles, &computerequest.ComputeRequestArgs{
			CPUs:            96,
			MemoryGib:       192,
			Arch:            computerequest.Amd64,
			ComputeFamilies: families,
		})
		if len(matched) != 1 || *matched[0].Name != "cx2-metal-96x192" {
			t.Errorf("families %v: unexpected result %v", families, matched)
		}
	}
}

func TestFilterBareMetalProfilesRejectsGPURequests(t *testing.T) {
	profiles := []vpcv1.BareMetalServerProfile{
		bareMetalProfile("bx2-metal-96x384", "balanced", "amd64", 96, 384),
	}
	matched := filterBareMetalProfiles(profiles, &computerequest.ComputeRequestArgs{
		CPUs:            96,
		MemoryGib:       384,
		Arch:            computerequest.Amd64,
		GPUs:            1,
		GPUManufacturer: "NVIDIA",
	})
	if len(matched) != 0 {
		t.Fatalf("expected no profile for a GPU request, got %d", len(matched))
	}
}

func TestBareMetalCoreCountVariants(t *testing.T) {
	fixed := vpcv1.BareMetalServerProfileCpuCoreCountFixedTypeFixedConst
	enum := vpcv1.BareMetalServerProfileCpuCoreCountEnumTypeEnumConst
	value := int64(48)
	profile := bareMetalProfile("cx2-metal-96x192", "compute", "amd64", 96, 192)
	if cores, ok := bareMetalCoreCount(profile); !ok || cores != 96 {
		t.Errorf("fixed core count: got %d ok=%v", cores, ok)
	}

	profile.CpuCoreCount = &vpcv1.BareMetalServerProfileCpuCoreCount{
		Type:   &enum,
		Values: []int64{16, 32},
	}
	if cores, ok := bareMetalCoreCount(profile); !ok || cores != 16 {
		t.Errorf("enum core count: got %d ok=%v", cores, ok)
	}

	profile.CpuCoreCount = &vpcv1.BareMetalServerProfileCpuCoreCount{
		Type:  &fixed,
		Value: &value,
	}
	if cores, ok := bareMetalCoreCount(profile); !ok || cores != value {
		t.Errorf("fixed core count after reset: got %d ok=%v", cores, ok)
	}
}

func TestBareMetalMemoryGiBVariants(t *testing.T) {
	rng := vpcv1.BareMetalServerProfileMemoryRangeTypeRangeConst
	minimum := int64(256)
	profile := bareMetalProfile("cx2-metal-96x192", "compute", "amd64", 96, 192)
	if mem, ok := bareMetalMemoryGiB(profile); !ok || mem != 192 {
		t.Errorf("fixed memory: got %d ok=%v", mem, ok)
	}

	profile.Memory = &vpcv1.BareMetalServerProfileMemory{
		Type: &rng,
		Min:  &minimum,
	}
	if mem, ok := bareMetalMemoryGiB(profile); !ok || mem != 256 {
		t.Errorf("range memory: got %d ok=%v", mem, ok)
	}
}
