package subnet

import (
	"fmt"

	vpcCommon "github.com/adrianriobo/qenvs/pkg/infra/aws/vpc"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PublicSubnetRequest struct {
	VPCID             string
	InternetGatewayID string
	CIDR              string
	AvailabilityZone  string
	Name              string
}

type PublicSubnetResources struct {
	Subnet                *ec2.Subnet
	RouteTable            *ec2.RouteTable
	RouteTableAssociation *ec2.RouteTableAssociation
	EIP                   *ec2.Eip
	NatGateway            *ec2.NatGateway
}

func (s PublicSubnetRequest) CreatePublicSubnet(ctx *pulumi.Context) (*PublicSubnetResources, error) {
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
	eip, err := ec2.NewEip(ctx,
		fmt.Sprintf("%s-%s", "eip", s.Name),
		&ec2.EipArgs{
			Vpc: pulumi.Bool(true),
		})
	if err != nil {
		return nil, err
	}
	n, err := ec2.NewNatGateway(ctx,
		fmt.Sprintf("%s-%s", "natgateway", s.Name),
		&ec2.NatGatewayArgs{
			AllocationId: eip.ID(),
			SubnetId:     sn.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(fmt.Sprintf("%s-%s", "natgateway", s.Name)),
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
					GatewayId: pulumi.String(s.InternetGatewayID),
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
	return &PublicSubnetResources{
			Subnet:                sn,
			RouteTable:            r,
			RouteTableAssociation: a,
			EIP:                   eip,
			NatGateway:            n},
		nil
}
