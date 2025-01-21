package serverless

// Mapts requires the cluster to exist previously wit specific naming
// check hacks/aws/serverless.sh to
var (
	maptServerlessDefaultPrefix = "mapt-serverless-manager"
)

// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#task_size
const (
	LimitCPU    = "2048"
	LimitMemory = "4096"
)

type serverlessRequestArgs struct {
	// need this to find the right ECS cluster to run this serverless
	region string
	// command and scheduling to be used for it
	command, scheduleExpression string
	// optional params in case we create serverless inside a stack
	prefix, componentID string
}
