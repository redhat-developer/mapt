package rhel

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type RHELRequest struct {
	VersionMajor string
	compute.Request
}

type RHELResources struct {
	compute.Resources
}
