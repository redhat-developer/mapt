package windows

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type WindowsRequest struct {
	compute.Request
}

type WindowsCompute struct {
	compute.Compute
}
