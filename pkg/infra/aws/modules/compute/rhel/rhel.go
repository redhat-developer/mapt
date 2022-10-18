package rhel

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r RHELRequest) Create(ctx *pulumi.Context) (*RHELResources, error) {
	awsKeyPair, privateKey, err := compute.ManageKeypair(ctx, r.keyPair, r.Name, OutputPrivateKey)
	if err != nil {
		return nil, err
	}
	rhelIngressRule := securityGroup.SSH_TCP
	if r.Public {
		rhelIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		rhelIngressRule.SG = r.BastionSG
	}

	sg, err := securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  "rhel sg group",
		IngressRules: []securityGroup.IngressRules{rhelIngressRule}}.Create(ctx)
	if err != nil {
		return nil, err
	}

	amiNameRegex := fmt.Sprintf(defaultAMIPattern, r.VersionMajor)
	ami, err := ami.GetAMIByName(ctx, amiNameRegex)
	if err != nil {
		return nil, err
	}
	var sir *ec2.SpotInstanceRequest
	var i *ec2.Instance
	if len(r.SpotPrice) > 0 {
		sir, err = ec2.NewSpotInstanceRequest(ctx,
			r.Name,
			&ec2.SpotInstanceRequestArgs{
				SubnetId:            r.Subnets[0].ID(),
				Ami:                 pulumi.String(ami.Id),
				InstanceType:        pulumi.String(defaultInstanceType),
				KeyName:             awsKeyPair.KeyName,
				VpcSecurityGroupIds: pulumi.StringArray{sg.SG.ID()},
				SpotPrice:           pulumi.String(r.SpotPrice),
				// BlockDurationMinutes: pulumi.Int(defaultBlockDurationMinutes),
				WaitForFulfillment: pulumi.Bool(true),
				Tags: pulumi.StringMap{
					"Name": pulumi.String(r.Name),
				},
			})
		if err != nil {
			return nil, err
		}
		ctx.Export(OutputPrivateIP,
			util.If(r.Public,
				sir.PublicIp,
				sir.PrivateIp))
	} else {
		i, err = ec2.NewInstance(ctx,
			r.Name,
			&ec2.InstanceArgs{
				SubnetId:            r.Subnets[0].ID(),
				Ami:                 pulumi.String(ami.Id),
				InstanceType:        pulumi.String(defaultInstanceType),
				KeyName:             awsKeyPair.KeyName,
				VpcSecurityGroupIds: pulumi.StringArray{sg.SG.ID()},
				Tags: pulumi.StringMap{
					"Name": pulumi.String(r.Name),
				},
			})
		if err != nil {
			return nil, err
		}
		ctx.Export(OutputPrivateIP,
			util.If(r.Public,
				i.PublicIp,
				i.PrivateIp))
	}
	ctx.Export(OutputUsername, pulumi.String(defaultAMIUser))
	rhel := RHELResources{
		AWSKeyPair:          awsKeyPair,
		PrivateKey:          privateKey,
		Instance:            i,
		SpotInstanceRequest: sir,
	}
	if r.Public {
		return &rhel, rhel.waitForInit(ctx)
	}
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &rhel, nil
}

func (c RHELResources) waitForInit(ctx *pulumi.Context) error {
	instance := command.RemoteInstance{
		Instace:             c.Instance,
		SpotInstanceRequest: c.SpotInstanceRequest,
		Username:            defaultAMIUser,
		PrivateKey:          c.PrivateKey}
	return instance.RemoteExec(
		ctx,
		command.CommandPing,
		"rhel-WaitForConnect")
}
