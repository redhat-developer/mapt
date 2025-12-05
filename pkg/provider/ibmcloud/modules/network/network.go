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
	subnet, err := ibmcloud.NewIsSubnet(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "issubnet"),
		&ibmcloud.IsSubnetArgs{
			Name:          pulumi.String(args.Name),
			Vpc:           vpc.ID(),
			Zone:          pulumi.String(*args.Zone),
			Ipv4CidrBlock: pulumi.String(cidrSN),
			ResourceGroup: args.RG.ID(),
		})
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
		"pgw-attachment",
		&ibmcloud.IsSubnetPublicGatewayAttachmentArgs{
			Subnet:        subnet.ID(),
			PublicGateway: publicGateway.ID(),
		})
	if err != nil {
		return nil, err
	}
	securityGroup, err := ibmcloud.NewIsSecurityGroup(ctx,
		"power11-sg",
		&ibmcloud.IsSecurityGroupArgs{
			Name:          pulumi.String("power11-security-group"),
			Vpc:           vpc.ID(),
			ResourceGroup: args.RG.ID(),
		})
	if err != nil {
		return nil, err
	}
	// Security Rules - SSH (custom port 2222)
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx, "allow-ssh",
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     securityGroup.ID(),
			Direction: pulumi.String("inbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
			Tcp: &ibmcloud.IsSecurityGroupRuleTcpArgs{
				PortMin: pulumi.Int(2222),
				PortMax: pulumi.Int(2222),
			},
		})
	if err != nil {
		return nil, err
	}
	// if err != nil {
	// 	return err
	// }
	// loadBalancer, err := ibmcloud.NewIsLb(ctx, "power11-lb", &ibmcloud.IsLbArgs{
	// 	Name: pulumi.String("power11-alb"),
	// 	Subnets: pulumi.StringArray{
	// 		lbSubnet.ID(),
	// 	},
	// 	Type:          pulumi.String("public"),
	// 	ResourceGroup: resourceGroup.ID(),
	// 	SecurityGroups: pulumi.StringArray{
	// 		securityGroup.ID(),
	// 	},
	// 	Tags: pulumi.StringArray{
	// 		pulumi.String("power11"),
	// 		pulumi.String("public-access"),
	// 	},
	// }, pulumi.Timeouts(&pulumi.CustomTimeouts{
	// 	Create: "15m",
	// 	Update: "15m",
	// 	Delete: "15m",
	// }))

	return &Network{
		VPC:           vpc,
		Subnet:        subnet,
		SecurityGroup: securityGroup}, nil
}
