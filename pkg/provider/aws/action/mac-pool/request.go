package macpool

import (
	"fmt"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
)

func request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	if len(r.Ticket) == 0 {
		// Generate ticket
		ticket, err := ticket()
		if err != nil {
			return err
		}
		r.Ticket = *ticket
	}
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
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}

	mr := macMachine.Request{
		Prefix:       *hi.Prefix,
		Version:      *hi.OSVersion,
		Architecture: *hi.Arch,
		VPCID:        &r.Machine.VPCID,
		SSHSGID:      &r.Machine.SSHSGID,
		Timeout:      r.Timeout,
	}

	// TODO here we would change based on the integration-mode requested
	// possible values remote-shh, gh-selfhosted-runner, cirrus-persistent-worker
	err = mr.ManageRequest(hi)
	if err != nil {
		return err
	}

	// We update the runID on the dedicated host
	err = tag.Update(maptContext.TagKeyRunID,
		maptContext.RunID(),
		*hi.Region,
		*hi.Host.HostId)
	if err != nil {
		return err
	}
	return writeTicket(&r.Ticket)
}

// func requestRemote(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
// 	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
// 		return err
// 	}
// 	rARNs, err := data.GetResourcesMatchingTags(
// 		data.ResourceTypeECS,
// 		requestTags(
// 			r.PoolName,
// 			r.Architecture,
// 			r.OSVersion))
// 	if err != nil {
// 		return err
// 	}
// 	err = tag.Update(macConstants.TagKeyTicket,
// 		r.Ticket,
// 		*hi.Region,
// 		*hi.Host.HostId)
// 	if err != nil {
// 		return err
// 	}
// 	return writeTicket(&r.Ticket)
// }

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

// transform pool request to machine request
// need if we need to expand the pool
func (r *HouseKeepRequestArgs) fillMacRequest() *macMachine.Request {
	return &macMachine.Request{
		Prefix:       r.Pool.Prefix,
		Architecture: r.Pool.Architecture,
		Version:      r.Pool.OSVersion,
		// Network and Security
		VPCID: &r.Machine.VPCID,
		// SubnetID: &r.Machine.SubnetID,
		SSHSGID: &r.Machine.SSHSGID,
	}
}
