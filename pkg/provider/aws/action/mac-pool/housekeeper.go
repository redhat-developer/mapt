package macpool

import (
	"fmt"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// House keeper is the function executed serverless to check if is there any
// machine non locked which had been running more than 24h.
// It should check if capacity allows to remove the machine
func houseKeeper(ctx *maptContext.ContextArgs, r *MacPoolRequestArgs) error {
	// Create mapt Context, this is a special case where we need change the context
	// based on the operation
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
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
			maptContext.SetProjectName(r.PoolName)
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

// Run serverless operation for house keeping
func (r *MacPoolRequestArgs) scheduleHouseKeeper() error {
	return serverless.Create(
		&serverless.ServerlessArgs{
			Command: houseKeepingCommand(
				r.PoolName,
				r.Architecture,
				r.OSVersion,
				r.OfferedCapacity,
				r.MaxSize,
				r.FixedLocation),
			ScheduleType:      &serverless.Repeat,
			Schedulexpression: houseKeepingInterval,
			LogGroupName: fmt.Sprintf("%s-%s-%s",
				r.PoolName,
				r.Architecture,
				r.OSVersion)})
}

func houseKeepingCommand(poolName, arch, osVersion string,
	offeredCapacity, maxSize int,
	fixedLocation bool) string {
	cmd := fmt.Sprintf(houseKeepingCommandRegex,
		poolName, arch, osVersion,
		offeredCapacity, maxSize)
	if fixedLocation {
		cmd += houseKeepingFixedLocationParam
	}
	return cmd
}
