package securitygroup

import (
	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Pick ideas from
// https://github.com/terraform-aws-modules/terraform-aws-security-group/blob/master/rules.tf
var SSH_TCP = IngressRules{Description: "SSH", FromPort: 22, ToPort: 22, Protocol: "tcp"}

var egressAll = &ec2.SecurityGroupEgressArgs{
	FromPort: pulumi.Int(0),
	ToPort:   pulumi.Int(0),
	Protocol: pulumi.String("-1"),
	CidrBlocks: pulumi.StringArray{
		pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
	},
	Ipv6CidrBlocks: pulumi.StringArray{
		pulumi.String(infra.NETWORKING_CIDR_ANY_IPV6),
	}}
