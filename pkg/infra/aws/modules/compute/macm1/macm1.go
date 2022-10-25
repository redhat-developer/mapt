package macm1

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r MacM1Request) Create(ctx *pulumi.Context) (*MacM1Resources, error) {
	var macM1 MacM1Resources
	awsKeyPair, privateKey, err := compute.ManageKeypair(
		ctx, r.KeyPair, r.Name, OutputPrivateKey)
	if err != nil {
		return nil, err
	}
	macM1.AWSKeyPair = awsKeyPair
	macM1.PrivateKey = privateKey
	ingressRule := securityGroup.SSH_TCP
	if r.Public {
		ingressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		ingressRule.SG = r.BastionSG
	}
	sg, err := securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  "mac m1 sg group",
		IngressRules: []securityGroup.IngressRules{ingressRule}}.Create(ctx)
	if err != nil {
		return nil, err
	}

	ami, err := ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Filters)
	if err != nil {
		return nil, err
	}

	err = r.onDemandInstance(ctx, ami.Id, awsKeyPair, sg.SG, &macM1)
	if err != nil {
		return nil, err
	}

	ctx.Export(OutputUsername, pulumi.String(r.Specs.AMI.DefaultUser))
	if r.Public {
		return &macM1, macM1.waitForInit(ctx)
	}
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &macM1, nil
}

func (c MacM1Resources) waitForInit(ctx *pulumi.Context) error {
	instance := command.RemoteInstance{
		Instance:   c.Instance,
		InstanceIP: &c.InstanceIP,
		Username:   c.Username,
		PrivateKey: c.PrivateKey}
	return instance.RemoteExec(
		ctx,
		command.CommandPing,
		"macm1-WaitForConnect")
}

func (r MacM1Request) onDemandInstance(ctx *pulumi.Context,
	amiID string, keyPair *ec2.KeyPair, sg *ec2.SecurityGroup,
	rhel *MacM1Resources) error {
	i, err := ec2.NewInstance(ctx,
		r.Name,
		&ec2.InstanceArgs{
			SubnetId:                 r.Subnets[0].ID(),
			Ami:                      pulumi.String(amiID),
			InstanceType:             pulumi.String(r.Specs.InstaceTypes[0]),
			KeyName:                  keyPair.KeyName,
			AssociatePublicIpAddress: pulumi.Bool(r.Public),
			VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String(r.Name),
				"HOST_ID": pulumi.String(r.Specs.ID),
			},
		})
	if err != nil {
		return err
	}
	rhel.Instance = i
	rhel.Username = r.Specs.AMI.DefaultUser
	ctx.Export(OutputPrivateIP,
		util.If(r.Public,
			i.PublicIp,
			i.PrivateIp))
	return nil
}
