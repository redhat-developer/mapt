package bastion

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
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
	outputBastionUserPrivateKey = "bastion_id_rsa"
	outputBastionUsername       = "bastion_username"
	outputBastionHost           = "bastion_host"
)

type BastionArgs struct {
	Prefix string
	VPC    *ec2.Vpc
	Subnet *ec2.Subnet
}

type BastionResult struct {
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
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *BastionArgs) (*BastionResult, error) {
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			args.Prefix, bastionMachineID, "pk")}
	keyResources, err := kpr.Create(ctx, mCtx)
	if err != nil {
		return nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", args.Prefix, outputBastionUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	sgs, err := securityGroups(ctx, mCtx, args.Prefix, args.VPC)
	if err != nil {
		return nil, err
	}
	i, err := instance(ctx, mCtx, &instaceArgs{
		prefix:         args.Prefix,
		subnet:         args.Subnet,
		keyResources:   keyResources,
		securityGroups: sgs,
	})
	if err != nil {
		return nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", args.Prefix, outputBastionUsername),
		pulumi.String(amiDefaultUsername))
	ctx.Export(fmt.Sprintf("%s-%s", args.Prefix, outputBastionHost), i.PublicIp)
	return &BastionResult{
		Instance:   i,
		PrivateKey: keyResources.PrivateKey,
		Usarname:   amiDefaultUsername,
		Port:       defaultSSHPort,
	}, nil
}

// Allow connect bastion on ssh port
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix string, vpc *ec2.Vpc) (pulumi.StringArray, error) {
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(prefix, bastionMachineID, "sg"),
		VPC:          vpc,
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

type instaceArgs struct {
	prefix         string
	subnet         *ec2.Subnet
	keyResources   *keypair.KeyPairResources
	securityGroups pulumi.StringArray
}

func instance(ctx *pulumi.Context, mCtx *mc.Context, args *instaceArgs) (*ec2.Instance, error) {
	ami, err := ami.GetAMIByName(ctx, amiRegex, nil, nil, mCtx.TargetHostingPlace())
	if err != nil {
		return nil, err
	}
	instanceArgs := ec2.InstanceArgs{
		SubnetId:         args.subnet.ID(),
		ImageId:          pulumi.String(ami.ImageId),
		InstanceType:     pulumi.String(instanceType),
		KeyName:          args.keyResources.AWSKeyPair.KeyName,
		SecurityGroupIds: args.securityGroups,
		BlockDeviceMappings: ec2.InstanceBlockDeviceMappingArray{
			&ec2.InstanceBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/sda1"),
				Ebs: &ec2.InstanceEbsArgs{
					VolumeSize: pulumi.Int(diskSize),
				},
			},
		},
		// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
	}
	i, err := ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(args.prefix, bastionMachineID, "instance"),
		&instanceArgs)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// Write exported values in context to files o a selected target folder
func WriteOutputs(stackResult auto.UpResult,
	prefix string,
	destinationFolder string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputBastionUserPrivateKey): "bastion_id_rsa",
		fmt.Sprintf("%s-%s", prefix, outputBastionUsername):       "bastion_username",
		fmt.Sprintf("%s-%s", prefix, outputBastionHost):           "bastion_host",
	}
	return output.Write(stackResult, destinationFolder, results)
}
