package macpool

import (
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
)

const (
	houseKeepingCommand            = "aws mac-pool house-keep --name %s --arch %s --version %s --offered-capacity %d --max-size %d --serverless "
	houseKeepingFixedLocationParam = "--fixed-location "
	// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html#eb-rate-expressions
	houseKeepingInterval = "15 minutes"
)

type MacPoolRequestArgs struct {
	// Prefix for the resources related to mac
	// this is relevant in case of an orchestration with multiple
	// macs on the same stack
	Prefix string

	// Pool specs
	PoolName     string
	Architecture string
	OSVersion    string
	// Pool params
	// Capacity is the number of machines in the pool ready to process a workload
	// at any given time
	OfferedCapacity int
	// Max is the max capacity of machines in the pool. Even if capacity is not meet if number of machines
	// are equal to max it will not create more machines
	MaxSize int
	// If fixed location is set machines only created on current region, in case no capacity it will not
	// create on different regions
	FixedLocation bool
}

type RequestMachineArgs struct {
	PoolName     string
	Architecture string
	OSVersion    string
	// This is sampler for type of integration / usage
	// for the request machine
	SetupGHActionsRunner bool
}

type ReleaseMachineArgs struct {
}

type pool struct {
	// pool params
	name            string
	offeredCapacity int
	maxSize         int
	// all machines in the pool
	machines []*mac.HostInformation
	// non locked machines (not running a workload)
	currentOfferedMachines []*mac.HostInformation
	// machines which can be destroyed:
	// non locked + older than 24 hours
	// also this slice is order by age, so first is the oldest
	destroyableMachines []*mac.HostInformation
}

func (p *pool) currentOfferedCapacity() int { return len(p.currentOfferedMachines) }
func (p *pool) currentPoolSize() int        { return len(p.machines) }
