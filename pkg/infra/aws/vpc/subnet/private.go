package subnet

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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
	RouteTableAssociation *ec2.RouteTableAssociation
}

func (r PrivateSubnetRequest) Create(ctx *pulumi.Context) (*PrivateSubnetResources, error) {
	snName := fmt.Sprintf("%s-%s", "subnet", r.Name)
	sn, err := ec2.NewSubnet(ctx,
		snName,
		&ec2.SubnetArgs{
			VpcId:            r.VPC.ID(),
			CidrBlock:        pulumi.String(r.CIDR),
			AvailabilityZone: pulumi.String(r.AvailabilityZone),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(snName),
			},
		})
	if err != nil {
		return nil, err
	}
	rtName := fmt.Sprintf("%s-%s", "routeTable", r.Name)
	rt, err := ec2.NewRouteTable(ctx,
		rtName,
		&ec2.RouteTableArgs{
			VpcId:  r.VPC.ID(),
			Routes: getRoutes(r.NatGateway),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(rtName),
			},
		})
	if err != nil {
		return nil, err
	}
	rta, err := ec2.NewRouteTableAssociation(ctx,
		fmt.Sprintf("%s-%s", "routeTableAssociation", r.Name),
		&ec2.RouteTableAssociationArgs{
			SubnetId:     sn.ID(),
			RouteTableId: rt.ID(),
		})
	if err != nil {
		return nil, err
	}
	return &PrivateSubnetResources{
			Subnet:                sn,
			RouteTable:            rt,
			RouteTableAssociation: rta},
		nil
}

func getRoutes(natGateway *ec2.NatGateway) ec2.RouteTableRouteArray {
	if natGateway != nil {
		return ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
				GatewayId: natGateway.ID(),
			}}
	}
	return nil
}
