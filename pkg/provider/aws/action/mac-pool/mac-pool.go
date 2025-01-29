package macpool

import (
	"fmt"
	"strings"
	"time"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	macUtil "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/util"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// Create works as an orchestrator for create n machines based on offered capacity
// if pool already exists just change the params for the HouseKeeper
// also the HouseKeep will take care of regulate the capacity

// Even if we want to destroy the pool we will set params to max size 0
func Create(ctx *maptContext.ContextArgs, r *MacPoolRequestArgs) error {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}
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
func HouseKeeper(ctx *maptContext.ContextArgs, r *MacPoolRequestArgs) error {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	// Get full info on the pool
	p, err := getPool(r.PoolName, r.Architecture, r.OSVersion)
	if err != nil {
		return err
	}
	// Pool under expected offered capacity
	if p.currentOfferedCapacity() < r.OfferedCapacity {
		if p.currentPoolSize() < r.MaxSize {
			logging.Debug("house keeper will try to add machines as offered capacity is lower than expected")
			return r.addCapacity(p)
		}
		// if number of machines in the pool + to max machines
		// we do nothing
		logging.Debug("house keeper will not do any action as pool size is currently at max size")
		return nil
	}
	// Pool over expected offered capacity need to destroy machines
	if p.currentOfferedCapacity() > r.OfferedCapacity {
		if len(p.destroyableMachines) > 0 {
			logging.Debug("house keeper will try to destroy machines as offered capacity is higher than expected")
			// Need to check if any offered can be destroy
			return r.destroyCapacity(p)
		}
	}
	logging.Debug("house keeper will not do any action as offered capacity is met by the pool")
	// Otherwise nonLockedMachines meet Capacity so we do nothing
	return nil
}

func Request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	// First get full info on the pool and the next machine for request
	p, err := getPool(r.PoolName, r.Architecture, r.OSVersion)
	if err != nil {
		return err
	}
	hi, err := p.getNextMachineForRequest()
	if err != nil {
		return err
	}

	// Create mapt Context
	ctx.ProjectName = *hi.ProjectName
	ctx.BackedURL = *hi.BackedURL
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	mr := macMachine.Request{
		Prefix:               *hi.Prefix,
		Version:              *hi.OSVersion,
		Architecture:         *hi.Arch,
		SetupGHActionsRunner: r.SetupGHActionsRunner,
	}

	// TODO here we would change based on the integration-mode requested
	// possible values remote-shh, gh-selfhosted-runner, cirrus-persistent-worker
	err = mr.ManageRequest(hi)
	if err != nil {
		return err
	}
	// We update the runID on the dedicated host
	return tag.Update(maptContext.TagKeyRunID,
		maptContext.RunID(),
		*hi.Region,
		*hi.Host.HostId)
}

func Release(ctx *maptContext.ContextArgs, hostID string) error {
	return macUtil.Release(ctx, hostID)
}

func (r *MacPoolRequestArgs) addMachinesToPool(n int) error {
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
		if err = mr.CreateAvailableMacMachine(dh); err != nil {
			return err
		}
	}
	return nil
}

// Run serverless operation for house keeping
func (r *MacPoolRequestArgs) scheduleHouseKeeper() error {
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
func (r *MacPoolRequestArgs) addCapacity(p *pool) error {
	allowed := p.maxSize - p.offeredCapacity
	needed := p.offeredCapacity - p.currentOfferedCapacity()
	if needed <= allowed {
		return r.addMachinesToPool(needed)
	}
	return r.addMachinesToPool(allowed)
}

// If we need less or equal than the max allowed on the pool we create all of them
// if need are more than allowed we can create just the allowed
func (r *MacPoolRequestArgs) destroyCapacity(p *pool) error {
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
func getPool(poolName, arch, osVersion string) (*pool, error) {
	// Get machines in the pool
	poolID := &macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	var p pool
	var err error
	p.machines, err = macHost.GetPoolDedicatedHostsInformation(poolID)
	if err != nil {
		return nil, err
	}
	// non-locked
	p.currentOfferedMachines = util.ArrayFilter(p.machines,
		func(h *mac.HostInformation) bool {
			isLocked, err := macUtil.IsMachineLocked(h)
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
	return &p, nil
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
func (r *MacPoolRequestArgs) fillHostRequest() *macHost.PoolMacDedicatedHostRequestArgs {
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
func (r *MacPoolRequestArgs) fillMacRequest() *macMachine.Request {
	return &macMachine.Request{
		Prefix:       r.Prefix,
		Architecture: r.Architecture,
		Version:      r.OSVersion,
		// SetupGHActionsRunner: r.SetupGHActionsRunner,
		// Airgap:               r.Airgap,
	}
}
