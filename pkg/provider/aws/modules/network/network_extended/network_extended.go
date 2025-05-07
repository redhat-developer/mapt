package network_extended

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
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
	AZ     []string
	// Create a load balancer
	// If !airgap lb will be public facing
	// If airgap lb will be internal
	CreateLoadBalancer      *bool
	Airgap                  bool
	AirgapPhaseConnectivity Connectivity
}

func (r *NetworkRequest) Network(ctx *pulumi.Context) (
	vpc *ec2.Vpc,
	targetSubnet []*ec2.Subnet,
	targetRouteTableAssociation *ec2.RouteTableAssociation,
	b *bastion.Bastion,
	lb *lb.LoadBalancer,
	err error) {
	if !r.Airgap {
		vpc, targetSubnet, err = r.manageNetworking(ctx)
	} else {
		var publicSubnet *ec2.Subnet
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
		lb, err = r.createLoadBalancer(ctx, targetSubnet[0])
	}
	return
}

// Create a standard network
func (r *NetworkRequest) manageNetworking(ctx *pulumi.Context) (*ec2.Vpc, []*ec2.Subnet, error) {
	net, err := ns.NetworkRequest{
		CIDR:               cidrVN,
		Name:               resourcesUtil.GetResourceName(r.Prefix, r.ID, "net"),
		Region:             r.Region,
		AvailabilityZones:  r.AZ,
		PublicSubnetsCIDRs: []string{cidrPublicSN},
		SingleNatGateway:   true,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, nil, err
	}

	var subnets []*ec2.Subnet
	for _, sn := range net.PublicSNResources {
		subnets = append(subnets, sn.Subnet)
	}
	return net.VPCResources.VPC, subnets, nil
}

func (r *NetworkRequest) createLoadBalancer(ctx *pulumi.Context,
	subnet *ec2.Subnet) (*lb.LoadBalancer, error) {
	lbArgs := &lb.LoadBalancerArgs{
		LoadBalancerType: pulumi.String("network"),
	}
	snMapping := &lb.LoadBalancerSubnetMappingArgs{
		SubnetId: subnet.ID()}
	lbArgs.SubnetMappings = lb.LoadBalancerSubnetMappingArray{
		snMapping,
	}
	// If airgap the load balancer is internal facing
	if r.Airgap {
		snMapping.PrivateIpv4Address = pulumi.String(internalLBIp)
		lbArgs.Internal = pulumi.Bool(true)
	}
	return lb.NewLoadBalancer(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "lb"), lbArgs)
}
