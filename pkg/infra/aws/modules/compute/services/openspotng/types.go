package openspotng

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type OpenspotNGRequest struct {
	compute.Request
}

type OpenspotNGCompute struct {
	compute.Compute
}
