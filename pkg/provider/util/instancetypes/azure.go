package instancetypes

import (
	"context"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

func getAzureVMSKUs(cpus, memory int32, arch arch, nestedVirt bool) ([]string, error) {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientFactory, err := armcompute.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	pager := clientFactory.NewResourceSKUsClient().NewListPager(&armcompute.ResourceSKUsClientListOptions{Filter: nil,
		IncludeExtendedLocations: nil,
	})

	vmTypes := []string{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		vmTypes = append(vmTypes, FilterVMs(page, filterCPUsAndMemory(cpus, memory, arch, nestedVirt))...)
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
	// IaaS or PaaS
	VMDeploymentTypes []string
	// Fast SSD
	PremiumIO                    bool
	AcceleratedNetworkingEnabled bool
	EncryptionAtHostSupported    bool
}

func (vm *virtualMachine) nestedVirtSupported() bool {
	// standard D,E and F series are the VM families
	// supporting nested virtualization
	var dSeriesPattern = `standardD.*v[3-6]Family$`
	var eSeriesPattern = `standardE.*v[3-6]Family$`
	var fSeriesPattern = `standardF.*v\dFamily$`

	dSeries := regexp.MustCompile(dSeriesPattern)
	if dSeries.Match([]byte(vm.Family)) {
		return true
	}

	eSeries := regexp.MustCompile(eSeriesPattern)
	if eSeries.Match([]byte(vm.Family)) {
		return true
	}

	fSeries := regexp.MustCompile(fSeriesPattern)

	return fSeries.Match([]byte(vm.Family))
}

func (vm *virtualMachine) hypervGen2Supported() bool {
	return slices.Contains(vm.HyperVGenerations, "V2")
}

func (vm *virtualMachine) emptyDiskSupported() bool {
	return vm.MaxResourceVolumeMB == 0
}

func (vm *virtualMachine) baseFeaturesSupported() bool {
	return vm.AcceleratedNetworkingEnabled && vm.PremiumIO && vm.EncryptionAtHostSupported &&
		vm.emptyDiskSupported() && vm.hypervGen2Supported()
}

func resourceSKUToVirtualMachine(res *armcompute.ResourceSKU) *virtualMachine {
	if res.ResourceType != nil && *res.ResourceType != "virtualMachines" {
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
		case "VMDeploymentTypes":
			vm.VMDeploymentTypes = strings.Split(*capability.Value, ",")
		default:
			continue
		}
	}
	return vm
}

func filterCPUsAndMemory(cpus, memory int32, arch arch, nestedVirt bool) filterFunc {
	return func(ctx context.Context, vm *virtualMachine, wg *sync.WaitGroup, vmCh chan<- string) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			if vm == nil {
				return
			}
			if nestedVirt && !vm.nestedVirtSupported() {
				return
			}

			if vm.VCPUs >= cpus && vm.Memory >= memory && vm.Arch == arch.String() &&
				vm.baseFeaturesSupported() {
				vmCh <- vm.Name
			}
		}
	}
}

// sort the VirtualMachine slice based on vcpus
// for the above to happen need to have a slice of VirtualMachines in memory first
// so no go routines needed

func FilterVMs(skus armcompute.ResourceSKUsClientListResponse, filter filterFunc) []string {
	chVmTypes := make(chan string, maxResults)
	vmTypes := []string{}
	virtualMachines := []*virtualMachine{}
	wg := &sync.WaitGroup{}
	ctx, cancelFn := context.WithCancel(context.Background())

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
		go filter(ctx, v, wg, chVmTypes)
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
			if len(vmTypes) == maxResults {
				cancelFn()
				return vmTypes
			}
		case <-c:
			cancelFn()
			return vmTypes
		}
	}
}

type AzureInstanceRequest struct {
	CPUs       int32
	MemoryGib  int32
	Arch       arch
	NestedVirt bool
}

func (r *AzureInstanceRequest) GetMachineTypes() ([]string, error) {
	if err := validate(r.CPUs, r.MemoryGib, r.Arch); err != nil {
		return nil, err
	}
	return getAzureVMSKUs(r.CPUs, r.MemoryGib, r.Arch, r.NestedVirt)
}
