package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/macm1"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/rhel"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
)

type corporateEnvironmentRequest struct {
	name    string
	network *network.NetworkRequest
	bastion *compute.Request
	rhel    *rhel.RHELRequest
	macm1   *macm1.MacM1Request
}
