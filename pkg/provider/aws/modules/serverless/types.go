package serverless

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	defaultPrefix      string = "mapt"
	defaultComponentID string = "sf"
)

var (
	// This is mostly used to prefix resources used by mapt on serverless mode
	// i.e. ECS clusters which are created and kept after destroy as they always
	// will be used by mapt serverless features
	maptServerlessDefaultPrefix = "mapt-serverless-manager"
	MaptServerlessClusterName   = fmt.Sprintf("%s-%s", maptServerlessDefaultPrefix, "cluster")
	maptServerlessExecRoleName  = fmt.Sprintf("%s-%s", maptServerlessDefaultPrefix, "sch-role")
)

const (
	TaskExecDefaultSubnetID = "default_subnetid"
	TaskExecDefaultVPCID    = "default_vpcid"
	TaskExecDefaultSGID     = "default_sgid"
)

type scheduleType string

var (
	Repeat  scheduleType = "rate"
	OneTime scheduleType = "at"
)

// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#task_size
const (
	LimitCPU    = "1024"
	LimitMemory = "2048"
)

type ServerlessArgs struct {
	Prefix        string
	Region        string
	ContainerName string
	Command       string
	// From here params are optional
	LogGroupName string
	Tags         map[string]string
	// If no schedule info is added just create the task spec
	ScheduleType      *scheduleType
	Schedulexpression string
	// Optional information to use for execute the task
	ExecutionDefaults map[string]*string
}

type serverlessRequestArgs struct {
	containerName string
	// need this to find the right ECS cluster to run this serverless
	region string
	// command and scheduling to be used for it
	command            string
	scheduleType       *scheduleType
	scheduleExpression string

	// optional if we want to set the name for the log group were logs are sent
	// to facilitate find it out
	logGroupName string
	// optional params in case we create serverless inside a stack
	prefix, componentID string
	tags                pulumi.StringMap
	executionDefaults   map[string]*string
}
