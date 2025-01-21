package macpool

import (
	"fmt"
	"strings"
	"time"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// request and release (same approach as mac standard, but request never create machine underneath, this is handled by pool-capacity-keeper)
// TODO Important
// release and request will not behave the same if the run as targer vs as selfhosted runner. In that case for release we do not
// want it be added as selfhosted but only on request??

func Create(r *RequestArgs) error {
	// Initially create pool with number of machines matching available capacity
	// this is the number of machines free to accept workloads
	// if err := validateNoExistingPool(); err != nil {
	// 	return err
	// }
	if err := r.addMachinesToPool(r.OfferedCapacity); err != nil {
		return err
	}
	if err := r.scheduleHouseKeeper(); err != nil {
		return err
	}
	return nil
}

// House keeper is the function executed serverless to check if is there any
// machine non locked which had been running more than 24h.
// It should check if capacity allows to remove the machine
func HouseKeeper(r *RequestArgs) error {
	// First get full info on the pool
	p, err := getPool(r.PoolName, r.Architecture, r.OSVersion)
	if err != nil {
		return err
	}
	// Pool under expected offered capacity
	if p.currentOfferedCapacity() < p.offeredCapacity {
		if p.currentPoolSize() < p.maxSize {
			return r.addCapacity(p)
		}
		// if number of machines in the pool + to max machines
		// we do nothing
		return nil
	}
	// Pool over expected offered capacity need to destroy machines
	if p.currentOfferedCapacity() > p.offeredCapacity {
		if len(p.destroyableMachines) > 0 {
			// Need to check if any offered can be destroy
			return r.destroyCapacity(p)
		}
	}
	// Otherwise nonLockedMachines meet Capacity so we do nothing
	return nil
}

func Request(r *RequestMachineArgs) error {
	// First get full info on the pool and the next machine for request
	p, err := getPool(r.PoolName, r.Architecture, r.OSVersion)
	if err != nil {
		return err
	}
	hi, err := p.getNextMachineForRequest()
	if err != nil {
		return err
	}
	mr := macMachine.Request{
		Prefix:               *hi.Prefix,
		Version:              *hi.OSVersion,
		Architecture:         *hi.Arch,
		SetupGHActionsRunner: r.SetupGHActionsRunner,
	}
	// mr := r.fillMacRequest()
	err = mr.ReplaceUserAccess(hi)
	if err != nil {
		return err
	}
	// We update the runID on the dedicated host
	return tag.Update(maptContext.TagKeyRunID,
		maptContext.RunID(),
		*hi.Region,
		*hi.Host.HostId)
}

func Release(r *ReleaseMachineArgs, debug bool, debugLevel uint) error {
	host, err := data.GetDedicatedHost(r.MachineID)
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	// Set context based on info from dedicated host to be released
	maptContext.InitBase(
		*hi.ProjectName,
		*hi.BackedURL,
		debug, debugLevel, false)
	// Set a default request
	mr := &macMachine.Request{
		Prefix:       *hi.Prefix,
		Architecture: *hi.Arch,
		Version:      *hi.OSVersion,
		// We do not want to enable join any ci/cd managed group
		// this should be done on request
		// TODO this should be extended to cirrus
		SetupGHActionsRunner: false,
	}
	return mr.ReplaceMachine(hi)
}

func (r *RequestArgs) addMachinesToPool(n int) error {
	if err := validateBackedURL(); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		hr := r.fillHostRequest()
		dh, err := macHost.CreatePoolDedicatedHost(hr)
		if err != nil {
			return err
		}
		mr := r.fillMacRequest()
		if err = mr.CreateMacMachine(dh); err != nil {
			return err
		}
	}
	return nil
}

// Run serverless operation for house keeping
func (r *RequestArgs) scheduleHouseKeeper() error {
	return serverless.CreateRepeatedlyAsStack(
		getHouseKeepingCommand(
			r.PoolName,
			r.Architecture,
			r.OSVersion,
			r.OfferedCapacity,
			r.MaxSize,
			r.FixedLocation),
		houseKeepingInterval)
}

func getHouseKeepingCommand(poolName, arch, osVersion string,
	offeredCapacity, maxSize int,
	fixedLocation bool) string {
	cmd := fmt.Sprintf(houseKeepingCommand,
		poolName, arch, osVersion,
		offeredCapacity, maxSize)
	if fixedLocation {
		cmd += houseKeepingFixedLocationParam
	}
	return cmd
}

