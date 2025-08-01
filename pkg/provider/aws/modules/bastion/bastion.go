package bastion

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	instanceType = "t2.small"
	amiRegex     = "amzn2-ami-hvm-*-x86_64-ebs"
	// mzn2-ami-hvm-*-x86_64-gp2
	amiDefaultUsername = "ec2-user"
	bastionMachineID   = "bastion"

	diskSize       int = 100
	defaultSSHPort int = 22

	// Outputs
	OutputBastionUserPrivateKey = "bastion_id_rsa"
	OutputBastionUsername       = "bastion_username"
	OutputBastionHost           = "bastion_host"
)

type BastionRequest struct {
	Prefix string
	VPC    *ec2.Vpc
	Subnet *ec2.Subnet
	// Custon key for the outputs
	// in case we want to create a stack with 2 bastions
	// we need to use diff keys for outputs which should be controlled
	// outside
	OutputKeyPrivateKey string
	OutputKeyHost       string
	OutputKeyUsername   string
}

type Bastion struct {
	Instance   *ec2.Instance
	PrivateKey *tls.PrivateKey
	Usarname   string
	Port       int
}

// This module allows to create a bastion host
// It will export to context, based on keys from request:
// * private key
// * username
// * host
// It will also return the required refs to resources as BastionsResources to
// allow orchestrated within the wrapping stack
func (r *BastionRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*Bastion, error) {
	// Get AMI
	ami, err := ami.GetAMIByName(ctx, amiRegex, nil, nil)
	if err != nil {
		return nil, err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, bastionMachineID, "pk")}
	keyResources, err := kpr.Create(ctx, mCtx)
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputKeyPrivateKey, keyResources.PrivateKey.PrivateKeyPem)
	sgs, err := r.securityGroup(ctx, mCtx)
	if err != nil {
		return nil, err
	}
	i, err := r.instance(ctx, mCtx, ami, keyResources, sgs)
	if err != nil {
		return nil, err
	}
	return &Bastion{
		Instance:   i,
		PrivateKey: keyResources.PrivateKey,
		Usarname:   amiDefaultUsername,
		Port:       defaultSSHPort,
	}, nil
}

// Allow connect bastion on ssh port
func (r *BastionRequest) securityGroup(ctx *pulumi.Context, mCtx *mc.Context) (pulumi.StringArray, error) {
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(r.Prefix, bastionMachineID, "sg"),
		VPC:          r.VPC,
		Description:  fmt.Sprintf("sg for %s", bastionMachineID),
		IngressRules: []securityGroup.IngressRules{sshIngressRule},
	}.Create(ctx, mCtx)
	if err != nil {
		return nil, err
	}
	sgs := util.ArrayConvert([]*ec2.SecurityGroup{sg.SG},
		func(sg *ec2.SecurityGroup) pulumi.StringInput {
			return sg.ID()
		})
	return pulumi.StringArray(sgs[:]), nil
}

func (r *BastionRequest) instance(ctx *pulumi.Context, mCtx *mc.Context,
	ami *ec2.LookupAmiResult,
	keyResources *keypair.KeyPairResources,
	securityGroups pulumi.StringArray,
) (*ec2.Instance, error) {
	instanceArgs := ec2.InstanceArgs{
		SubnetId:                 r.Subnet.ID(),
		Ami:                      pulumi.String(ami.Id),
		InstanceType:             pulumi.String(instanceType),
		KeyName:                  keyResources.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(true),
		VpcSecurityGroupIds:      securityGroups,
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(diskSize),
		},
		Tags: mCtx.ResourceTags(),
	}
	i, err := ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, bastionMachineID, "instance"),
		&instanceArgs)
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputKeyUsername, pulumi.String(amiDefaultUsername))
	ctx.Export(r.OutputKeyHost, i.PublicIp)
	return i, nil
}

// Write exported values in context to files o a selected target folder
func WriteOutputs(stackResult auto.UpResult,
	prefix string,
	destinationFolder string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", prefix, OutputBastionUserPrivateKey): "bastion_id_rsa",
		fmt.Sprintf("%s-%s", prefix, OutputBastionUsername):       "bastion_username",
		fmt.Sprintf("%s-%s", prefix, OutputBastionHost):           "bastion_host",
	}
	return output.Write(stackResult, destinationFolder, results)
}
