package securitygroup

import (
	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type IngressRules struct {
	Description string
	FromPort    int
	ToPort      int
	Protocol    string
}

type SGRequest struct {
	Name         string
	Description  string
	IngressRules []IngressRules
	VPC          *ec2.Vpc
}

type SGResources struct {
	SG *ec2.SecurityGroup
}

func (r SGRequest) Create(ctx *pulumi.Context) (*SGResources, error) {
	sg, err := ec2.NewSecurityGroup(ctx,
		r.Name,
		&ec2.SecurityGroupArgs{
			Description: pulumi.String(r.Description),
			VpcId:       r.VPC.ID(),
			Ingress:     getSecurityGroupIngressArray(r.IngressRules),
			Egress:      ec2.SecurityGroupEgressArray{egressAll},
			Tags: pulumi.StringMap{
				"Name": pulumi.String(r.Name),
			},
		})
	if err != nil {
		return nil, err
	}
	return &SGResources{SG: sg}, nil
}

func getSecurityGroupIngressArray(rules []IngressRules) (sgia ec2.SecurityGroupIngressArray) {
	for _, r := range rules {
		sgia = append(sgia, &ec2.SecurityGroupIngressArgs{
			Description: pulumi.String(r.Description),
			FromPort:    pulumi.Int(r.FromPort),
			ToPort:      pulumi.Int(r.ToPort),
			Protocol:    pulumi.String(r.Protocol),
			CidrBlocks: pulumi.StringArray{
				pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
			},
			Ipv6CidrBlocks: pulumi.StringArray{
				pulumi.String(infra.NETWORKING_CIDR_ANY_IPV6),
			},
		})
	}
	return sgia
}
