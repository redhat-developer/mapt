package pool

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ecs"
)

// Operations are based on mapt params around mac-pool
var (
	operationHouseKeep = "mac-pool-housekeep"
	cmdRegexHouseKeep  = "aws mac-pool house-keep --name %s --arch %s --version %s --offered-capacity %d --max-size %d --vpcid %s --ssh-sgid %s --project-name %s --backed-url %s --serverless"
	// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html#eb-rate-expressions
	scheduleIntervalHouseKeep = "27 minutes"

	operationRequest = "mac-pool-request"
	cmdRegexRequest  = "aws mac-pool request --name %s --arch %s --version %s --serverless "

	operationRelease = "mac-pool-release"
	cmdRelease       = "aws mac-pool release --serverless "

	remoteCommandParamsRegex = "--vpcid %s --ssh-sgid %s --serverless "
	paramTicket              = "--ticket"
)

// this is a business identificator to assing to resources related to the serverless management
func serverlessName(poolName, arch, osVersion, operation string) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		operation,
		poolName,
		arch,
		osVersion)
}

// Return the map of tags wich should identify unique
// resquest operation spec for a pool
func serverlessTags(poolName, arch, osVersion, operation string) (m map[string]string) {
	poolID := macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	m = poolID.AsTags()
	m[macConstants.TagKeyPoolOperationName] = operation
	return
}

func serverlessTaskARN(poolName, arch, osVersion, operation string) (*string, error) {
	rARNs, err := data.GetResourcesMatchingTags(
		data.ResourceTypeECS,
		serverlessTags(
			poolName,
			arch,
			osVersion,
			operation))
	if err != nil {
		return nil, err
	}
	arn, err := data.ActiveTasks(rARNs)
	if err != nil {
		return nil, err
	}
	return arn, nil

}

// Command to be executed within the task is same as remote command without --remote flag
// and with specific network and security data (coming from task def as tags)
func commandToTask(vpcID, sshSGID *string) string {
	rawCmd := strings.Join(os.Args[1:], " ")
	cmd := strings.Replace(rawCmd, "--remote ", "", 1)
	remoteParams := fmt.Sprintf(remoteCommandParamsRegex,
		vpcID, sshSGID)
	return fmt.Sprintf("%s %s", cmd, remoteParams)
}

func getExecutionDefaultsFromTask(region *string, taskDefArn *string) (vpcID, sshSGID *string, err error) {
	var tags map[string]*string
	tags, err = ecs.GetTags(region, taskDefArn)
	if err != nil {
		return
	}
	vpcID = tags[serverless.TaskExecDefaultVPCID]
	sshSGID = tags[serverless.TaskExecDefaultSGID]
	return
}
