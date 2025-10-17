package securitygroup

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
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

func (r SGRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*SGResources, error) {
	// Create the security group without inline rules
	sg, err := ec2.NewSecurityGroup(ctx,
		r.Name,
		&ec2.SecurityGroupArgs{
			GroupDescription: pulumi.String(r.Description),
			VpcId:           r.VPC.ID(),
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}

	// Create ingress rules as separate resources
	for i, rule := range r.IngressRules {
		ingressArgs := &ec2.SecurityGroupIngressArgs{
			GroupId:     sg.ID(),
			Description: pulumi.String(rule.Description),
			FromPort:    pulumi.Int(rule.FromPort),
			ToPort:      pulumi.Int(rule.ToPort),
			IpProtocol:  pulumi.String(rule.Protocol),
		}
		if rule.SG != nil {
			ingressArgs.SourceSecurityGroupId = rule.SG.ID()
		} else if len(rule.CidrBlocks) > 0 {
			ingressArgs.CidrIp = pulumi.String(rule.CidrBlocks)
		} else {
			ingressArgs.CidrIp = pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4)
		}
		_, err = ec2.NewSecurityGroupIngress(ctx,
			fmt.Sprintf("%s-ingress-%d", r.Name, i),
			ingressArgs)
		if err != nil {
			return nil, err
		}
	}

	// Create default egress rule for all traffic
	_, err = ec2.NewSecurityGroupEgress(ctx,
		fmt.Sprintf("%s-egress-all", r.Name),
		&ec2.SecurityGroupEgressArgs{
			GroupId:    sg.ID(),
			FromPort:   pulumi.Int(EgressAll.FromPort),
			ToPort:     pulumi.Int(EgressAll.ToPort),
			IpProtocol: pulumi.String(EgressAll.IpProtocol),
			CidrIp:     pulumi.String(EgressAll.CidrIp),
		})
	if err != nil {
		return nil, err
	}

	return &SGResources{SG: sg}, nil
}

// getSecurityGroupIngressArray is no longer needed as we create separate ingress resources
// This function is kept for compatibility but should not be used
