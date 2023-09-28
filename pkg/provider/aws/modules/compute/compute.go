package compute

import (
	"github.com/adrianriobo/qenvs/pkg/provider/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (c *Compute) getSecurityGroupsIDs() pulumi.StringArrayInput {
	sgs := util.ArrayConvert(c.SG,
		func(sg *ec2.SecurityGroup) pulumi.StringInput {
			return sg.ID()
		})
	return pulumi.StringArray(sgs[:])
}

func (c *Compute) remoteExec(ctx *pulumi.Context, cmd pulumi.StringPtrInput, cmdName string,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	instance := command.RemoteInstance{
		Instance:   c.Instance,
		InstanceIP: &c.InstanceIP,
		Username:   c.Username,
		PrivateKey: c.PrivateKey}
	return instance.RemoteExec(
		ctx,
		cmd,
		cmdName,
		dependecies)
}
