package subnet

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
)

type PrivateSubnetRequest struct {
	VPC        *ec2.Vpc
	Subnet     *ec2.Subnet
	NatGateway *ec2.NatGateway

	CIDR             string
	AvailabilityZone string
	Name             string
}

type PrivateSubnetResources struct {
	Subnet                *ec2.Subnet
	RouteTable            *ec2.RouteTable
	RouteTableAssociation *ec2.SubnetRouteTableAssociation
	Route                 *ec2.Route
}

func (r PrivateSubnetRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*PrivateSubnetResources, error) {
	snName := fmt.Sprintf("%s-%s", "subnet", r.Name)
	sn, err := ec2.NewSubnet(ctx,
		snName,
		&ec2.SubnetArgs{
			VpcId:            r.VPC.ID(),
			CidrBlock:        pulumi.String(r.CIDR),
			AvailabilityZone: pulumi.String(r.AvailabilityZone),
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}
	rtName := fmt.Sprintf("%s-%s", "routeTable", r.Name)
	rt, err := ec2.NewRouteTable(ctx,
		rtName,
		&ec2.RouteTableArgs{
			VpcId: r.VPC.ID(),
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}
	rta, err := ec2.NewSubnetRouteTableAssociation(ctx,
		fmt.Sprintf("%s-%s", "routeTableAssociation", r.Name),
		&ec2.SubnetRouteTableAssociationArgs{
			SubnetId:     sn.ID(),
			RouteTableId: rt.ID(),
		})
	if err != nil {
		return nil, err
	}

	// Create route if NAT gateway is provided
	var route *ec2.Route
	if r.NatGateway != nil {
		route, err = ec2.NewRoute(ctx,
			fmt.Sprintf("%s-%s", "route", r.Name),
			&ec2.RouteArgs{
				RouteTableId:         rt.ID(),
				DestinationCidrBlock: pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
				NatGatewayId:         r.NatGateway.ID(),
			})
		if err != nil {
			return nil, err
		}
	}

	return &PrivateSubnetResources{
			Subnet:                sn,
			RouteTable:            rt,
			RouteTableAssociation: rta,
			Route:                 route},
		nil
}

