package data

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	// standard D,E and F series are the VM families
	// supporting nested virtualization
	excludedFamilies = "Av4Family,ASv4Family,CSv3Family"
	dSeriesPattern   = `^[Ss]tandardD.*v[3-6]Family$`
	eSeriesPattern   = `^[Ss]tandardE.*v[3-6]Family$`
	fSeriesPattern   = `^[Ss]tandardF.*v\dFamily$`
	//
	lowerCpuPattern = `^[Ss]tandard.*-.*_v\d$`
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector { return &ComputeSelector{} }

func (c *ComputeSelector) Select(ctx context.Context, args *cr.ComputeRequestArgs) ([]string, error) {
	return getAzureVMSKUs(ctx, args)
}

// FilterComputeSizesByDiskControllerType returns the subset of computeSizes that are
// available in location AND support requiredType. Sizes without a DiskControllerTypes
// capability are assumed to support only SCSI (Azure historical default).
func FilterComputeSizesByDiskControllerType(ctx context.Context, location *string, computeSizes []string, requiredType string) ([]string, error) {
	if location == nil {
		return nil, fmt.Errorf("location cannot be nil")
	}
	creds, subscriptionID, err := getCredentials()
	if err != nil {
		return nil, err
	}
	client, err := armcompute.NewResourceSKUsClient(*subscriptionID, creds, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	supported := []string{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sku := range page.Value {
			if sku.ResourceType == nil || *sku.ResourceType != string(RTVirtualMachines) {
				continue
			}
			if sku.Name == nil || !slices.Contains(computeSizes, *sku.Name) {
				continue
			}
			inLocation := false
			for _, loc := range sku.Locations {
				if loc != nil && strings.EqualFold(*loc, *location) {
					inLocation = true
					break
				}
			}
			if !inLocation {
				continue
			}
			diskTypes := diskControllerTypesFromCapabilities(sku.Capabilities)
			if diskControllerTypeSupported(diskTypes, requiredType) && !slices.Contains(supported, *sku.Name) {
				supported = append(supported, *sku.Name)
			}
		}
	}
	return supported, nil
}

// diskControllerTypesFromCapabilities extracts the DiskControllerTypes value from SKU
// capabilities. Returns nil when the capability is absent.
func diskControllerTypesFromCapabilities(caps []*armcompute.ResourceSKUCapabilities) []string {
	for _, c := range caps {
		if c.Name != nil && *c.Name == "DiskControllerTypes" && c.Value != nil {
			return splitDiskControllerTypes(*c.Value)
		}
	}
	return nil
}

// diskControllerTypeSupported reports whether requiredType is satisfied by the supported
// set. Empty requiredType means no restriction (always passes). A nil/empty supported
// set means the capability is absent; Azure sizes that predate NVMe default to SCSI, so
// absence is treated as SCSI-only.
func diskControllerTypeSupported(supported []string, requiredType string) bool {
	if requiredType == "" {
		return true
	}
	if len(supported) == 0 {
		return strings.EqualFold(requiredType, "SCSI")
	}
	for _, t := range supported {
		if strings.EqualFold(t, requiredType) {
			return true
		}
	}
	return false
}

// FilterNoLocalStorageSizes returns only the sizes from computeSizes that have no
// NVMe-only local storage (L-series). Temp disks (MaxResourceVolumeMB > 0) are allowed
// — they are ephemeral scratch space that does not interfere with RHEL AI's OS disk.
// Sizes not found in the Azure SKU catalog (typo or restricted SKU) are logged as
// warnings and excluded.
func FilterNoLocalStorageSizes(ctx context.Context, computeSizes []string) ([]string, error) {
	creds, subscriptionID, err := getCredentials()
	if err != nil {
		return nil, err
	}
	client, err := armcompute.NewResourceSKUsClient(*subscriptionID, creds, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	capabilities := make(map[string]*virtualMachine, len(computeSizes))
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sku := range page.Value {
			if sku.ResourceType == nil || *sku.ResourceType != string(RTVirtualMachines) {
				continue
			}
			if sku.Name == nil || !slices.Contains(computeSizes, *sku.Name) {
				continue
			}
			if _, seen := capabilities[*sku.Name]; seen {
				continue
			}
			if vm := resourceSKUToVirtualMachine(sku); vm != nil {
				capabilities[*sku.Name] = vm
			}
		}
	}
	valid, dropped, unknown := filterNVMeStorage(computeSizes, capabilities)
	for _, size := range dropped {
		logging.Warnf("dropping compute size %q: has NVMe-only local storage, incompatible with RHEL AI", size)
	}
	for _, size := range unknown {
		logging.Warnf("dropping compute size %q: not found in Azure SKU catalog (typo or restricted SKU)", size)
	}
	return valid, nil
}

// filterNVMeStorage classifies each size into valid (no NVMe-only local storage),
// dropped (has NVMe local storage — e.g. L-series), or unknown (absent from capabilities).
func filterNVMeStorage(computeSizes []string, capabilities map[string]*virtualMachine) (valid, dropped, unknown []string) {
	for _, size := range computeSizes {
		vm, ok := capabilities[size]
		if !ok {
			unknown = append(unknown, size)
			continue
		}
		if vm.NvmeDiskSizeInMiB > 0 {
			dropped = append(dropped, size)
		} else {
			valid = append(valid, size)
		}
	}
	return valid, dropped, unknown
}

func getAzureVMSKUs(ctx context.Context, args *cr.ComputeRequestArgs) ([]string, error) {
	ensureAzureEnvs()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	subscriptionId := SubscriptionID()
	clientFactory, err := armcompute.NewClientFactory(
		subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	pager := clientFactory.NewResourceSKUsClient().NewListPager(
		&armcompute.ResourceSKUsClientListOptions{
			Filter:                   nil,
			IncludeExtendedLocations: nil,
		})
	vmTypes := []string{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		vmTypes = append(vmTypes,
			filterVMs(ctx, page,
				filterCPUsAndMemory(args))...)
	}
	return vmTypes, nil
}

type filterFunc func(context.Context, *virtualMachine, *sync.WaitGroup, chan<- string)

type virtualMachine struct {
	Name         string
	Family       string
	VCPUs        int32
	VCPUsPerCore int32
	Memory       int32
	// Hyperv gen1 or gen2
	HyperVGenerations []string
	Arch              string
	// Spot capable
	LowPriorityCapable  bool
	MaxResourceVolumeMB int32
	GPUs                int32
	// L-series VMs expose NVMe storage separately from the temp disk
	NvmeDiskSizeInMiB int32
	// Used by the disk-controller-type fix (PR #823) to cross-reference SKU capabilities
	DiskControllerTypes []string
	// IaaS or PaaS
	VMDeploymentTypes []string
	// Fast SSD
	PremiumIO                    bool
	AcceleratedNetworkingEnabled bool
	EncryptionAtHostSupported    bool
}

func (vm *virtualMachine) nestedVirtSupported() bool {
	dSeries := regexp.MustCompile(dSeriesPattern)
	eSeries := regexp.MustCompile(eSeriesPattern)
	fSeries := regexp.MustCompile(fSeriesPattern)

	if (dSeries.Match([]byte(vm.Family)) ||
		eSeries.Match([]byte(vm.Family)) ||
		fSeries.Match([]byte(vm.Family))) && !isExcludedFamily(vm.Family) {
		return true
	}
	return false
}

func isExcludedFamily(family string) bool {
	excluded := strings.Split(excludedFamilies, ",")
	for _, ex := range excluded {
		if strings.HasSuffix(family, strings.TrimSpace(ex)) {
			return true
		}
	}
	return false
}

func (vm *virtualMachine) hypervGen2Supported() bool {
	return slices.Contains(vm.HyperVGenerations, "V2")
}

func (vm *virtualMachine) noLocalStorageAttached() bool {
	return vm.MaxResourceVolumeMB == 0 && vm.NvmeDiskSizeInMiB == 0
}

func (vm *virtualMachine) baseFeaturesSupported() bool {
	return vm.AcceleratedNetworkingEnabled && vm.PremiumIO && vm.EncryptionAtHostSupported &&
		vm.noLocalStorageAttached() && vm.hypervGen2Supported()
}

func resourceSKUToVirtualMachine(res *armcompute.ResourceSKU) *virtualMachine {
	if res.ResourceType != nil && *res.ResourceType != string(RTVirtualMachines) {
		return nil
	}
	// If Machine type has any type of restriccions discard
	if len(res.Restrictions) > 0 {
		return nil
	}
	vm := &virtualMachine{
		Name:   *res.Name,
		Family: *res.Family,
	}
	for _, capability := range res.Capabilities {
		switch *capability.Name {
		case "vCPUs":
			vCpus, err := strconv.ParseInt(*capability.Value, 10, 32)
			if err != nil {
				return nil
			}
			vm.VCPUs = int32(vCpus)
		case "vCPUsPerCore":
			vCpusPerCore, err := strconv.ParseInt(*capability.Value, 10, 32)
			if err != nil {
				return nil
			}
			vm.VCPUsPerCore = int32(vCpusPerCore)
		case "MemoryGB":
			mem, err := strconv.ParseInt(*capability.Value, 10, 32)
			if err != nil {
				return nil
			}
			vm.Memory = int32(mem)
		case "HyperVGenerations":
			vm.HyperVGenerations = strings.Split(*capability.Value, ",")
		case "AcceleratedNetworkingEnabled":
			fastNet, err := strconv.ParseBool(*capability.Value)
			if err != nil {
				return nil
			}
			vm.AcceleratedNetworkingEnabled = fastNet
		case "EncryptionAtHostSupported":
			encryption, err := strconv.ParseBool(*capability.Value)
			if err != nil {
				return nil
			}
			vm.EncryptionAtHostSupported = encryption
		case "CpuArchitectureType":
			vm.Arch = *capability.Value
		case "LowPriorityCapable":
			lowPriority, err := strconv.ParseBool(*capability.Value)
			if err != nil {
				return nil
			}
			vm.LowPriorityCapable = lowPriority
		case "PremiumIO":
			premiumIO, err := strconv.ParseBool(*capability.Value)
			if err != nil {
				return nil
			}
			vm.PremiumIO = premiumIO
		case "MaxResourceVolumeMB":
			disk, err := strconv.ParseUint(*capability.Value, 10, 64)
			if err != nil {
				return nil
			}
			vm.MaxResourceVolumeMB = int32(disk)
		case "GPUs":
			gpus, err := strconv.ParseInt(*capability.Value, 10, 32)
			if err != nil {
				return nil
			}
			vm.GPUs = int32(gpus)
		case "NvmeDiskSizeInMiB":
			nvme, err := strconv.ParseUint(*capability.Value, 10, 32)
			if err != nil {
				return nil
			}
			vm.NvmeDiskSizeInMiB = int32(nvme)
		case "DiskControllerTypes":
			vm.DiskControllerTypes = strings.Split(*capability.Value, ",")
		case "VMDeploymentTypes":
			vm.VMDeploymentTypes = strings.Split(*capability.Value, ",")
		default:
			continue
		}
	}
	return vm
}

func filterCPUsAndMemory(args *cr.ComputeRequestArgs) filterFunc {
	return func(ctx context.Context, vm *virtualMachine, wg *sync.WaitGroup, vmCh chan<- string) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			if vm == nil {
				return
			}
			if args.NestedVirt && !vm.nestedVirtSupported() {
				return
			}
			if args.GPUs > 0 && vm.GPUs < args.GPUs {
				return
			}
			// GPU VMs (ND/NC-series) have large temp disks, so skip the
			// local-storage check that would otherwise reject them.
			featuresOK := false
			if args.GPUs > 0 {
				featuresOK = vm.AcceleratedNetworkingEnabled && vm.PremiumIO &&
					vm.EncryptionAtHostSupported && vm.hypervGen2Supported()
			} else {
				featuresOK = vm.baseFeaturesSupported()
			}
			if vm.VCPUs >= args.CPUs &&
				vm.Memory >= args.MemoryGib &&
				vm.Arch == args.Arch.String() &&
				featuresOK {
				dSeries := regexp.MustCompile(lowerCpuPattern)
				if !dSeries.Match([]byte(vm.Name)) {
					vmCh <- vm.Name
				}
			}
		}
	}
}

