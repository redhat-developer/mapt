package macpool

import (
	"fmt"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// House keeper is the function executed serverless to check if is there any
// machine non locked which had been running more than 24h.
// It should check if capacity allows to remove the machine
func houseKeeper(ctx *maptContext.ContextArgs, r *HouseKeepRequestArgs) error {
	// Create mapt Context, this is a special case where we need change the context
	// based on the operation
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Get full info on the pool
	p, err := getPool(r.Pool.Name, r.Pool.Architecture, r.Pool.OSVersion)
	if err != nil {
		return err
	}
	p.maxSize = r.Pool.MaxSize
	p.offeredCapacity = r.Pool.OfferedCapacity
	logging.Debugf("Current pool: %s", p.print())
	// Pool under expected offered capacity
	if p.currentOfferedCapacity() < r.Pool.OfferedCapacity {
		if p.currentPoolSize() < r.Pool.MaxSize {
			logging.Debug("house keeper will try to add machines as offered capacity is lower than expected")
			maptContext.SetProjectName(r.Pool.Name)
			return r.addCapacity(p)
		}
		// if number of machines in the pool + to max machines
		// we do nothing
		logging.Debug("house keeper will not do any action as pool size is currently at max size")
		return nil
	}
	// Pool over expected offered capacity need to destroy machines
	if p.currentOfferedCapacity() > r.Pool.OfferedCapacity {
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

func (r *HouseKeepRequestArgs) addMachinesToPool(n int) error {
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

// format for remote backed url when creating the dedicated host
// the backed url from param is used as base and the ID is appended as sub path
func validateBackedURL() error {
	if strings.Contains(maptContext.BackedURL(), "file://") {
		return fmt.Errorf("local backed url is not allowed for mac pool")
	}
	return nil
}

// transform pool request to host request
// need if we need to expand the pool
func (r *HouseKeepRequestArgs) fillHostRequest() *macHost.PoolMacDedicatedHostRequestArgs {
	return &macHost.PoolMacDedicatedHostRequestArgs{
		MacDedicatedHost: &macHost.MacDedicatedHostRequestArgs{
			Prefix:       r.Pool.Prefix,
			Architecture: r.Pool.Architecture,
			// FixedLocation: r.FixedLocation,
			VPCID:   &r.Machine.VPCID,
			SSHSGID: &r.Machine.SSHSGID,
		},
		PoolID: &macHost.PoolID{
			PoolName:  r.Pool.Name,
			Arch:      r.Pool.Architecture,
			OSVersion: r.Pool.OSVersion,
		},
		BackedURL: fmt.Sprintf("%s/%s",
			maptContext.BackedURL(),
			util.RandomID("mapt")),
	}
}

// If we need less or equal than the max allowed on the pool we create all of them
// if need are more than allowed we can create just the allowed
func (r *HouseKeepRequestArgs) addCapacity(p *pool) error {
	machinesToAdd := p.offeredCapacity - p.currentOfferedCapacity()
	if machinesToAdd+p.currentPoolSize() > p.maxSize {
		machinesToAdd = p.maxSize - p.currentPoolSize()
	}
	logging.Debugf("Adding %d machines", machinesToAdd)
	return r.addMachinesToPool(machinesToAdd)
}

// If we need less or equal than the max allowed on the pool we create all of them
// if need are more than allowed we can create just the allowed
// TODO review allocation time is on the wrong order
func (r *HouseKeepRequestArgs) destroyCapacity(p *pool) error {
	machinesToDestroy := p.currentOfferedCapacity() - r.Pool.OfferedCapacity
	for i := 0; i < machinesToDestroy; i++ {
		m := p.destroyableMachines[i]
		// TODO change this
		maptContext.SetProjectName(*m.ProjectName)
		if err := aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: mac.StackMacMachine,
			Region:    *m.Region,
			BackedURL: *m.BackedURL,
		}); err != nil {
			return err
		}
		if err := macHost.DestroyPoolDedicatedHost(&r.Pool.Prefix); err != nil {
			return err
		}
	}
	return nil
}
