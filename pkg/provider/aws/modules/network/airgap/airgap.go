package airgap

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/subnet"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/vpc"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AirgapNetworkRequest struct {
	CIDR   string
	Name   string
	Region string

	AvailabilityZone string
	// This subnet is will be created first as private
	// on the orchestrate on a 2nd phase a param will remove the
	// nat gateway
	TargetSubnetCIDR string
	PublicSubnetCIDR string
	SetAsAirgap      bool
}

type AirgapNetworkResources struct {
	VPCResources     *vpc.VPCResources
	AvailabilityZone string
	Region           string
	TargetSubnet     *subnet.PrivateSubnetResources
	PublicSubnet     *subnet.PublicSubnetResources
}

func (r AirgapNetworkRequest) CreateNetwork(ctx *pulumi.Context) (*AirgapNetworkResources, error) {
	var result = AirgapNetworkResources{}
	var err error
	// VPC creation
	vpcRequest := vpc.VPCRequest{CIDR: r.CIDR, Name: r.Name}
	result.VPCResources, err = vpcRequest.CreateNetwork(ctx)
	if err != nil {
		return nil, err
	}
	// Manage Public Subnet
	result.PublicSubnet, err =
		r.managePublicSubnet(ctx,
			result.VPCResources.VPC,
			result.VPCResources.InternetGateway,
			"public")
	if err != nil {
		return nil, err
	}
	var natGateway *ec2.NatGateway = nil
	if !r.SetAsAirgap {
		natGateway = result.PublicSubnet.NatGateway
	}
	// Manage Private Subnets
	result.TargetSubnet, err =
		r.manageTargetSubnet(ctx,
			result.VPCResources.VPC,
			natGateway, "airgap")
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r AirgapNetworkRequest) managePublicSubnet(ctx *pulumi.Context,
	vpc *ec2.Vpc, igw *ec2.InternetGateway, namePrefix string) (
	*subnet.PublicSubnetResources, error) {
	publicSNRequest :=
		subnet.PublicSubnetRequest{
			VPC:              vpc,
			InternetGateway:  igw,
			CIDR:             r.PublicSubnetCIDR,
			AvailabilityZone: r.AvailabilityZone,
			Name:             fmt.Sprintf("%s%s", namePrefix, r.Name),
			// Depending on the phase we create or not a NatGateway
			AddNatGateway: !r.SetAsAirgap}
	return publicSNRequest.Create(ctx)
}

func (r AirgapNetworkRequest) manageTargetSubnet(ctx *pulumi.Context,
	vpc *ec2.Vpc, ngw *ec2.NatGateway,
	namePrefix string) (
	*subnet.PrivateSubnetResources, error) {
	privateSNRequest :=
		subnet.PrivateSubnetRequest{
			VPC:              vpc,
			NatGateway:       ngw,
			CIDR:             r.TargetSubnetCIDR,
			AvailabilityZone: r.AvailabilityZone,
			Name:             fmt.Sprintf("%s%s", namePrefix, r.Name)}
	return privateSNRequest.Create(ctx)
}
