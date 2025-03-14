package macpool

import (
	"fmt"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func releaseRemote(ctx *maptContext.ContextArgs, hostID string) error {
	if err := maptContext.Init(ctx); err != nil {
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
	return fmt.Errorf("not implemented yet")
}

func releaseTaskSpec(poolName, arch, osVersion string) error {
	name := serverlessName(
		poolName,
		arch,
		osVersion, releaseOperation)
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
