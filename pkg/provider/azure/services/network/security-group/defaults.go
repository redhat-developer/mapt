package securitygroup

// Pick ideas from
// https://github.com/terraform-aws-modules/terraform-aws-security-group/blob/master/rules.tf
var (
	SSH_PORT int = 22
	RDP_PORT int = 3389
	SSH_TCP      = IngressRules{Description: "SSH", FromPort: SSH_PORT, ToPort: SSH_PORT, Protocol: "tcp"}
	RDP_TCP      = IngressRules{Description: "RDP", FromPort: RDP_PORT, ToPort: RDP_PORT, Protocol: "tcp"}
)
