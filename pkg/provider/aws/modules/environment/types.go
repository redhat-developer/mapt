package environment

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network"
)

type singleHostRequest struct {
	name          string
	network       *network.NetworkRequest
	bastion       *compute.Request
	hostRequested compute.ComputeRequest
}
