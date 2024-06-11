package securitygroup

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
)

type IngressRules struct {
	Description string
	FromPort    int
	ToPort      int
	Protocol    string
	CidrBlocks  string
	SG          *ec2.SecurityGroup
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
			Tags:        maptContext.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	return &SGResources{SG: sg}, nil
}

func getSecurityGroupIngressArray(rules []IngressRules) (sgia ec2.SecurityGroupIngressArray) {
	for _, r := range rules {
		args := &ec2.SecurityGroupIngressArgs{
			Description: pulumi.String(r.Description),
			FromPort:    pulumi.Int(r.FromPort),
			ToPort:      pulumi.Int(r.ToPort),
			Protocol:    pulumi.String(r.Protocol),
		}
		if r.SG != nil {
			args.SecurityGroups = pulumi.StringArray{r.SG.ID()}
		} else if len(r.CidrBlocks) > 0 {
			args.CidrBlocks = pulumi.StringArray{pulumi.String(r.CidrBlocks)}
		} else {
			args.CidrBlocks = pulumi.StringArray{pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4)}
		}
		sgia = append(sgia, args)
	}
	return sgia
}
