package macm1

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r HostRequest) Create(ctx *pulumi.Context) (*HostResources, error) {
	var host = HostResources{
		Name: r.Name,
	}
	awsKeyPair, privateKey, err := compute.ManageKeypair(
		ctx, r.KeyPair, r.Name,
		fmt.Sprintf("%s%s", r.Name, OutputPrivateIP))
	if err != nil {
		return nil, err
	}
	host.AWSKeyPair = awsKeyPair
	host.PrivateKey = privateKey
	ingressRule := securityGroup.SSH_TCP
	if r.Public {
		ingressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		ingressRule.SG = r.BastionSG
	}
	_, err = securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  fmt.Sprintf("%s sg group", r.Name),
		IngressRules: []securityGroup.IngressRules{ingressRule}}.Create(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// func (c HostResources) remoteExec(ctx *pulumi.Context, cmdName, cmd string) error {
// 	instance := command.RemoteInstance{
// 		Instance:   c.Instance,
// 		InstanceIP: &c.InstanceIP,
// 		Username:   c.Username,
// 		PrivateKey: c.PrivateKey}
// 	return instance.RemoteExec(
// 		ctx,
// 		cmd,
// 		cmdName)
// }
