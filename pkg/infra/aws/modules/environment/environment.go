package environment

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
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
	var b *compute.Resources
	if r.bastion != nil {
		logging.Debug("Creating bastion")
		// Compose runtime resources info
		bastionRequest := compute.Request{
			ProjecName: fmt.Sprintf("%s-%s", r.name, "bastion"),
			VPC:        network.VPCResources.VPC,
			Subnets:    []*ec2.Subnet{network.PublicSNResources[0].Subnet},
			Specs:      &supportmatrix.S_BASTION,
			//  need to complete Specs: ,
		}
		b, err = bastionRequest.Create(ctx, &bastionRequest)
		if err != nil {
			return err
		}
	}
	if r.rhel != nil {
		logging.Debug("Creating rhel")
		fillCompute(&r.rhel.Request, network, b)
		_, err = r.rhel.Create(ctx, r.rhel)
		if err != nil {
			return err
		}
	}
	if r.macm1 != nil {
		logging.Debug("Creating macm1")
		fillCompute(&r.macm1.Request, network, b)
		_, err = r.macm1.Create(ctx, r.macm1)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillCompute(request *compute.Request, network *network.NetworkResources,
	bastion *compute.Resources) {
	request.VPC = network.VPCResources.VPC
	if request.Public {
		request.Subnets = []*ec2.Subnet{network.PublicSNResources[0].Subnet}
	} else {
		request.Subnets = []*ec2.Subnet{network.PrivateSNResources[0].Subnet}
	}
	request.AvailabilityZones = []string{network.AvailabilityZones[0]}
	if bastion != nil {
		request.BastionSG = bastion.SG
	}
}
