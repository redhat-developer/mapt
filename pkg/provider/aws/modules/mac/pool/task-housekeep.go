package pool

import (
	"fmt"

	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
)

func houseKeeperRemote(tARN *string,
	name, arch, osVersion *string,
	offeredCapacity, maxSize *int,
	vpcID, azID, subnetID, sshSGID *string) error {
	// How to handle the region...coming from create operation we are always using "us-east-1"
	region := "us-east-1"
	command := houseKeepCommand(*name, *arch, *osVersion,
		*offeredCapacity, *maxSize,
		vpcID, azID, subnetID, sshSGID)
	containerName := houseKeepContainerName(*name, *arch, *osVersion)
	return ecs.RunTaskWithCommand(&region, tARN, &serverless.MaptServerlessClusterName,
		&containerName, &command,
		subnetID, sshSGID)
}

func houseKeeperTaskSpecScheduler(ctx *pulumi.Context, p *PoolArgs,
	vpcID, azID, subnetID, sgID *string) (*awsxecs.FargateTaskDefinition, error) {
	cn := houseKeepContainerName(
		p.Name,
		p.Arch,
		p.OSVersion)
	return serverless.Deploy(ctx,
		&serverless.ServerlessArgs{
			Prefix:        operationHouseKeep,
			Region:        p.Region,
			ContainerName: cn,
			Command: houseKeepCommand(
				p.Name, p.Arch, p.OSVersion,
				p.OfferedCapacity, p.MaxSize,
				vpcID, azID, subnetID, sgID),
			ScheduleType:      &serverless.Repeat,
			Schedulexpression: scheduleIntervalHouseKeep,
			LogGroupName:      cn,
			// These values are required to setup the scheduler as the container
			// running the task should be executed within same subnet and with sshsgid
			// in order to ssh into mac machine
			ExecutionDefaults: map[string]*string{
				TaskExecDefaultVPCID:               vpcID,
				TaskExecDefaultAZID:                azID,
				serverless.TaskExecDefaultSubnetID: subnetID,
				serverless.TaskExecDefaultSGID:     sgID,
			}})
}

func houseKeepContainerName(name, arch, osVersion string) string {
	return fmt.Sprintf("housekeeper-%s-%s-%s",
		name,
		arch,
		osVersion)
}

func houseKeepCommand(poolName, arch, osVersion string,
	offeredCapacity, maxSize int, vpcID, azID, subnetID, sshSGID *string) string {
	cmd := fmt.Sprintf(cmdRegexHouseKeep,
		poolName, arch, osVersion,
		offeredCapacity, maxSize,
		*vpcID, *azID, *subnetID, *sshSGID,
		maptContext.ProjectName(), maptContext.BackedURL())
	return cmd
}
