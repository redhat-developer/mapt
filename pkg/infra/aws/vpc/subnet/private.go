package subnet

import (
	"fmt"

	vpcCommon "github.com/adrianriobo/qenvs/pkg/infra/aws/vpc"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PrivateSubnetRequest struct {
	VPCID            string
	NatGatewayID     string
	CIDR             string
	AvailabilityZone string
	Name             string
}

type PrivateSubnetResources struct {
	Subnet                *ec2.Subnet
	RouteTable            *ec2.RouteTable
	RouteTableAssociation *ec2.RouteTableAssociation
}

func (s PrivateSubnetRequest) CreatePrivateSubnet(ctx *pulumi.Context) (*PrivateSubnetResources, error) {
	sn, err := ec2.NewSubnet(ctx,
		s.Name,
		&ec2.SubnetArgs{
			VpcId:            pulumi.String(s.VPCID),
			CidrBlock:        pulumi.String(s.CIDR),
			AvailabilityZone: pulumi.String(s.AvailabilityZone),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(s.Name),
			},
		})
	if err != nil {
		return nil, err
	}
	r, err := ec2.NewRouteTable(ctx,
		fmt.Sprintf("%s-%s", "routeTable", s.Name),
		&ec2.RouteTableArgs{
			VpcId: pulumi.String(s.VPCID),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String(vpcCommon.CIDR_ANY_IPV4),
					GatewayId: pulumi.String(s.NatGatewayID),
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String(s.Name),
			},
		})
	if err != nil {
		return nil, err
	}
	a, err := ec2.NewRouteTableAssociation(ctx,
		fmt.Sprintf("%s-%s", "routeTableAssociation", s.Name),
		&ec2.RouteTableAssociationArgs{
			SubnetId:     sn.ID(),
			RouteTableId: r.ID(),
		})
	if err != nil {
		return nil, err
	}
	return &PrivateSubnetResources{
			Subnet:                sn,
			RouteTable:            r,
			RouteTableAssociation: a},
		nil
}
