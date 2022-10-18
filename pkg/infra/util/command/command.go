package command

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// remoteCommandTimeout int = 300
	// https://www.pulumi.com/docs/intro/concepts/resources/options/customtimeouts/
	remoteTimeout string = "10m"

	CommandCloudInitWait string = "sudo cloud-init status --wait"
	CommandPing          string = "echo ping"

	defaultSSHPort int = 22
)

type RemoteInstance struct {
	SpotInstanceRequest *ec2.SpotInstanceRequest
	Instace             *ec2.Instance
	Username            string
	PrivateKey          *tls.PrivateKey
}

// Remote command success if error = nil
type RemoteCommand func(ctx *pulumi.Context, remoteCommand, remoteCommandName string) error

func (r RemoteInstance) RemoteExec(ctx *pulumi.Context, remoteCommand, remoteCommandName string) error {
	remoteIP, err := r.getRemoteHost()
	if err != nil {
		return err
	}
	_, err = remote.NewCommand(ctx, remoteCommandName, &remote.CommandArgs{
		Connection: remote.ConnectionArgs{
			Host:       remoteIP,
			PrivateKey: r.PrivateKey.PrivateKeyOpenssh,
			User:       pulumi.String(r.Username),
			Port:       pulumi.Float64(defaultSSHPort),
		},
		Create: pulumi.String(remoteCommand),
		Update: pulumi.String(remoteCommand),
	}, pulumi.Timeouts(
		&pulumi.CustomTimeouts{
			Create: remoteTimeout,
			Update: remoteTimeout}))
	if err != nil {
		return err
	}
	return nil
}

func (r RemoteInstance) getRemoteHost() (pulumi.StringOutput, error) {
	if r.Instace != nil {
		return r.Instace.PublicIp, nil
	}
	if r.SpotInstanceRequest != nil {
		return r.SpotInstanceRequest.PublicIp, nil
	}
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
