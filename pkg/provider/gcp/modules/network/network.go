package network

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	cidrSubnet = "10.0.1.0/24"
)

type NetworkArgs struct {
	Prefix string
	ID     string
	Region string
	Zone   string
}

type NetworkResult struct {
	Network  *compute.Network
	Subnet   *compute.Subnetwork
	Firewall *compute.Firewall
	PublicIP *compute.Address
}

func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error) {
	nr := &NetworkResult{}

	// VPC Network
	networkName := resourcesUtil.GetResourceName(args.Prefix, args.ID, "vpc")
	n, err := compute.NewNetwork(ctx, networkName, &compute.NetworkArgs{
		AutoCreateSubnetworks: pulumi.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	nr.Network = n

	// Subnet
	subnetName := resourcesUtil.GetResourceName(args.Prefix, args.ID, "sn")
	sn, err := compute.NewSubnetwork(ctx, subnetName, &compute.SubnetworkArgs{
		Network:     n.ID(),
		IpCidrRange: pulumi.String(cidrSubnet),
		Region:      pulumi.String(args.Region),
	})
	if err != nil {
		return nil, err
	}
	nr.Subnet = sn

	// Firewall — allow SSH ingress
	fwName := resourcesUtil.GetResourceName(args.Prefix, args.ID, "fw")
	fw, err := compute.NewFirewall(ctx, fwName, &compute.FirewallArgs{
		Network:   n.SelfLink,
		Direction: pulumi.String("INGRESS"),
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("tcp"),
				Ports:    pulumi.StringArray{pulumi.String("22")},
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
		},
		TargetTags: pulumi.StringArray{
			pulumi.String(fmt.Sprintf("%s-ssh", args.ID)),
		},
	})
	if err != nil {
		return nil, err
	}
	nr.Firewall = fw

	// Static external IP
	ipName := resourcesUtil.GetResourceName(args.Prefix, args.ID, "ip")
	ip, err := compute.NewAddress(ctx, ipName, &compute.AddressArgs{
		Region:      pulumi.String(args.Region),
		AddressType: pulumi.String("EXTERNAL"),
		NetworkTier: pulumi.String("PREMIUM"),
	})
	if err != nil {
		return nil, err
	}
	nr.PublicIP = ip

	return nr, nil
}
