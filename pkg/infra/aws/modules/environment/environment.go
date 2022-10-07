package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r corporateEnvironmentRequest) deployer(ctx *pulumi.Context) error {
	network.DefaultNetworkRequest(ctx, r.name)
	return nil
}
