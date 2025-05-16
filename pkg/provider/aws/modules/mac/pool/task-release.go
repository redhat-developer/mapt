package pool

import (
	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func ReleaseRemote(ctx *maptContext.ContextArgs, ticket string) error {
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Get host as context will be fullfilled with info coming from the tags on the host
	host, err := data.GetDedicatedHostByTag(map[string]string{macConstants.TagKeyTicket: ticket})
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	tARN, err := serverlessTaskARN(
		*hi.PoolName,
		*hi.Arch,
		*hi.OSVersion,
		operationRelease)
	if err != nil {
		return err
	}
	logging.Debugf("Got ARN for task spec %s", *tARN)
	// How to handle the region...coming from create operation we are always using "us-east-1"
	region := "us-east-1"
	vpcID, sshSGID, err := getExecutionDefaultsFromTask(&region, tARN)
	if err != nil {
		return err
	}
	command := commandToTask(vpcID, sshSGID)
	containerName := releaseTaskContainerName(*hi.PoolName, *hi.Arch, *hi.OSVersion)
	subnetID, err := data.GetSubnetID(&data.SubnetRequestArgs{
		Region: &region,
		VpcId:  vpcID})
	if err != nil {
		return err
	}
	return ecs.RunTaskWithCommand(&region, tARN, &serverless.MaptServerlessClusterName,
		&containerName, &command,
		subnetID, sshSGID)
}

func releaseTaskSpec(ctx *pulumi.Context, p *PoolArgs,
	vpcID, subnetID, sgID *string) (*awsxecs.FargateTaskDefinition, error) {
	cn := releaseTaskContainerName(p.Name, p.Arch, p.OSVersion)
	return serverless.Deploy(
		ctx,
		&serverless.ServerlessArgs{
			Prefix:        operationRelease,
			Region:        p.Region,
			ContainerName: cn,
			Command:       cmdRelease,
			LogGroupName:  cn,
			ExecutionDefaults: map[string]*string{
				serverless.TaskExecDefaultVPCID:    vpcID,
				serverless.TaskExecDefaultSubnetID: subnetID,
				serverless.TaskExecDefaultSGID:     sgID,
			},
			Tags: serverlessTags(
				p.Name,
				p.Arch,
				p.OSVersion,
				operationRelease)})
}

func releaseTaskContainerName(poolName, arch, osVersion string) string {
	return serverlessName(poolName, arch, osVersion, operationRelease)
}
