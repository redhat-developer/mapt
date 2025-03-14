package macpool

import (
	"fmt"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
)

func request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	// If remote run through serverless
	if maptContext.IsRemote() {
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
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}

	mr := macMachine.Request{
		Prefix:       *hi.Prefix,
		Version:      *hi.OSVersion,
		Architecture: *hi.Arch,
		Timeout:      r.Timeout,
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
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	rARNs, err := data.GetResourcesMatchingTags(
		data.ResourceTypeECS,
		requestTags(
			r.PoolName,
			r.Architecture,
			r.OSVersion))
	if err != nil {
		return err
	}
	if len(rARNs) > 1 {
		return fmt.Errorf(
			"should be only one task spec matching tags. Found %s",
			strings.Join(rARNs, ","))
	}
	// We got the arn value for the task
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
			Tags: requestTags(
				r.PoolName,
				r.Architecture,
				r.OSVersion)})
}

func requestCommand(poolName, arch, osVersion string) string {
	cmd := fmt.Sprintf(requestCommandRegex,
		poolName, arch, osVersion)
	return cmd
}

// Return the map of tags wich should identify unique
// resquest operation spec for a pool
func requestTags(poolName, arch, osVersion string) (m map[string]string) {
	poolID := macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	m = poolID.AsTags()
	m[macConstants.TagKeyPoolOperationName] = requestOperation
	return
}
