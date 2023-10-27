package environment

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	network "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network/standard"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r singleHostRequest) deployer(ctx *pulumi.Context) error {
	logging.Debug("Creating network")
	network, err := r.network.CreateNetwork(ctx)
	if err != nil {
		return err
	}

	if r.hostRequested != nil {
		logging.Debug("Creating requested host %")
		fillCompute(r.hostRequested.GetRequest(), network, nil)
		_, err = r.hostRequested.Create(ctx, r.hostRequested)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillCompute(request *compute.Request, network *network.NetworkResources,
	bastion *compute.Compute) {
	request.VPC = network.VPCResources.VPC
	if request.Public {
		request.Subnets = []*ec2.Subnet{network.PublicSNResources[0].Subnet}
	} else {
		request.Subnets = []*ec2.Subnet{network.PrivateSNResources[0].Subnet}
	}
	request.AvailabilityZones = []string{network.AvailabilityZones[0]}
	request.Region = network.Region
	if bastion != nil {
		request.BastionSG = bastion.SG[0]
	}
}
