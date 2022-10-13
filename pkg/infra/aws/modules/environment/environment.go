package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/bastion"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r corporateEnvironmentRequest) deployer(ctx *pulumi.Context) error {
	network, err := network.DefaultNetworkRequest(ctx, r.name).CreateNetwork(ctx)
	if err != nil {
		return err
	}
	_, err = bastion.BastionRequest{
		Name:          r.name,
		HA:            false,
		VPC:           network.VPCResources.VPC,
		PublicSubnets: []*ec2.Subnet{network.PublicSNResources[0].Subnet},
	}.Create(ctx)
	if err != nil {
		return err
	}
	return nil
}
