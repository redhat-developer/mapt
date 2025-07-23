package mac

import (
	"github.com/go-playground/validator/v10"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

type MacRequestArgs struct {
	mCtx *mc.Context `validate:"required"`
	// Prefix for the resources related to mac
	// this is relevant in case of an orchestration with multiple
	// macs on the same stack
	Prefix string

	// Machine params
	Architecture string
	Version      string

	// Location params
	FixedLocation    bool
	Region           *string
	AvailabilityZone *string

	// Topology paras
	Airgap bool
}

func (a *MacRequestArgs) validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(a)
}

const (
	DefaultArch      = "m2"
	DefaultOSVersion = "15"
)
