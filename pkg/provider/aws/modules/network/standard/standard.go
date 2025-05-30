package standard

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/subnet"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/vpc"
)

const (
	StackCreateNetworkName        string = "Manage-Network"
	StackCreateNetworkOutputVPCID string = "VPCID"
)

var (
	DefaultCIDRNetwork string = "10.0.0.0/16"

	DefaultCIDRPublicSubnets [6]string = [6]string{
		"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24",
		"10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"}
	DefaultLBIPs [6]string = [6]string{
		"10.0.1.15", "10.0.2.15", "10.0.3.15",
		"10.0.4.15", "10.0.5.15", "10.0.6.15"}
	DefaultCIDRPrivateSubnets [6]string = [6]string{
		"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24",
		"10.0.104.0/24", "10.0.105.0/24", "10.0.106.0/24"}
	DefaultCIDRIntraSubnets [6]string = [6]string{
		"10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24",
		"10.0.204.0/24", "10.0.205.0/24", "10.0.206.0/24"}
	DefaultAvailabilityZones [3]string = [3]string{
		"us-east-1b", "us-east-1c", "us-east-1d"}
	DefaultRegion string = "us-east-1"
)

type natgatewayType string

var (
	NONE   natgatewayType = "none"
	SINGLE natgatewayType = "single"
	ALL    natgatewayType = "all"
)

type NetworkRequest struct {
	CIDR                string
	Name                string
	Region              string
	AvailabilityZones   []string
	PublicSubnetsCIDRs  []string
	PrivateSubnetsCIDRs []string
	IntraSubnetsCIDRs   []string
	NatGatewayType      natgatewayType
	PublicToIntra       *bool
}

type NetworkResources struct {
	VPCResources       *vpc.VPCResources
	AvailabilityZones  []string
	Region             string
	PublicSNResources  []*subnet.PublicSubnetResources
	PrivateSNResources []*subnet.PrivateSubnetResources
	IntraSNResources   []*subnet.PrivateSubnetResources
}

func DefaultNetworkRequest(name, regionName string) NetworkRequest {
	return NetworkRequest{
		Name:                name,
		CIDR:                DefaultCIDRNetwork,
		AvailabilityZones:   data.GetAvailabilityZones("")[:3],
		PublicSubnetsCIDRs:  DefaultCIDRPublicSubnets[:],
		PrivateSubnetsCIDRs: DefaultCIDRPrivateSubnets[:],
		IntraSubnetsCIDRs:   DefaultCIDRIntraSubnets[:],
		NatGatewayType:      ALL}

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
	return managePrivateSubnets(vpc, ngws, ctx, r.PrivateSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.NatGatewayType)
}

func (r NetworkRequest) manageIntraSubnets(vpc *ec2.Vpc, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PrivateSubnetResources, err error) {
	return managePrivateSubnets(vpc, nil, ctx, r.IntraSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.NatGatewayType)
}

func managePrivateSubnets(vpc *ec2.Vpc, ngws []*ec2.NatGateway, ctx *pulumi.Context,
	snsCIDRs, azs []string, name, namePrefix string, ngwType natgatewayType) (subnets []*subnet.PrivateSubnetResources, err error) {
	if len(snsCIDRs) > 0 {
		for i := 0; i < len(snsCIDRs); i++ {
			privateSNRequest :=
				subnet.PrivateSubnetRequest{
					VPC:              vpc,
					NatGateway:       getNatGateway(ngwType, ngws, i),
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
	return r.NatGatewayType != NONE || r.NatGatewayType == SINGLE && i == 0 || len(r.PrivateSubnetsCIDRs) > 0
}

func getNatGateways(source []*subnet.PublicSubnetResources) (ngws []*ec2.NatGateway) {
	for _, subnet := range source {
		ngws = append(ngws, subnet.NatGateway)
	}
	return
}

func getNatGateway(ngwType natgatewayType, ngws []*ec2.NatGateway, i int) *ec2.NatGateway {
	if ngws == nil {
		return nil
	}
	if ngwType == SINGLE {
		return ngws[0]
	}
	return ngws[i]
}
