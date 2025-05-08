package pool

import (
	"fmt"

	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func RequestRemote(ctx *maptContext.ContextArgs, name, arch, osVersion, ticket *string) error {
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	tARN, err := serverlessTaskARN(*name, *arch, *osVersion, operationRequest)
	if err != nil {
		return err
	}
	logging.Debugf("Got ARN for task spec %s", *tARN)
	// How to handle the region...coming from create operation we are always using "us-east-1"
	region := "us-east-1"
	vpcID, azID, subnetID, sshSGID, err := getExecutionDefaultsFromTask(&region, tARN)
	if err != nil {
		return err
	}
	command := requestCommandTotask(vpcID, azID, subnetID, sshSGID, ticket)
	containerName := requestTaskContainerName(*name, *arch, *osVersion)
	// Run task serverless
	return ecs.RunTaskWithCommand(&region, tARN, &serverless.MaptServerlessClusterName,
		&containerName, &command,
		subnetID, sshSGID)
}

func requestCommandTotask(vpcID, azID, subnetID, sshSGID, ticket *string) string {
	command := commandToTask(vpcID, azID, subnetID, sshSGID)
	return fmt.Sprintf("%s %s %s", command, paramTicket, *ticket)

}

// Run serverless operation request
// check how we will call it from the request?
// may add tags and find or add arn to stack?
func requestTaskSpec(ctx *pulumi.Context, p *PoolArgs,
	vpcID, azID, subnetID, sgID *string) (*awsxecs.FargateTaskDefinition, error) {
	name := requestTaskContainerName(
		p.Name,
		p.Arch,
		p.OSVersion)
	return serverless.Deploy(ctx,
		&serverless.ServerlessArgs{
			Prefix:        operationRequest,
			Region:        p.Region,
			ContainerName: name,
			Command: requestCommand(
				p.Name,
				p.Arch,
				p.OSVersion),
			LogGroupName: name,
			ExecutionDefaults: map[string]*string{
				TaskExecDefaultVPCID: vpcID,
				// request do no need AZ cause for mac machines are tied to dedicated hosts
				// which already contains AZ information
				// TaskExecDefaultAZID:                azID,
				serverless.TaskExecDefaultSubnetID: subnetID,
				serverless.TaskExecDefaultSGID:     sgID,
			},
			Tags: serverlessTags(
				p.Name,
				p.Arch,
				p.OSVersion,
				operationRequest)})
}

func requestCommand(poolName, arch, osVersion string) string {
	cmd := fmt.Sprintf(cmdRegexRequest,
		poolName, arch, osVersion)
	return cmd
}

func requestTaskContainerName(poolName, arch, osVersion string) string {
	return serverlessName(poolName, arch, osVersion, operationRequest)
}
