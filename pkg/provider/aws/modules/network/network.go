package network

import (
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	lb "github.com/pulumi/pulumi-aws-native/sdk/go/aws/elasticloadbalancingv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	bastion "github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	na "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/airgap"
	ns "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	utilNetwork "github.com/redhat-developer/mapt/pkg/util/network"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	cidrVN       = "10.0.0.0/16"
	cidrPublicSN = "10.0.2.0/24"
	cidrIntraSN  = "10.0.101.0/24"
)

type Connectivity int

const (
	ON Connectivity = iota
	OFF
)

type NetworkArgs struct {
	Prefix string
	ID     string
	Region string
	AZ     string
	// Create a load balancer
	// If !airgap lb will be public facing
	// If airgap lb will be internal
	CreateLoadBalancer      bool
	Airgap                  bool
	AirgapPhaseConnectivity Connectivity
}

type NetworkResult struct {
	Vpc                         *ec2.Vpc
	Subnet                      *ec2.Subnet
	SubnetRouteTableAssociation *ec2.SubnetRouteTableAssociation
	VpcGatewayAttachment        *ec2.VpcGatewayAttachment
	Eip                         *ec2.Eip
	LoadBalancer                *lb.LoadBalancer
	// If Airgap true on args
	Bastion *bastion.BastionResult
}

func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error) {
	var err error
	result := &NetworkResult{}
	if !args.Airgap {
		result, err = standardNetwork(ctx, mCtx, args)
		if err != nil {
			return nil, err
		}
	} else {
		var publicSubnet *ec2.Subnet
		result, publicSubnet, err =
			airgapNetworking(ctx, mCtx, args)
		if err != nil {
			return nil, err
		}
		result.Bastion, err = bastion.Create(ctx, mCtx,
			&bastion.BastionArgs{
				Prefix: args.Prefix,
				VPC:    result.Vpc,
				Subnet: publicSubnet,
			})
		if err != nil {
			return nil, err
		}
	}
	result.Eip, err = ec2.NewEip(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ID, "lbeip"),
		&ec2.EipArgs{})
	if err != nil {
		return nil, err
	}
	if args.CreateLoadBalancer {
		lba := &loadBalancerArgs{
			prefix:               &args.Prefix,
			id:                   &args.ID,
			subnet:               result.Subnet,
			vpcGatewayAttachment: result.VpcGatewayAttachment,
		}
		if !args.Airgap {
			lba.eip = result.Eip
		}
		result.LoadBalancer, err = loadBalancer(ctx, lba)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func standardNetwork(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error) {
	net, err := ns.NetworkRequest{
		MCtx:               mCtx,
		CIDR:               cidrVN,
		Name:               resourcesUtil.GetResourceName(args.Prefix, args.ID, "net"),
		Region:             args.Region,
		AvailabilityZones:  []string{args.AZ},
		PublicSubnetsCIDRs: []string{cidrPublicSN},
		NatGatewayMode:     &ns.NatGatewayModeNone,
		MapPublicIp:        true,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, err
	}
	return &NetworkResult{
		Vpc:                  net.VPCResources.VPC,
		Subnet:               net.PublicSNResources[0].Subnet,
		VpcGatewayAttachment: net.VPCResources.VpcGatewayAttachment,
	}, nil
}

func airgapNetworking(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, *ec2.Subnet, error) {
	net, err := na.AirgapNetworkRequest{
		CIDR:             cidrVN,
		Name:             resourcesUtil.GetResourceName(args.Prefix, args.ID, "net"),
		Region:           args.Region,
		AvailabilityZone: args.AZ,
		PublicSubnetCIDR: cidrPublicSN,
		TargetSubnetCIDR: cidrIntraSN,
		SetAsAirgap:      args.AirgapPhaseConnectivity == OFF}.CreateNetwork(ctx, mCtx)
	if err != nil {
		return nil, nil, err
	}
	return &NetworkResult{
		Vpc:                         net.VPCResources.VPC,
		Subnet:                      net.TargetSubnet.Subnet,
		SubnetRouteTableAssociation: net.TargetSubnet.RouteTableAssociation,
		VpcGatewayAttachment:        net.VPCResources.VpcGatewayAttachment,
	}, net.PublicSubnet.Subnet, nil
}

type loadBalancerArgs struct {
	prefix, id           *string
	subnet               *ec2.Subnet
	vpcGatewayAttachment *ec2.VpcGatewayAttachment
	// If eip != nil it means it is not airgap
	eip *ec2.Eip
}

func (a *loadBalancerArgs) airgap() bool { return a.eip == nil }

func loadBalancer(ctx *pulumi.Context, args *loadBalancerArgs) (*lb.LoadBalancer, error) {
	lbArgs := &lb.LoadBalancerArgs{
		Type: pulumi.String("network"),
	}
	snMapping := &lb.LoadBalancerSubnetMappingArgs{
		SubnetId: args.subnet.ID()}
	lbArgs.SubnetMappings = lb.LoadBalancerSubnetMappingArray{
		snMapping,
	}
	if args.airgap() {
		// If airgap the load balancer is internal facing
		internalLBIp, err := utilNetwork.RandomIp(cidrIntraSN)
		if err != nil {
			return nil, err
		}
		snMapping.PrivateIPv4Address = pulumi.String(*internalLBIp)
		lbArgs.Scheme = pulumi.String("internal")
	} else {
		// It load balancer is public facing
		snMapping.AllocationId = args.eip.AllocationId
	}
	var dependsOn []pulumi.Resource
	if args.vpcGatewayAttachment != nil {
		dependsOn = append(dependsOn, args.vpcGatewayAttachment)
	}
	lb, err := lb.NewLoadBalancer(ctx,
		resourcesUtil.GetResourceName(*args.prefix, *args.id, "lb"),
		lbArgs,
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, err
	}
	return lb, nil
}
