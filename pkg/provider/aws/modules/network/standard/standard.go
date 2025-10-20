package standard

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
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

	DefaultLBIPs [3]string = [3]string{
		"10.0.1.15", "10.0.2.15", "10.0.3.15"}
	DefaultAvailabilityZones [3]string = [3]string{
		"us-east-1b", "us-east-1c", "us-east-1d"}
	DefaultRegion string = "us-east-1"
)

// GeneratePublicSubnetCIDRs generates CIDR blocks for public subnets based on the number of availability zones
func GeneratePublicSubnetCIDRs(azCount int) []string {
	return generateCIDRBlocks(DefaultCIDRNetwork, azCount, 1)
}

// GeneratePrivateSubnetCIDRs generates CIDR blocks for private subnets based on the number of availability zones
func GeneratePrivateSubnetCIDRs(azCount int) []string {
	return generateCIDRBlocks(DefaultCIDRNetwork, azCount, 101)
}

// GenerateIntraSubnetCIDRs generates CIDR blocks for intra subnets based on the number of availability zones
func GenerateIntraSubnetCIDRs(azCount int) []string {
	return generateCIDRBlocks(DefaultCIDRNetwork, azCount, 201)
}

// generateCIDRBlocks generates CIDR blocks for a given number of subnets
func generateCIDRBlocks(baseCIDR string, count int, offset int) []string {
	cidrs := make([]string, count)
	for i := 0; i < count; i++ {
		cidrs[i] = fmt.Sprintf("10.0.%d.0/24", offset+i)
	}
	return cidrs
}

type NatGatewayMode string

var (
	NatGatewayModeNone   NatGatewayMode = "none"
	NatGatewayModeSingle NatGatewayMode = "single"
	NatGatewayModeHA     NatGatewayMode = "ha"
	NatGatewayModeCustom NatGatewayMode = "ha"
)

type NetworkRequest struct {
	MCtx                *mc.Context
	CIDR                string
	Name                string
	Region              string
	AvailabilityZones   []string
	PublicSubnetsCIDRs  []string
	PrivateSubnetsCIDRs []string
	IntraSubnetsCIDRs   []string
	NatGatewayMode      *NatGatewayMode
	PublicToIntra       *bool
	MapPublicIp         bool
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
	azs := data.GetAvailabilityZones("", nil)[:3]
	azCount := len(azs)
	return NetworkRequest{
		Name:                name,
		CIDR:                DefaultCIDRNetwork,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  GeneratePublicSubnetCIDRs(azCount),
		PrivateSubnetsCIDRs: GeneratePrivateSubnetCIDRs(azCount),
		IntraSubnetsCIDRs:   GenerateIntraSubnetCIDRs(azCount),
		NatGatewayMode:      &NatGatewayModeSingle,
		MapPublicIp:         false,
	}
}

func (r NetworkRequest) CreateNetwork(ctx *pulumi.Context) (*NetworkResources, error) {
	// Data validation
	if err := r.validate(); err != nil {
		return nil, err
	}
	// VPC creation
	vpcRequest := vpc.VPCRequest{CIDR: r.CIDR, Name: r.Name}
	vpcResult, err := vpcRequest.CreateNetwork(ctx, r.MCtx)
	if err != nil {
		return nil, err
	}
	ctx.Export(StackCreateNetworkOutputVPCID, vpcResult.VPC.ID())
	// Manage Public Subnets
	publicSNResults, err :=
		r.managePublicSubnets(r.MCtx, vpcResult.VPC, vpcResult.InternetGateway, ctx, "public")
	if err != nil {
		return nil, err
	}
	// Manage Private Subnets
	privateSNResults, err :=
		r.managePrivateSubnets(r.MCtx, vpcResult.VPC, getNatGateways(publicSNResults), ctx, "private")
	if err != nil {
		return nil, err
	}
	// Manage Intra Subnets
	intraSNResults, err :=
		r.manageIntraSubnets(r.MCtx, vpcResult.VPC, ctx, "intra")
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
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(r); err != nil {
		return err
	}
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

func (r NetworkRequest) managePublicSubnets(mCtx *mc.Context, vpc *ec2.Vpc,
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
					AddNatGateway:    r.checkIfNatGatewayRequired(i),
					MapPublicIp:      r.MapPublicIp,
				}
			subnet, err := publicSNRequest.Create(ctx, mCtx)
			if err != nil {
				return nil, err
			}
			subnets = append(subnets, subnet)
		}
	}
	return
}

func (r NetworkRequest) managePrivateSubnets(mCtx *mc.Context, vpc *ec2.Vpc,
	ngws []*ec2.NatGateway, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PrivateSubnetResources, err error) {
	return managePrivateSubnets(mCtx, vpc, ngws, ctx, r.PrivateSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.NatGatewayMode == &NatGatewayModeSingle)
}

func (r NetworkRequest) manageIntraSubnets(mCtx *mc.Context, vpc *ec2.Vpc, ctx *pulumi.Context, namePrefix string) (subnets []*subnet.PrivateSubnetResources, err error) {
	return managePrivateSubnets(mCtx, vpc, nil, ctx, r.IntraSubnetsCIDRs, r.AvailabilityZones, r.Name, namePrefix, r.NatGatewayMode == &NatGatewayModeSingle)
}

func managePrivateSubnets(mCtx *mc.Context, vpc *ec2.Vpc, ngws []*ec2.NatGateway, ctx *pulumi.Context,
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
			subnet, err := privateSNRequest.Create(ctx, mCtx)
			if err != nil {
				return nil, err
			}
			subnets = append(subnets, subnet)
		}
	}
	return
}

func (r NetworkRequest) checkIfNatGatewayRequired(i int) bool {
	return r.NatGatewayMode == &NatGatewayModeSingle && i == 0 || len(r.PrivateSubnetsCIDRs) > 0
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
