package command

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// remoteCommandTimeout int = 300
	// https://www.pulumi.com/docs/intro/concepts/resources/options/customtimeouts/
	RemoteTimeout string = "40m"

	// https://bugs.launchpad.net/ubuntu/+source/cloud-init/+bug/2048522
	CommandCloudInitWait string = "sudo cloud-init status --long --wait || [[ $? -eq 2 || $? -eq 0 ]]"
	CommandPing          string = "echo ping"

	defaultSSHPort int = 22
)

type RemoteInstance struct {
	InstanceIP *pulumi.StringOutput
	Instance   *ec2.Instance
	Username   string
	PrivateKey *tls.PrivateKey
}

// Remote command success if error = nil
type RemoteCommand func(ctx *pulumi.Context, remoteCommand, remoteCommandName string) error

func (r RemoteInstance) RemoteExec(ctx *pulumi.Context, remoteCommand pulumi.StringPtrInput, remoteCommandName string,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	remoteIP, err := r.getRemoteHost()
	if err != nil {
		return nil, err
	}
	return remote.NewCommand(ctx, remoteCommandName, &remote.CommandArgs{
		Connection: remote.ConnectionArgs{
			Host:           remoteIP,
			PrivateKey:     r.PrivateKey.PrivateKeyOpenssh,
			User:           pulumi.String(r.Username),
			Port:           pulumi.Float64(defaultSSHPort),
			DialErrorLimit: pulumi.Int(-1),
		},
		Create: remoteCommand,
		Update: remoteCommand,
	}, pulumi.Timeouts(
		&pulumi.CustomTimeouts{
			Create: RemoteTimeout,
			Update: RemoteTimeout}),
		pulumi.DependsOn(dependecies))
}

func (r RemoteInstance) getRemoteHost() (pulumi.StringOutput, error) {
	if r.Instance != nil {
		return r.Instance.PublicIp, nil
	}
	if r.InstanceIP != nil {
		return *r.InstanceIP, nil
	}
	// if len(r.InstanceIP) > 0 {
	// 	return pulumi.String(r.InstanceIP).ToStringOutput(), nil

	return pulumi.StringOutput{}, fmt.Errorf("a valid instance or spot request is required to exec a remote command")
}

// func (r RemoteInstance) RemoteCommandAwait(c RemoteCommand, ctx *pulumi.Context, remoteCommand string) error {
// 	// this is to avoid garbage leak using direclty ticker
// 	ticker := time.NewTicker(500 * time.Millisecond)
// 	defer func() { ticker.Stop() }()
// 	timer := time.After(time.Duration(remoteCommandTimeout) * time.Second)
// 	for {
// 		select {
// 		case <-ticker.C:
// 			if err := c(ctx, remoteCommand); err == nil {
// 				return nil
// 			}
// 		case <-timer:
// 			return fmt.Errorf("command %s failed", remoteCommand)
// 		}
// 	}
// }
