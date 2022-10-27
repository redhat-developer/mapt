package rhel

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type RHELRequest struct {
	VersionMajor string
	// username and password to handle rh subscription on userdata
	SubscriptionUsername string
	SubscriptionPassword string
	compute.Request
}

type RHELCompute struct {
	compute.Compute
}
