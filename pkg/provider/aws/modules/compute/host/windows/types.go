package windows

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
)

type WindowsRequest struct {
	compute.Request
}

type WindowsCompute struct {
	compute.Compute
}
