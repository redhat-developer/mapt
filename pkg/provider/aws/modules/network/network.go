package network

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	na "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/airgap"
	ns "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	cidrVN       = "10.0.0.0/16"
	cidrPublicSN = "10.0.2.0/24"
	cidrIntraSN  = "10.0.101.0/24"
	internalLBIp = "10.0.101.15"
)

type Connectivity int

const (
	ON Connectivity = iota
	OFF
)

type NetworkRequest struct {
	Prefix string
	ID     string
	Region string
	AZ     string
	// Create a load balancer
	// If !airgap lb will be public facing
	// If airgap lb will be internal
	CreateLoadBalancer      *bool
	LoadBalancerIp          bool
	Airgap                  bool
	AirgapPhaseConnectivity Connectivity
}

func (r *NetworkRequest) Network(ctx *pulumi.Context) (
	vpc *ec2.Vpc,
	targetSubnet *ec2.Subnet,
	targetRouteTableAssociation *ec2.RouteTableAssociation,
	b *bastion.Bastion,
	lb *lb.LoadBalancer,
	lbEIP *ec2.Eip,
	err error) {
	if !r.Airgap {
		vpc, targetSubnet, err = r.manageNetworking(ctx)
	} else {
		var publicSubnet *ec2.Subnet
		if vpc, publicSubnet, targetSubnet, targetRouteTableAssociation, err =
			r.manageAirgapNetworking(ctx); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		br := bastion.BastionRequest{
			Prefix: r.Prefix,
			VPC:    vpc,
			Subnet: publicSubnet,
			// private key for bastion will be exported with this key
			OutputKeyPrivateKey: fmt.Sprintf("%s-%s", r.Prefix, bastion.OutputBastionUserPrivateKey),
			OutputKeyUsername:   fmt.Sprintf("%s-%s", r.Prefix, bastion.OutputBastionUsername),
			OutputKeyHost:       fmt.Sprintf("%s-%s", r.Prefix, bastion.OutputBastionHost),
		}
		b, err = br.Create(ctx)
	}
	if r.CreateLoadBalancer != nil && *r.CreateLoadBalancer {
		lb, lbEIP, err = r.createLoadBalancer(ctx, targetSubnet)
	}
	return
}

// Create a standard network (only one public subnet)
func (r *NetworkRequest) manageNetworking(ctx *pulumi.Context) (*ec2.Vpc, *ec2.Subnet, error) {
	net, err := ns.NetworkRequest{
		CIDR:               cidrVN,
		Name:               resourcesUtil.GetResourceName(r.Prefix, r.ID, "net"),
		Region:             r.Region,
		AvailabilityZones:  []string{r.AZ},
		PublicSubnetsCIDRs: []string{cidrPublicSN},
		NatGatewayType:     ns.SINGLE,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, nil, err
	}
	return net.VPCResources.VPC,
		net.PublicSNResources[0].Subnet,
		nil
}

// Create an airgap scenario (on and off phases will be executed to remove the nat gateway on the off phase)
func (r *NetworkRequest) manageAirgapNetworking(ctx *pulumi.Context) (
	vpc *ec2.Vpc,
	publicSubnet *ec2.Subnet,
	targetSubnet *ec2.Subnet,
	targetRouteTableAssociation *ec2.RouteTableAssociation,
	err error) {
	net, err := na.AirgapNetworkRequest{
		CIDR:             cidrVN,
		Name:             resourcesUtil.GetResourceName(r.Prefix, r.ID, "net"),
		Region:           r.Region,
		AvailabilityZone: r.AZ,
		PublicSubnetCIDR: cidrPublicSN,
		TargetSubnetCIDR: cidrIntraSN,
		SetAsAirgap:      r.AirgapPhaseConnectivity == OFF}.CreateNetwork(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return net.VPCResources.VPC,
		net.PublicSubnet.Subnet,
		net.TargetSubnet.Subnet,
		net.TargetSubnet.RouteTableAssociation,
		nil
}

func (r *NetworkRequest) createLoadBalancer(ctx *pulumi.Context,
	subnet *ec2.Subnet) (*lb.LoadBalancer, *ec2.Eip, error) {
	lbArgs := &lb.LoadBalancerArgs{
		LoadBalancerType:         pulumi.String("network"),
		EnableDeletionProtection: pulumi.Bool(false),
	}
	snMapping := &lb.LoadBalancerSubnetMappingArgs{
		SubnetId: subnet.ID()}
	lbArgs.SubnetMappings = lb.LoadBalancerSubnetMappingArray{
		snMapping,
	}
	var lbEIP *ec2.Eip
	var err error
	if r.Airgap {
		// If airgap the load balancer is internal facing
		snMapping.PrivateIpv4Address = pulumi.String(internalLBIp)
		lbArgs.Internal = pulumi.Bool(true)
	} else {
		// It load balancer is public facing
		lbEIP, err = ec2.NewEip(ctx,
			resourcesUtil.GetResourceName(r.Prefix, r.ID, "lbeip"),
			&ec2.EipArgs{})
		if err != nil {
			return nil, nil, err
		}
		snMapping.AllocationId = lbEIP.ID()
	}
	lb, err := lb.NewLoadBalancer(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "lb"),
		lbArgs)
	if err != nil {
		return nil, nil, err
	}
	return lb, lbEIP, nil
}
