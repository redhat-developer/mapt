package compute

import (
	utilRemote "github.com/adrianriobo/qenvs/pkg/infra/util/remote"
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

func (c *Compute) RemoteExec(ctx *pulumi.Context,
	cmd pulumi.StringPtrInput, cmdName string,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	instance := utilRemote.RemoteInstance{
		Instance:   c.Instance,
		InstanceIP: &c.InstanceIP,
		Username:   c.Username,
		PrivateKey: c.PrivateKey.PrivateKeyOpenssh}
	return instance.RemoteExec(
		ctx,
		cmd,
		cmdName,
		dependecies)
}

func ExecOnRemoteInstance(ctx *pulumi.Context,
	cmd pulumi.StringPtrInput, cmdName string,
	remoteInstance utilRemote.RemoteInstance,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	return remoteInstance.RemoteExec(
		ctx,
		cmd,
		cmdName,
		dependecies)
}

func (c *Compute) RemoteCopy(ctx *pulumi.Context,
	localPath, remotePath string, name string,
	dependecies []pulumi.Resource) (*remote.CopyFile, error) {
	instance := utilRemote.RemoteInstance{
		Instance:   c.Instance,
		InstanceIP: &c.InstanceIP,
		Username:   c.Username,
		PrivateKey: c.PrivateKey.PrivateKeyOpenssh}
	return instance.CopyFile(
		ctx,
		localPath,
		remotePath,
		name,
		dependecies)
}

func CopyOnRemoteInstance(ctx *pulumi.Context,
	localPath, remotePath string, name string,
	remoteInstance utilRemote.RemoteInstance,
	dependecies []pulumi.Resource) (*remote.CopyFile, error) {
	return remoteInstance.CopyFile(
		ctx,
		localPath,
		remotePath,
		name,
		dependecies)
}
