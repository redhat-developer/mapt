package macpool

import (
	"fmt"

	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	"github.com/redhat-developer/mapt/pkg/util"
)

type PoolRequestArgs struct {
	// Prefix for the resources related to mac
	// this is relevant in case of an orchestration with multiple
	// macs on the same stack
	Prefix string

	// Pool specs
	Name         string
	Architecture string
	OSVersion    string
	// Pool params
	// Capacity is the number of machines in the pool ready to process a workload
	// at any given time
	OfferedCapacity int
	// Max is the max capacity of machines in the pool. Even if capacity is not meet if number of machines
	// are equal to max it will not create more machines
	MaxSize int
}

// Custom values to setup within machines in the cluster
type MachineRequestArgs struct {
	VPCID string
	// This values now will be calculated
	// AZID     string
	// SubnetID string
	SSHSGID string
}

type HouseKeepRequestArgs struct {
	// Pool identification and capacity
	Pool *PoolRequestArgs
	// House keeper requires extra info for setup network and security for managed machines
	Machine *MachineRequestArgs
}

type RequestMachineArgs struct {
	// Pool identification
	PoolName     string
	Architecture string
	OSVersion    string
	// Ticket to assing the dedicated host
	Ticket string
	// Request needs this information to manage the machine
	Machine *MachineRequestArgs
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
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

func (p *pool) currentOfferedCapacity() int {
	return util.If(p.currentOfferedMachines == nil,
		0,
		len(p.currentOfferedMachines))
}

func (p *pool) currentPoolSize() int {
	return util.If(
		p.machines == nil,
		0,
		len(p.machines))
}

func (p *pool) print() string {
	return fmt.Sprintf("name: %s, offeredCapacity: %d, maxSize: %d, currentOfferedCapacity: %d, currentSize: %d",
		p.name, p.offeredCapacity, p.maxSize, p.currentOfferedCapacity(), p.currentPoolSize())
}
