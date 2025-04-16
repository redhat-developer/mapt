package macpool

import (
	"fmt"
	"os"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
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
	tARN, err := serverlessTaskARN(r.PoolName,
		r.Architecture,
		r.OSVersion,
		requestOperation)
	if err != nil {
		return err
	}
	logging.Debugf("Got ARN for task spec %s", *tARN)
	// How to handle the region...coming from create operation we are always using "us-east-1"
	defaultRegion := "us-east-1"
	containerName := requestTaskContainerName(r.PoolName, r.Architecture, r.OSVersion)
	// just pass the params
	cmd := strings.Join(os.Args[1:], " ")
	return ecs.RunTaskWithCommand(&defaultRegion, tARN,
		&serverless.MaptServerlessClusterName, &containerName, &cmd)
}

// Run serverless operation request
// check how we will call it from the request?
// may add tags and find or add arn to stack?
func requestTaskSpec(r *MacPoolRequestArgs) error {
	name := requestTaskContainerName(
		r.PoolName,
		r.Architecture,
		r.OSVersion)
	return serverless.Create(
		&serverless.ServerlessArgs{
			ContainerName: name,
			Command: requestCommand(
				r.PoolName,
				r.Architecture,
				r.OSVersion),
			LogGroupName: name,
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

func requestTaskContainerName(poolName, arch, osVersion string) string {
	return serverlessName(poolName, arch, osVersion, requestOperation)
}
