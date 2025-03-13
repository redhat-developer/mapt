package macpool

import (
	"fmt"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
)

func request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	// If remote run through serverless
	if r.Remote {
		return requestRemote(ctx, r)
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
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	mr := macMachine.Request{
		Prefix:               *hi.Prefix,
		Version:              *hi.OSVersion,
		Architecture:         *hi.Arch,
		SetupGHActionsRunner: r.SetupGHActionsRunner,
		Timeout:              r.Timeout,
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

func requestRemote(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	return fmt.Errorf("not implemented yet")
}

// Run serverless operation request
// check how we will call it from the request?
// may add tags and find or add arn to stack?
func (r *MacPoolRequestArgs) createRequestTaskSpec() error {
	return serverless.Create(
		&serverless.ServerlessArgs{
			Command: requestCommand(
				r.PoolName,
				r.Architecture,
				r.OSVersion),
			LogGroupName: fmt.Sprintf("%s-%s-%s-request",
				r.PoolName,
				r.Architecture,
				r.OSVersion),
			Tags: map[string]string{
				macConstants.TagKeyArch:      r.Architecture,
				macConstants.TagKeyOSVersion: r.OSVersion,
				macConstants.TagKeyPoolName:  r.PoolName,
			}})
}

func requestCommand(poolName, arch, osVersion string) string {
	cmd := fmt.Sprintf(requestCommandRegex,
		poolName, arch, osVersion)
	return cmd
}
