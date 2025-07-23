package serverless

import (
	"github.com/go-playground/validator/v10"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

var (
	// stackName = "mapt-serverless"

	// This is mostly used to prefix resources used by mapt on serverless mode
	// i.e. ECS clusters which are created and kept after destroy as they always
	// will be used by mapt serverless features
	maptServerlessDefaultPrefix = "mapt-serverless-manager"
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

type serverlessRequest struct {
	mCtx *mc.Context `validate:"required"`
	// need this to find the right ECS cluster to run this serverless
	region string
	// command and scheduling to be used for it
	command            string
	scheduleType       scheduleType
	scheduleExpression string

	// optional if we want to set the name for the log group were logs are sent
	// to facilitate find it out
	logGroupName string
	// optional params in case we create serverless inside a stack
	prefix, componentID string
}

func (r *serverlessRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}
