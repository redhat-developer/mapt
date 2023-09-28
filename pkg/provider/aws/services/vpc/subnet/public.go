package subnet

import (
	"fmt"

	infra "github.com/adrianriobo/qenvs/pkg/provider"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PublicSubnetRequest struct {
	VPC              *ec2.Vpc
	InternetGateway  *ec2.InternetGateway
	CIDR             string
	AvailabilityZone string
	Name             string
	AddNatGateway    bool
}

type PublicSubnetResources struct {
	Subnet                *ec2.Subnet
	RouteTable            *ec2.RouteTable
	RouteTableAssociation *ec2.RouteTableAssociation
	EIP                   *ec2.Eip
	NatGateway            *ec2.NatGateway
}

func (r PublicSubnetRequest) Create(ctx *pulumi.Context) (*PublicSubnetResources, error) {
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
	eipName := fmt.Sprintf("%s-%s", "eip", r.Name)
	eip, err := ec2.NewEip(ctx,
		eipName,
		&ec2.EipArgs{
			Vpc: pulumi.Bool(true),
		})
	if err != nil {
		return nil, err
	}
	var n *ec2.NatGateway
	if r.AddNatGateway {
		nName := fmt.Sprintf("%s-%s", "natgateway", r.Name)
		n, err = ec2.NewNatGateway(ctx,
			nName,
			&ec2.NatGatewayArgs{
				AllocationId: eip.ID(),
				SubnetId:     sn.ID(),
				Tags: pulumi.StringMap{
					"Name": pulumi.String(nName),
				},
			})
		if err != nil {
			return nil, err
		}
	}
	rtName := fmt.Sprintf("%s-%s", "routeTable", r.Name)
	rt, err := ec2.NewRouteTable(ctx,
		rtName,
		&ec2.RouteTableArgs{
			VpcId: r.VPC.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
					GatewayId: r.InternetGateway.ID(),
				},
			},
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
	return &PublicSubnetResources{
			Subnet:                sn,
			RouteTable:            rt,
			RouteTableAssociation: rta,
			EIP:                   eip,
			NatGateway:            n},
		nil
}
