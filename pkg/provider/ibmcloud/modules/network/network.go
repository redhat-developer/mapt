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
	Tags        pulumi.StringArray
}

type Network struct {
	VPC           *ibmcloud.IsVpc
	Subnet        *ibmcloud.IsSubnet
	SecurityGroup *ibmcloud.IsSecurityGroup
	Floatingip    *ibmcloud.IsFloatingIp
}

// SecurityGroupArgs defines the inputs for NewSecurityGroupWithSSH.
type SecurityGroupArgs struct {
	Prefix      string
	ComponentID string
	Name        string
	VPC         pulumi.StringInput
	RG          *ibmcloud.ResourceGroup
	Tags        pulumi.StringArray
}

// NewSecurityGroupWithSSH creates a security group with inbound SSH (port 22)
// and unrestricted outbound rules. RG may be nil to use the account default.
func NewSecurityGroupWithSSH(ctx *pulumi.Context, args *SecurityGroupArgs) (*ibmcloud.IsSecurityGroup, error) {
	sgArgs := &ibmcloud.IsSecurityGroupArgs{
		Name: pulumi.String(args.Name),
		Vpc:  args.VPC,
		Tags: args.Tags,
	}
	if args.RG != nil {
		sgArgs.ResourceGroup = args.RG.ID()
	}
	sg, err := ibmcloud.NewIsSecurityGroup(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "sg"),
		sgArgs)
	if err != nil {
		return nil, err
	}
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "ssh"),
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     sg.ID(),
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
			Group:     sg.ID(),
			Direction: pulumi.String("outbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
		})
	if err != nil {
		return nil, err
	}
	return sg, nil
}

// FloatingIPArgs defines the inputs for NewFloatingIP.
type FloatingIPArgs struct {
	Prefix      string
	ComponentID string
	Name        string
	Zone        pulumi.StringInput
	RG          *ibmcloud.ResourceGroup
	Tags        pulumi.StringArray
}

// NewFloatingIP creates an IBM Cloud VPC floating IP.
// RG may be nil to use the account default resource group.
func NewFloatingIP(ctx *pulumi.Context, args *FloatingIPArgs) (*ibmcloud.IsFloatingIp, error) {
	fipArgs := &ibmcloud.IsFloatingIpArgs{
		Name: pulumi.String(args.Name),
		Zone: args.Zone,
		Tags: args.Tags,
	}
	if args.RG != nil {
		fipArgs.ResourceGroup = args.RG.ID()
	}
	return ibmcloud.NewIsFloatingIp(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "fip"),
		fipArgs)
}

func New(ctx *pulumi.Context, args *NetworkArgs) (*Network, error) {
	vpc, err := ibmcloud.NewIsVpc(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "isvpc"),
		&ibmcloud.IsVpcArgs{
			Name:          pulumi.String(args.Name),
			ResourceGroup: args.RG.ID(),
			Tags:          args.Tags,
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
			Tags:          args.Tags,
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
			Tags:          args.Tags,
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
	securityGroup, err := NewSecurityGroupWithSSH(ctx, &SecurityGroupArgs{
		Prefix:      args.Prefix,
		ComponentID: args.ComponentID,
		Name:        args.Name,
		VPC:         vpc.ID(),
		RG:          args.RG,
		Tags:        args.Tags,
	})
	if err != nil {
		return nil, err
	}
	fip, err := NewFloatingIP(ctx, &FloatingIPArgs{
		Prefix:      args.Prefix,
		ComponentID: args.ComponentID,
		Name:        args.Name,
		Zone:        pulumi.String(*args.Zone),
		RG:          args.RG,
		Tags:        args.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &Network{
		VPC:           vpc,
		Subnet:        subnet,
		SecurityGroup: securityGroup,
		Floatingip:    fip}, nil
}
