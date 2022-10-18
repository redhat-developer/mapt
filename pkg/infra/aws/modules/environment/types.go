package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/bastion"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/rhel"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
)

type corporateEnvironmentRequest struct {
	name    string
	network *network.NetworkRequest
	bastion *bastion.BastionRequest
	rhel    *rhel.RHELRequest
}
