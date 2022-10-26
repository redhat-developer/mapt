package compute

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	supportMatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
)

type Request struct {
	ProjecName        string
	Specs             *supportMatrix.SupportedHost
	Public            bool
	BastionSG         *ec2.SecurityGroup
	KeyPair           *ec2.KeyPair
	VPC               *ec2.Vpc
	AvailabilityZones []string
	Subnets           []*ec2.Subnet
	SpotPrice         string
}

type Resources struct {
	Specs      *supportMatrix.SupportedHost
	Name       string
	Username   string
	Instance   *ec2.Instance
	InstanceIP pulumi.StringOutput
	AWSKeyPair *ec2.KeyPair
	SG         *ec2.SecurityGroup
	// contains value if key is created within this module
	PrivateKey *tls.PrivateKey
}

type ComputeRequest interface {
	GetName() string
	manageKeypair(ctx *pulumi.Context, result *Resources) error
	manageSecurityGroup(ctx *pulumi.Context, result *Resources) error
	createOnDemand(ctx *pulumi.Context, amiID string,
		dh *ec2.DedicatedHost, result *Resources) error
	createSpotInstance(ctx *pulumi.Context, amiID string, result *Resources) error
}

type ComputeRequestType interface {
	GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error)
	GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error)
	GetPostScript() (string, error)
}

type Compute interface {
	remoteExec(ctx *pulumi.Context, cmdName, cmd string) error
	OutputPrivateKey() string
	OutputHost() string
	OutputUsername() string
}
