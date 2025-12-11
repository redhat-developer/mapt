package network

import (
	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	cidrVN = "10.0.0.0/16"
	cidrSN = "10.0.2.0/24"
)

type NetworkArgs struct {
	Prefix      string
	ComponentID string
	Name        string
	RG          *ibmcloud.ResourceGroup
	Zone        *string
}

type Network struct {
	VPC           *ibmcloud.IsVpc
	Subnet        *ibmcloud.IsSubnet
	SecurityGroup *ibmcloud.IsSecurityGroup
	Floatingip    *ibmcloud.IsFloatingIp
}

func New(ctx *pulumi.Context, args *NetworkArgs) (*Network, error) {
	vpc, err := ibmcloud.NewIsVpc(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "isvpc"),
		&ibmcloud.IsVpcArgs{
			Name:          pulumi.String(args.Name),
			ResourceGroup: args.RG.ID(),
		})
	if err != nil {
		return nil, err
	}
	vpcap, err := ibmcloud.NewIsVpcAddressPrefix(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "isvpcaddpre"),
		&ibmcloud.IsVpcAddressPrefixArgs{
			Vpc:  vpc.ID(),
			Zone: pulumi.String(*args.Zone),
			Cidr: pulumi.String(cidrVN),
			Name: pulumi.String(args.Name),
		})
	if err != nil {
		return nil, err
	}
	subnet, err := ibmcloud.NewIsSubnet(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "issubnet"),
		&ibmcloud.IsSubnetArgs{
			Name:          pulumi.String(args.Name),
			Vpc:           vpc.ID(),
			Zone:          pulumi.String(*args.Zone),
			Ipv4CidrBlock: pulumi.String(cidrSN),
			ResourceGroup: args.RG.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{vpcap}))
	if err != nil {
		return nil, err
	}
	publicGateway, err := ibmcloud.NewIsPublicGateway(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "pgw"),
		&ibmcloud.IsPublicGatewayArgs{
			Name:          pulumi.String(args.Name),
			Vpc:           vpc.ID(),
			Zone:          pulumi.String(*args.Zone),
			ResourceGroup: args.RG.ID(),
		})
	if err != nil {
		return nil, err
	}
	_, err = ibmcloud.NewIsSubnetPublicGatewayAttachment(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "pgwa"),
		&ibmcloud.IsSubnetPublicGatewayAttachmentArgs{
			Subnet:        subnet.ID(),
			PublicGateway: publicGateway.ID(),
		})
	if err != nil {
		return nil, err
	}
	securityGroup, err := ibmcloud.NewIsSecurityGroup(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "sg"),
		&ibmcloud.IsSecurityGroupArgs{
			Name:          pulumi.String(args.Name),
			Vpc:           vpc.ID(),
			ResourceGroup: args.RG.ID(),
		})
	if err != nil {
		return nil, err
	}
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "ssh"),
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     securityGroup.ID(),
			Direction: pulumi.String("inbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
			Tcp: &ibmcloud.IsSecurityGroupRuleTcpArgs{
				PortMin: pulumi.Int(22),
				PortMax: pulumi.Int(22),
			},
		})
	if err != nil {
		return nil, err
	}
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "outb"),
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     securityGroup.ID(),
			Direction: pulumi.String("outbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
		},
	)
	if err != nil {
		return nil, err
	}
	fip, err := ibmcloud.NewIsFloatingIp(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "fip"),
		&ibmcloud.IsFloatingIpArgs{
			Name:          pulumi.String(args.Name),
			Zone:          pulumi.String(*args.Zone),
			ResourceGroup: args.RG.ID(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &Network{
		VPC:           vpc,
		Subnet:        subnet,
		SecurityGroup: securityGroup,
		Floatingip:    fip}, nil
}