// sort the VirtualMachine slice based on vcpus
// for the above to happen need to have a slice of VirtualMachines in memory first
// so no go routines needed
func filterVMs(ctx context.Context, skus armcompute.ResourceSKUsClientListResponse, filter filterFunc) []string {
	chVmTypes := make(chan string, cr.MaxResults)
	vmTypes := []string{}
	virtualMachines := []*virtualMachine{}
	wg := &sync.WaitGroup{}
	childCtx, cancelFn := context.WithCancel(ctx)

	for _, v := range skus.Value {
		vm := resourceSKUToVirtualMachine(v)
		if vm != nil {
			virtualMachines = append(virtualMachines, vm)
		}
	}

	slices.SortStableFunc(virtualMachines, func(vm1, vm2 *virtualMachine) int {
		if vm1.VCPUs > vm2.VCPUs {
			return 1
		}
		if vm1.VCPUs < vm2.VCPUs {
			return -1
		}
		return 0
	})

	for _, v := range virtualMachines {
		wg.Add(1)
		go filter(childCtx, v, wg, chVmTypes)
	}
	c := make(chan int)

	go func() {
		defer close(c)
		wg.Wait()
	}()

	for {
		select {
		case vmtype := <-chVmTypes:
			if !slices.Contains(vmTypes, vmtype) {
				vmTypes = append(vmTypes, vmtype)
			}
			if len(vmTypes) == cr.MaxResults {
				cancelFn()
				return vmTypes
			}
		case <-c:
			cancelFn()
			return vmTypes
		}
	}
}
