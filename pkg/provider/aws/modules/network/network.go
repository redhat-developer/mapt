package network

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	bastion "github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	na "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/airgap"
	ns "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
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
	ServiceEndpoints        []string
	// VpcID deploys into an existing VPC instead of creating one.
	// Airgap is not supported when VpcID is set.
	VpcID *string
}

type NetworkResult struct {
	Vpc                         *ec2.Vpc
	Subnet                      *ec2.Subnet
	SubnetRouteTableAssociation *ec2.RouteTableAssociation
	Eip                         *ec2.Eip
	LoadBalancer                *lb.LoadBalancer
	// If Airgap true on args
	Bastion *bastion.BastionResult
	// IsPublic is false when the selected subnet has no internet gateway route
	// (private subnet in an existing VPC). In that case no EIP or LB is created;
	// the machine connects outbound only and SSH readiness checks are skipped.
	IsPublic bool
}

func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error) {
	var err error
	var result *NetworkResult
	switch {
	case args.VpcID != nil:
		result, err = existingVPCNetwork(ctx, mCtx, args)
		if err != nil {
			return nil, err
		}
	case args.Airgap:
		var publicSubnet *ec2.Subnet
		result, publicSubnet, err = airgapNetworking(ctx, mCtx, args)
		if err != nil {
			return nil, err
		}
		result.IsPublic = true
		result.Bastion, err = bastion.Create(ctx, mCtx,
			&bastion.BastionArgs{
				Prefix: args.Prefix,
				VPC:    result.Vpc,
				Subnet: publicSubnet,
			})
		if err != nil {
			return nil, err
		}
	default:
		result, err = standardNetwork(ctx, mCtx, args)
		if err != nil {
			return nil, err
		}
		result.IsPublic = true
	}
	// EIP: only for truly public, non-airgap deployments.
	// Airgap machines are private (reachable only via bastion); the internal LB
	// does not need an EIP. Private-VPC deployments have no public access at all.
	if result.IsPublic && !args.Airgap {
		result.Eip, err = ec2.NewEip(ctx,
			resourcesUtil.GetResourceName(args.Prefix, args.ID, "lbeip"),
			&ec2.EipArgs{
				Tags: mCtx.ResourceTags(),
			})
		if err != nil {
			return nil, err
		}
	}
	// LB: created for any public deployment that requests one.
	// Public deployments attach the EIP; airgap deployments get an internal LB (no EIP).
	if args.CreateLoadBalancer && result.IsPublic {
		lba := &loadBalancerArgs{
			prefix: &args.Prefix,
			id:     &args.ID,
			mCtx:   mCtx,
			subnet: result.Subnet,
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

func existingVPCNetwork(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error) {
	subnetID, err := data.GetPublicSubnetIDInAZ(ctx.Context(), args.Region, *args.VpcID, args.AZ)
	isPublic := true
	if err != nil {
		// No public subnet in this AZ. Fall back to any available subnet so the
		// machine can still run as an outbound-only workload (e.g. a GitLab runner).
		subnetID, err = data.GetAnySubnetIDInAZ(ctx.Context(), args.Region, *args.VpcID, args.AZ)
		if err != nil {
			return nil, err
		}
		isPublic = false
	}
	vpc, err := ec2.GetVpc(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ID, "vpc"),
		pulumi.ID(*args.VpcID), nil)
	if err != nil {
		return nil, err
	}
	subnet, err := ec2.GetSubnet(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ID, "subnet"),
		pulumi.ID(*subnetID), nil)
	if err != nil {
		return nil, err
	}
	return &NetworkResult{
		Vpc:      vpc,
		Subnet:   subnet,
		IsPublic: isPublic,
	}, nil
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
		ServiceEndpoints:          args.ServiceEndpoints,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, err
	}
	return &NetworkResult{
		Vpc:    net.VPCResources.VPC,
		Subnet: net.PublicSNResources[0].Subnet,
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
		SetAsAirgap:      args.AirgapPhaseConnectivity == OFF,
		ServiceEndpoints: args.ServiceEndpoints}.CreateNetwork(ctx, mCtx)
	if err != nil {
		return nil, nil, err
	}
	return &NetworkResult{
		Vpc:                         net.VPCResources.VPC,
		Subnet:                      net.TargetSubnet.Subnet,
		SubnetRouteTableAssociation: net.TargetSubnet.RouteTableAssociation,
	}, net.PublicSubnet.Subnet, nil
}

type loadBalancerArgs struct {
	prefix, id *string
	mCtx       *mc.Context
	subnet     *ec2.Subnet
	// If eip != nil it means it is not airgap
	eip *ec2.Eip
}

func (a *loadBalancerArgs) airgap() bool { return a.eip == nil }

func loadBalancer(ctx *pulumi.Context, args *loadBalancerArgs) (*lb.LoadBalancer, error) {
	lbArgs := &lb.LoadBalancerArgs{
		LoadBalancerType:         pulumi.String("network"),
		EnableDeletionProtection: pulumi.Bool(false),
		Tags:                     args.mCtx.ResourceTags(),
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
		snMapping.PrivateIpv4Address = pulumi.String(*internalLBIp)
		lbArgs.Internal = pulumi.Bool(true)
	} else {
		// It load balancer is public facing
		snMapping.AllocationId = args.eip.ID()
	}
	lb, err := lb.NewLoadBalancer(ctx,
		resourcesUtil.GetResourceName(*args.prefix, *args.id, "lb"),
		lbArgs)
	if err != nil {
		return nil, err
	}
	return lb, nil
}
