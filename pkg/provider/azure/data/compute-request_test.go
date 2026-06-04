package data

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

func ptr[T any](v T) *T { return &v }

// noLocalStorageAttached tests

func TestNoLocalStorageAttached_NoTempDiskNoNvme(t *testing.T) {
	vm := &virtualMachine{MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 0}
	if !vm.noLocalStorageAttached() {
		t.Error("expected true: VM with no temp disk and no NVMe should have no local storage")
	}
}

func TestNoLocalStorageAttached_HasTempDisk(t *testing.T) {
	vm := &virtualMachine{MaxResourceVolumeMB: 512, NvmeDiskSizeInMiB: 0}
	if vm.noLocalStorageAttached() {
		t.Error("expected false: VM with temp disk should have local storage")
	}
}

func TestNoLocalStorageAttached_HasNvmeDisk(t *testing.T) {
	// L-series bug case: MaxResourceVolumeMB=0 but NvmeDiskSizeInMiB>0
	vm := &virtualMachine{MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 5492736}
	if vm.noLocalStorageAttached() {
		t.Error("expected false: L-series VM with NVMe storage should have local storage")
	}
}

func TestNoLocalStorageAttached_HasBoth(t *testing.T) {
	vm := &virtualMachine{MaxResourceVolumeMB: 512, NvmeDiskSizeInMiB: 5492736}
	if vm.noLocalStorageAttached() {
		t.Error("expected false: VM with both temp disk and NVMe should have local storage")
	}
}

// resourceSKUToVirtualMachine parsing tests

func TestResourceSKUToVirtualMachine_ParsesNvmeDiskSizeInMiB(t *testing.T) {
	sku := &armcompute.ResourceSKU{
		ResourceType: ptr("virtualMachines"),
		Name:         ptr("Standard_L8aos_v4"),
		Family:       ptr("standardLasv4Family"),
		Capabilities: []*armcompute.ResourceSKUCapabilities{
			{Name: ptr("NvmeDiskSizeInMiB"), Value: ptr("5492736")},
		},
	}
	vm := resourceSKUToVirtualMachine(sku)
	if vm == nil {
		t.Fatal("expected non-nil virtualMachine")
	}
	if vm.NvmeDiskSizeInMiB != 5492736 {
		t.Errorf("NvmeDiskSizeInMiB: got %d, want 5492736", vm.NvmeDiskSizeInMiB)
	}
}

func TestResourceSKUToVirtualMachine_ParsesDiskControllerTypes(t *testing.T) {
	sku := &armcompute.ResourceSKU{
		ResourceType: ptr("virtualMachines"),
		Name:         ptr("Standard_L8aos_v4"),
		Family:       ptr("standardLasv4Family"),
		Capabilities: []*armcompute.ResourceSKUCapabilities{
			{Name: ptr("DiskControllerTypes"), Value: ptr("NVMe,SCSI")},
		},
	}
	vm := resourceSKUToVirtualMachine(sku)
	if vm == nil {
		t.Fatal("expected non-nil virtualMachine")
	}
	if len(vm.DiskControllerTypes) != 2 {
		t.Fatalf("DiskControllerTypes: got %v, want [NVMe SCSI]", vm.DiskControllerTypes)
	}
	if vm.DiskControllerTypes[0] != "NVMe" || vm.DiskControllerTypes[1] != "SCSI" {
		t.Errorf("DiskControllerTypes: got %v, want [NVMe SCSI]", vm.DiskControllerTypes)
	}
}

func TestResourceSKUToVirtualMachine_NvmeDiskSizeDefaultsToZero(t *testing.T) {
	sku := &armcompute.ResourceSKU{
		ResourceType: ptr("virtualMachines"),
		Name:         ptr("Standard_D8as_v5"),
		Family:       ptr("standardDasv5Family"),
		Capabilities: []*armcompute.ResourceSKUCapabilities{
			{Name: ptr("MaxResourceVolumeMB"), Value: ptr("307200")},
		},
	}
	vm := resourceSKUToVirtualMachine(sku)
	if vm == nil {
		t.Fatal("expected non-nil virtualMachine")
	}
	if vm.NvmeDiskSizeInMiB != 0 {
		t.Errorf("NvmeDiskSizeInMiB: got %d, want 0 for non-NVMe SKU", vm.NvmeDiskSizeInMiB)
	}
}

