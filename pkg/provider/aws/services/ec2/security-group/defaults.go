package securitygroup

import (
	infra "github.com/adrianriobo/qenvs/pkg/provider"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Pick ideas from
// https://github.com/terraform-aws-modules/terraform-aws-security-group/blob/master/rules.tf
var (
	SSH_PORT int = 22
	RDP_PORT int = 3389
	SSH_TCP      = IngressRules{Description: "SSH", FromPort: SSH_PORT, ToPort: SSH_PORT, Protocol: "tcp"}
	RDP_TCP      = IngressRules{Description: "RDP", FromPort: RDP_PORT, ToPort: RDP_PORT, Protocol: "tcp"}
)

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
