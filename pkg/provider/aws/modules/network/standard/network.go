package standard

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/subnet"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/vpc"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DefaultNetworkRequest(name, regionName string) NetworkRequest {
	return NetworkRequest{
		Name:                name,
		CIDR:                DefaultCIDRNetwork,
		AvailabilityZones:   data.GetAvailabilityZones()[:3],
		PublicSubnetsCIDRs:  DefaultCIDRPublicSubnets[:],
		PrivateSubnetsCIDRs: DefaultCIDRPrivateSubnets[:],
		IntraSubnetsCIDRs:   DefaultCIDRIntraSubnets[:],
		SingleNatGateway:    false}

}

func (r NetworkRequest) Deployer(ctx *pulumi.Context) (err error) {
	_, err = r.CreateNetwork(ctx)
	return
}

func (r NetworkRequest) CreateNetwork(ctx *pulumi.Context) (*NetworkResources, error) {
	// Data validation
	if err := r.validate(); err != nil {
		return nil, err
	}
	// VPC creation
	vpcRequest := vpc.VPCRequest{CIDR: r.CIDR, Name: r.Name}
	vpcResult, err := vpcRequest.CreateNetwork(ctx)
	if err != nil {
		return nil, err
	}
	ctx.Export(StackCreateNetworkOutputVPCID, vpcResult.VPC.ID())
	// Manage Public Subnets
	publicSNResults, err :=
		r.managePublicSubnets(vpcResult.VPC, vpcResult.InternetGateway, ctx, "public")
	if err != nil {
		return nil, err
	}
	// Manage Private Subnets
	privateSNResults, err :=
		r.managePrivateSubnets(vpcResult.VPC, getNatGateways(publicSNResults), ctx, "private")
	if err != nil {
		return nil, err
	}
	// Manage Intra Subnets
	intraSNResults, err :=
		r.manageIntraSubnets(vpcResult.VPC, ctx, "intra")
	if err != nil {
		return nil, err
	}
	return &NetworkResources{
			VPCResources:       vpcResult,
			AvailabilityZones:  r.AvailabilityZones,
			Region:             r.Region,
			PublicSNResources:  publicSNResults,
			PrivateSNResources: privateSNResults,
			IntraSNResources:   intraSNResults},
		nil
}

func (r NetworkRequest) validate() error {
	if len(r.PublicSubnetsCIDRs) > 0 &&
		len(r.PublicSubnetsCIDRs) > len(r.AvailabilityZones) {
		return fmt.Errorf("availability zones should be minimum same number as public subnets CIDRs blocks")
	}
	if len(r.PrivateSubnetsCIDRs) > 0 &&
		len(r.PrivateSubnetsCIDRs) > len(r.AvailabilityZones) {
		return fmt.Errorf("availability zones should be minimum same number as private subnets CIDRs blocks")
	}
	if len(r.IntraSubnetsCIDRs) > 0 &&
		len(r.IntraSubnetsCIDRs) > len(r.AvailabilityZones) {
		return fmt.Errorf("availability zones should be minimum same number as intra subnets CIDRs blocks")
	}
	return nil
}

func (r NetworkRequest) managePublicSubnets(vpc *ec2.Vpc,
	igw *ec2.InternetGateway, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PublicSubnetResources, err error) {
	if len(r.PublicSubnetsCIDRs) > 0 {
		for i := 0; i < len(r.PublicSubnetsCIDRs); i++ {
			publicSNRequest :=
				subnet.PublicSubnetRequest{
					VPC:              vpc,
					InternetGateway:  igw,
					CIDR:             r.PublicSubnetsCIDRs[i],
					AvailabilityZone: r.AvailabilityZones[i],
					Name:             fmt.Sprintf("%s%s%d", namePrefix, r.Name, i),
					AddNatGateway:    r.checkIfNatGatewayRequired(i)}
			subnet, err := publicSNRequest.Create(ctx)
			if err != nil {
				return nil, err
			}
			subnets = append(subnets, subnet)
		}
	}
	return
}

func (r NetworkRequest) managePrivateSubnets(vpc *ec2.Vpc,
	ngws []*ec2.NatGateway, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PrivateSubnetResources, err error) {
	return managePrivateSubnets(vpc, ngws, ctx, r.PrivateSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.SingleNatGateway)
}

func (r NetworkRequest) manageIntraSubnets(vpc *ec2.Vpc, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PrivateSubnetResources, err error) {
	return managePrivateSubnets(vpc, nil, ctx, r.IntraSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.SingleNatGateway)
}

func managePrivateSubnets(vpc *ec2.Vpc, ngws []*ec2.NatGateway, ctx *pulumi.Context,
	snsCIDRs, azs []string, name, namePrefix string, singleNatGateway bool) (subnets []*subnet.PrivateSubnetResources, err error) {
	if len(snsCIDRs) > 0 {
		for i := 0; i < len(snsCIDRs); i++ {
			privateSNRequest :=
				subnet.PrivateSubnetRequest{
					VPC:              vpc,
					NatGateway:       getNatGateway(singleNatGateway, ngws, i),
					CIDR:             snsCIDRs[i],
					AvailabilityZone: azs[i],
					Name:             fmt.Sprintf("%s%s%d", namePrefix, name, i)}
			subnet, err := privateSNRequest.Create(ctx)
			if err != nil {
				return nil, err
			}
			subnets = append(subnets, subnet)
		}
	}
	return
}

func (r NetworkRequest) checkIfNatGatewayRequired(i int) bool {
	return r.SingleNatGateway && i == 0 || len(r.PrivateSubnetsCIDRs) > 0
}

func getNatGateways(source []*subnet.PublicSubnetResources) (ngws []*ec2.NatGateway) {
	for _, subnet := range source {
		ngws = append(ngws, subnet.NatGateway)
	}
	return
}

func getNatGateway(singleNatGateway bool, ngws []*ec2.NatGateway, i int) *ec2.NatGateway {
	if ngws == nil {
		return nil
	}
	if singleNatGateway {
		return ngws[0]
	}
	return ngws[i]
}