// If we need less or equal than the max allowed on the pool we create all of them
// if need are more than allowed we can create just the allowed
func (r *RequestArgs) addCapacity(p *pool) error {
	allowed := p.maxSize - p.offeredCapacity
	needed := p.offeredCapacity - p.currentOfferedCapacity()
	if needed <= allowed {
		return r.addMachinesToPool(needed)
	}
	return r.addMachinesToPool(allowed)
}

// If we need less or equal than the max allowed on the pool we create all of them
// if need are more than allowed we can create just the allowed
func (r *RequestArgs) destroyCapacity(p *pool) error {
	machinesToDestroy := p.currentOfferedCapacity() - p.offeredCapacity
	for i := 0; i < machinesToDestroy; i++ {
		m := p.destroyableMachines[i]
		if err := aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: mac.StackMacMachine,
			Region:    *m.Region,
			BackedURL: *m.BackedURL,
		}); err != nil {
			return err
		}
		if err := aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: mac.StackDedicatedHost,
			// TODO check if needed to add region for backedURL
			Region:    *m.Region,
			BackedURL: *m.BackedURL,
		}); err != nil {
			return err
		}
	}
	return nil
}

// format for remote backed url when creating the dedicated host
// the backed url from param is used as base and the ID is appended as sub path
func validateBackedURL() error {
	if strings.Contains(maptContext.BackedURL(), "file://") {
		return fmt.Errorf("local backed url is not allowed for mac pool")
	}
	return nil
}

// This function will fill information about machines in the pool
// depending on their state and age full fill the struct to easily
// manage them
func getPool(poolName, arch, osVersion string) (p *pool, err error) {
	// Get machines in the pool
	poolID := &macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	p.machines, err = macHost.GetPoolDedicatedHostsInformation(poolID)
	if err != nil {
		return nil, err
	}
	// non-locked
	p.currentOfferedMachines = util.ArrayFilter(p.machines,
		func(h *mac.HostInformation) bool {
			isLocked, err := mac.IsMachineLocked(h)
			if err != nil {
				logging.Errorf("error checking locking for machine %s", *h.Host.AssetId)
				return false
			}
			return !isLocked
		})
	// non-locked + older than 24 hours
	macAgeDestroyRequeriemnt := time.Now().UTC().
		Add(-24 * time.Hour)
	p.destroyableMachines = util.ArrayFilter(p.currentOfferedMachines,
		func(h *mac.HostInformation) bool {
			return h.Host.AllocationTime.UTC().Before(macAgeDestroyRequeriemnt)
		})
	p.name = poolName
	return
}

// This is a boilerplate function to pick the best machine for
// next request, initially we just pick the newest machine from the
// offered machines, may we can optimize this
func (p *pool) getNextMachineForRequest() (*mac.HostInformation, error) {
	if len(p.currentOfferedMachines) == 0 {
		return nil, fmt.Errorf("no available machines to process the request")
	}
	mp := len(p.currentOfferedMachines) - 1
	return p.currentOfferedMachines[mp], nil
}

// transform pool request to host request
// need if we need to expand the pool
func (r *RequestArgs) fillHostRequest() *macHost.PoolMacDedicatedHostRequestArgs {
	return &macHost.PoolMacDedicatedHostRequestArgs{
		MacDedicatedHost: &macHost.MacDedicatedHostRequestArgs{
			Prefix:        r.Prefix,
			Architecture:  r.Architecture,
			FixedLocation: r.FixedLocation,
		},
		PoolID: &macHost.PoolID{
			PoolName:  r.PoolName,
			Arch:      r.Architecture,
			OSVersion: r.OSVersion,
		},
		BackedURL: fmt.Sprintf("%s/%s",
			maptContext.BackedURL(),
			util.RandomID("mapt")),
	}
}

// transform pool request to machine request
// need if we need to expand the pool
func (r *RequestArgs) fillMacRequest() *macMachine.Request {
	return &macMachine.Request{
		Prefix:       r.Prefix,
		Architecture: r.Architecture,
		Version:      r.OSVersion,
		// SetupGHActionsRunner: r.SetupGHActionsRunner,
		// Airgap:               r.Airgap,
	}
}