// filterNVMeStorage tests

func TestFilterNVMeStorage_DropsNvmeSizes(t *testing.T) {
	capabilities := map[string]*virtualMachine{
		"Standard_D8as_v5":  {MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 0},
		"Standard_L8aos_v4": {MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 5492736},
	}
	got, dropped, unknown := filterNVMeStorage([]string{"Standard_D8as_v5", "Standard_L8aos_v4"}, capabilities)
	if len(got) != 1 || got[0] != "Standard_D8as_v5" {
		t.Errorf("filtered: got %v, want [Standard_D8as_v5]", got)
	}
	if len(dropped) != 1 || dropped[0] != "Standard_L8aos_v4" {
		t.Errorf("dropped: got %v, want [Standard_L8aos_v4]", dropped)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown: got %v, want []", unknown)
	}
}

func TestFilterNVMeStorage_AllowsTempDiskSizes(t *testing.T) {
	capabilities := map[string]*virtualMachine{
		"Standard_NC64as_T4_v3": {MaxResourceVolumeMB: 32768, NvmeDiskSizeInMiB: 0},
	}
	got, dropped, unknown := filterNVMeStorage([]string{"Standard_NC64as_T4_v3"}, capabilities)
	if len(got) != 1 {
		t.Errorf("filtered: got %v, want [Standard_NC64as_T4_v3]", got)
	}
	if len(dropped) != 0 {
		t.Errorf("dropped: got %v, want []", dropped)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown: got %v, want []", unknown)
	}
}

func TestFilterNVMeStorage_PassesCleanSizes(t *testing.T) {
	capabilities := map[string]*virtualMachine{
		"Standard_D8as_v5": {MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 0},
	}
	got, dropped, unknown := filterNVMeStorage([]string{"Standard_D8as_v5"}, capabilities)
	if len(got) != 1 {
		t.Errorf("filtered: got %v, want [Standard_D8as_v5]", got)
	}
	if len(dropped) != 0 {
		t.Errorf("dropped: got %v, want []", dropped)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown: got %v, want []", unknown)
	}
}

func TestFilterNVMeStorage_ReportsUnknownSizes(t *testing.T) {
	capabilities := map[string]*virtualMachine{
		"Standard_D8as_v5": {MaxResourceVolumeMB: 0, NvmeDiskSizeInMiB: 0},
	}
	got, dropped, unknown := filterNVMeStorage(
		[]string{"Standard_D8as_v5", "Standard_Typo_v99"}, capabilities)
	if len(got) != 1 || got[0] != "Standard_D8as_v5" {
		t.Errorf("filtered: got %v, want [Standard_D8as_v5]", got)
	}
	if len(dropped) != 0 {
		t.Errorf("dropped: got %v, want []", dropped)
	}
	if len(unknown) != 1 || unknown[0] != "Standard_Typo_v99" {
		t.Errorf("unknown: got %v, want [Standard_Typo_v99]", unknown)
	}
}

// baseFeaturesSupported regression: L-series must be rejected

func TestBaseFeaturesSupported_LSeriesWithNvmeIsRejected(t *testing.T) {
	vm := &virtualMachine{
		MaxResourceVolumeMB:          0,
		NvmeDiskSizeInMiB:            5492736,
		AcceleratedNetworkingEnabled: true,
		PremiumIO:                    true,
		EncryptionAtHostSupported:    true,
		HyperVGenerations:            []string{"V1", "V2"},
	}
	if vm.baseFeaturesSupported() {
		t.Error("expected false: L-series VM with NVMe storage must not pass baseFeaturesSupported")
	}
}
