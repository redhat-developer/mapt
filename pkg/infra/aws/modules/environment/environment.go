package environment

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/bastion"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r corporateEnvironmentRequest) deployer(ctx *pulumi.Context) error {
	logging.Debug("Creating network")
	network, err := r.network.CreateNetwork(ctx)
	if err != nil {
		return err
	}
	var b *bastion.BastionResources
	if r.bastion != nil {
		logging.Debug("Creating bastion")
		// Compose runtime resources info
		_, err = bastion.BastionRequest{
			Name:          fmt.Sprintf("%s-%s", r.name, "bastion"),
			HA:            false,
			VPC:           network.VPCResources.VPC,
			PublicSubnets: []*ec2.Subnet{network.PublicSNResources[0].Subnet},
		}.Create(ctx)
		if err != nil {
			return err
		}
	}
	if r.rhel != nil {
		logging.Debug("Creating rhel")
		// Compose runtime resources info
		r.rhel.VPC = network.VPCResources.VPC
		if r.rhel.Public {
			r.rhel.Subnets = []*ec2.Subnet{network.PublicSNResources[0].Subnet}
		} else {
			r.rhel.Subnets = []*ec2.Subnet{network.PrivateSNResources[0].Subnet}
		}
		if b != nil {
			r.rhel.BastionSG = b.SG
		}
		_, err = r.rhel.Create(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
