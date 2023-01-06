package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
)

type singleHostRequest struct {
	name          string
	network       *network.NetworkRequest
	bastion       *compute.Request
	hostRequested compute.ComputeRequest
}
