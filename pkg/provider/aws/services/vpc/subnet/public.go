package subnet

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
)

type PublicSubnetRequest struct {
	VPC              *ec2.Vpc
	InternetGateway  *ec2.InternetGateway
	CIDR             string
	Region           string
	AvailabilityZone string
	Name             string
	AddNatGateway    bool
	MapPublicIp      bool
}

type PublicSubnetResources struct {
	Subnet                *ec2.Subnet
	RouteTable            *ec2.RouteTable
	RouteTableAssociation *ec2.RouteTableAssociation
	NatGateway            *ec2.NatGateway
	NatGatewayEip         *ec2.Eip
}

func (r PublicSubnetRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*PublicSubnetResources, error) {
	snName := fmt.Sprintf("%s-%s", "subnet", r.Name)
	sn, err := ec2.NewSubnet(ctx,
		snName,
		&ec2.SubnetArgs{
			VpcId:               r.VPC.ID(),
			CidrBlock:           pulumi.String(r.CIDR),
			AvailabilityZone:    pulumi.String(r.AvailabilityZone),
			Tags:                mCtx.ResourceTags(),
			MapPublicIpOnLaunch: pulumi.Bool(r.MapPublicIp),
		})
	if err != nil {
		return nil, err
	}
	var nEip *ec2.Eip
	var n *ec2.NatGateway
	if r.AddNatGateway {
		nEip, err := ec2.NewEip(ctx,
			fmt.Sprintf("%s-%s", "eip", r.Name),
			&ec2.EipArgs{
				Domain: pulumi.String("vpc"),
			})
		if err != nil {
			return nil, err
		}
		nName := fmt.Sprintf("%s-%s", "natgateway", r.Name)
		n, err = ec2.NewNatGateway(ctx,
			nName,
			&ec2.NatGatewayArgs{
				AllocationId: nEip.ID(),
				SubnetId:     sn.ID(),
				Tags:         mCtx.ResourceTags(),
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
			Tags: mCtx.ResourceTags(),
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
	// Manage endpoints
	err = endpoints(ctx, r.Name, r.Region, r.VPC, sn, rt)
	if err != nil {
		return nil, err
	}
	return &PublicSubnetResources{
			Subnet:                sn,
			RouteTable:            rt,
			RouteTableAssociation: rta,
			NatGateway:            n,
			NatGatewayEip:         nEip},
		nil
}

func endpoints(ctx *pulumi.Context, name, region string,
	vpc *ec2.Vpc, sn *ec2.Subnet, rt *ec2.RouteTable) error {
	sg, err := ec2.NewSecurityGroup(ctx,
		fmt.Sprintf("%s-%s", "endpoints", name),
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(443),
					ToPort:     pulumi.Int(443),
					CidrBlocks: pulumi.StringArray{vpc.CidrBlock},
				},
			},
		})
	if err != nil {
		return err
	}
	_, err = ec2.NewVpcEndpoint(ctx,
		fmt.Sprintf("%s-%s", "endpoint-s3", name),
		&ec2.VpcEndpointArgs{
			VpcId:           vpc.ID(),
			ServiceName:     pulumi.Sprintf("com.amazonaws.%s.s3", region),
			VpcEndpointType: pulumi.String("Gateway"),
			RouteTableIds:   pulumi.StringArray{rt.ID()},
		})
	if err != nil {
		return err
	}
	_, err = ec2.NewVpcEndpoint(ctx,
		fmt.Sprintf("%s-%s", "endpoint-ecr", name),
		&ec2.VpcEndpointArgs{
			VpcId:            vpc.ID(),
			ServiceName:      pulumi.Sprintf("com.amazonaws.%s.ecr.dkr", region),
			VpcEndpointType:  pulumi.String("Interface"),
			SubnetIds:        pulumi.StringArray{sn.ID()},
			SecurityGroupIds: pulumi.StringArray{sg.ID()},
		})
	if err != nil {
		return err
	}
	_, err = ec2.NewVpcEndpoint(ctx,
		fmt.Sprintf("%s-%s", "endpoint-ssm", name),
		&ec2.VpcEndpointArgs{
			VpcId:            vpc.ID(),
			ServiceName:      pulumi.Sprintf("com.amazonaws.%s.ssm", region),
			VpcEndpointType:  pulumi.String("Interface"),
			SubnetIds:        pulumi.StringArray{sn.ID()},
			SecurityGroupIds: pulumi.StringArray{sg.ID()},
		})
	if err != nil {
		return err
	}
	return nil
}
