package bastion

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
)

type BastionRequest struct {
	Name          string
	HA            bool
	keyPair       *ec2.KeyPair
	VPC           *ec2.Vpc
	PublicSubnets []*ec2.Subnet
	// loadBalancer *lb.LoadBalancer
}

type BastionResources struct {
	LaunchTemplate *ec2.LaunchTemplate
	Instance       *ec2.Instance
	SG             *ec2.SecurityGroup
	AWSKeyPair     *ec2.KeyPair
	// contains value if key is created within this module
	PrivateKey *tls.PrivateKey
}
