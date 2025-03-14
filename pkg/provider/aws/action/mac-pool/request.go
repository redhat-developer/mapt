package macpool

import (
	"fmt"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
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
	if err := maptContext.Init(ctx); err != nil {
		return err
	}
	tARN, err := serverlessTaskARN(r.PoolName,
		r.Architecture,
		r.OSVersion,
		requestOperation)
	if err != nil {
		return err
	}
	logging.Debugf("Got ARN for task spec %s", *tARN)
	return fmt.Errorf("not implemented yet")
}

// Run serverless operation request
// check how we will call it from the request?
// may add tags and find or add arn to stack?
func requestTaskSpec(r *MacPoolRequestArgs) error {
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
			Tags: serverlessTags(
				r.PoolName,
				r.Architecture,
				r.OSVersion,
				requestOperation)})
}

func requestCommand(poolName, arch, osVersion string) string {
	cmd := fmt.Sprintf(requestCommandRegex,
		poolName, arch, osVersion)
	return cmd
}
