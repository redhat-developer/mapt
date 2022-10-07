package network

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VPCRequest struct {
	CIDR string
	Name string
}

type VPCResources struct {
	VPC             *ec2.Vpc
	InternetGateway *ec2.InternetGateway
	SecurityGroup   *ec2.SecurityGroup
}

func (s VPCRequest) CreateNetwork(ctx *pulumi.Context) (*VPCResources, error) {
	vName := fmt.Sprintf("%s-%s", "vpc", s.Name)
	v, err := ec2.NewVpc(ctx, vName,
		&ec2.VpcArgs{
			CidrBlock: pulumi.String(s.CIDR),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(vName),
			},
		})
	if err != nil {
		return nil, err
	}
	iName := fmt.Sprintf("%s-%s", "igw", s.Name)
	i, err := ec2.NewInternetGateway(ctx,
		iName,
		&ec2.InternetGatewayArgs{
			VpcId: v.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(iName),
			},
		})
	if err != nil {
		return nil, err
	}
	sgName := fmt.Sprintf("%s-%s", "default", s.Name)
	sg, err := ec2.NewSecurityGroup(ctx,
		fmt.Sprintf("%s-%s", sgName, s.Name),
		&ec2.SecurityGroupArgs{
			Description: pulumi.String("Default"),
			VpcId:       v.ID(),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Self:     pulumi.BoolPtr(true),
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
					},
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String(sgName),
			},
		})
	if err != nil {
		return nil, err
	}
	return &VPCResources{
			VPC:             v,
			InternetGateway: i,
			SecurityGroup:   sg},
		nil
}
