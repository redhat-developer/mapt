package iam

import (
	"github.com/go-playground/validator/v10"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

const (
	stackName = "iam-manager"

	outputAccessKey = "accessKey"
	outputSecretKey = "secretKey"
)

type iamRequestArgs struct {
	mCtx *mc.Context `validate:"required"`
	// need this to find the right ECS cluster to run this serverless
	name string
	// command and scheduling to be used for it
	policyContent *string
	// optional params in case we create serverless inside a stack
	prefix, componentID string
}

func (r *iamRequestArgs) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}
