package environment

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	network "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network/standard"
)

type singleHostRequest struct {
	name          string
	network       *network.NetworkRequest
	bastion       *compute.Request
	hostRequested compute.ComputeRequest
}
