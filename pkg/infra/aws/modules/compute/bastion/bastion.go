package bastion

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r BastionRequest) Create(ctx *pulumi.Context) (*BastionResources, error) {
	awsKeyPair, privateKey, err := compute.ManageKeypair(ctx, r.keyPair, r.Name, OutputPrivateKey)
	if err != nil {
		return nil, err
	}
	bastionIngressRule := securityGroup.SSH_TCP
	bastionIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	sg, err := securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  "bastion sg group",
		IngressRules: []securityGroup.IngressRules{bastionIngressRule}}.Create(ctx)
	if err != nil {
		return nil, err
	}

	var instance *ec2.Instance
	if !r.HA {
		ami, err := ami.GetAMIByName(ctx, bastionDefaultAMI, nil)
		if err != nil {
			return nil, err
		}
		instance, err = ec2.NewInstance(ctx,
			r.Name,
			&ec2.InstanceArgs{
				Tags: pulumi.StringMap{
					"Name": pulumi.String(r.Name),
				},
				SubnetId:                 r.PublicSubnets[0].ID(),
				Ami:                      pulumi.String(ami.Id),
				InstanceType:             pulumi.String(bastionDefaultInstanceType),
				KeyName:                  awsKeyPair.KeyName,
				VpcSecurityGroupIds:      pulumi.StringArray{sg.SG.ID()},
				AssociatePublicIpAddress: pulumi.Bool(true),
			})
		if err != nil {
			return nil, err
		}
		ctx.Export(OutputPublicIP, instance.PublicIp)
		ctx.Export(OutputUsername, pulumi.String(bastionDefaultAMIUser))
	}
	bastion := BastionResources{
		AWSKeyPair: awsKeyPair,
		PrivateKey: privateKey,
		Instance:   instance,
		SG:         sg.SG,
	}
	// return &bastion, bastion.toRemoteIsn waitForInit(ctx)
	return &bastion, bastion.waitForInit(ctx)
	// return &bastion, nil
}

func (c BastionResources) waitForInit(ctx *pulumi.Context) error {
	return command.RemoteInstance{
		Instance:   c.Instance,
		Username:   bastionDefaultAMIUser,
		PrivateKey: c.PrivateKey}.RemoteExec(ctx, command.CommandPing, "bastion-WaitForConnect")
}
