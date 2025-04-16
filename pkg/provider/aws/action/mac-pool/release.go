package macpool

import (
	"os"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func releaseRemote(ctx *maptContext.ContextArgs, hostID string) error {
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Get host as context will be fullfilled with info coming from the tags on the host
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	tARN, err := serverlessTaskARN(
		*hi.PoolName,
		*hi.Arch,
		*hi.OSVersion,
		releaseOperation)
	if err != nil {
		return err
	}
	logging.Debugf("Got ARN for task spec %s", *tARN)
	// How to handle the region...coming from create operation we are always using "us-east-1"
	defaultRegion := "us-east-1"
	containerName := releaseTaskContainerName(*hi.PoolName, *hi.Arch, *hi.OSVersion)
	// just pass the params
	cmd := strings.Join(os.Args[1:], " ")
	return ecs.RunTaskWithCommand(&defaultRegion, tARN,
		&serverless.MaptServerlessClusterName, &containerName, &cmd)
}

func releaseTaskSpec(poolName, arch, osVersion string) error {
	name := releaseTaskContainerName(poolName, arch, osVersion)
	return serverless.Create(
		&serverless.ServerlessArgs{
			ContainerName: name,
			Command:       releaseCommand,
			LogGroupName:  name,
			Tags: serverlessTags(
				poolName,
				arch,
				osVersion,
				releaseOperation)})
}

func releaseTaskContainerName(poolName, arch, osVersion string) string {
	return serverlessName(poolName, arch, osVersion, releaseOperation)
}
