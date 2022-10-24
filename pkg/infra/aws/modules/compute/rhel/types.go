package rhel

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type RHELRequest struct {
	Name         string
	VersionMajor string
	SpotPrice    string
	Public       bool
	BastionSG    *ec2.SecurityGroup
	keyPair      *ec2.KeyPair
	VPC          *ec2.Vpc
	Subnets      []*ec2.Subnet
}

type RHELResources struct {
	// InstanceID string
	Instance   *ec2.Instance
	InstanceIP pulumi.StringOutput
	AWSKeyPair *ec2.KeyPair
	SG         *ec2.SecurityGroup
	// contains value if key is created within this module
	PrivateKey *tls.PrivateKey
}
