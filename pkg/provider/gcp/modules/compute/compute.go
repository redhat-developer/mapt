package compute

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	gcpCompute "github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/compute"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	defaultDiskSize int = 200
	defaultDiskType     = "pd-ssd"
	sshPort             = 22
)

type ComputeRequest struct {
	Prefix        string
	ID            string
	Zone          string
	Network       *gcpCompute.Network
	Subnet        *gcpCompute.Subnetwork
	PublicIP      *gcpCompute.Address
	MachineType   string
	ImageURI      string
	KeyPair       *tls.PrivateKey
	DiskSize      *int
	Spot          bool
	StartupScript string
	Username      string
	Tags          []string
}

type Compute struct {
	Instance *gcpCompute.Instance
	PublicIP *gcpCompute.Address
}

func (r *ComputeRequest) NewCompute(ctx *pulumi.Context) (*Compute, error) {
	instanceName := resourcesUtil.GetResourceName(r.Prefix, r.ID, "vm")

	diskSize := defaultDiskSize
	if r.DiskSize != nil {
		diskSize = *r.DiskSize
	}

	sshKeys := pulumi.Sprintf("%s:%s", r.Username, r.KeyPair.PublicKeyOpenssh)

	instanceArgs := &gcpCompute.InstanceArgs{
		MachineType: pulumi.String(r.MachineType),
		Zone:        pulumi.String(r.Zone),
		BootDisk: &gcpCompute.InstanceBootDiskArgs{
			InitializeParams: &gcpCompute.InstanceBootDiskInitializeParamsArgs{
				Image: pulumi.String(r.ImageURI),
				Size:  pulumi.Int(diskSize),
				Type:  pulumi.String(defaultDiskType),
			},
		},
		NetworkInterfaces: gcpCompute.InstanceNetworkInterfaceArray{
			&gcpCompute.InstanceNetworkInterfaceArgs{
				Network:    r.Network.SelfLink,
				Subnetwork: r.Subnet.SelfLink,
				AccessConfigs: gcpCompute.InstanceNetworkInterfaceAccessConfigArray{
					&gcpCompute.InstanceNetworkInterfaceAccessConfigArgs{
						NatIp:       r.PublicIP.Address,
						NetworkTier: pulumi.String("PREMIUM"),
					},
				},
			},
		},
		Metadata: pulumi.StringMap{
			"ssh-keys": sshKeys,
		},
		Tags: pulumi.ToStringArray(r.Tags),
	}

	if len(r.StartupScript) > 0 {
		instanceArgs.MetadataStartupScript = pulumi.String(r.StartupScript)
	}

	if r.Spot {
		instanceArgs.Scheduling = &gcpCompute.InstanceSchedulingArgs{
			ProvisioningModel:         pulumi.String("SPOT"),
			InstanceTerminationAction: pulumi.String("STOP"),
		}
	}

	instance, err := gcpCompute.NewInstance(ctx, instanceName, instanceArgs)
	if err != nil {
		return nil, err
	}

	return &Compute{
		Instance: instance,
		PublicIP: r.PublicIP,
	}, nil
}

func (c *Compute) GetHostIP() pulumi.StringOutput {
	return c.PublicIP.Address
}

func (c *Compute) Readiness(ctx *pulumi.Context, prefix, id, username string,
	privateKey *tls.PrivateKey) (*remote.Command, error) {
	cmdName := resourcesUtil.GetResourceName(prefix, id, "chk")
	return remote.NewCommand(ctx, cmdName, &remote.CommandArgs{
		Connection: remote.ConnectionArgs{
			Host:           c.GetHostIP(),
			PrivateKey:     privateKey.PrivateKeyOpenssh,
			User:           pulumi.String(username),
			Port:           pulumi.Float64(sshPort),
			DialErrorLimit: pulumi.Int(-1),
		},
		Create: pulumi.String(command.CommandCloudInitWait),
		Update: pulumi.String(command.CommandCloudInitWait),
	}, pulumi.Timeouts(
		&pulumi.CustomTimeouts{
			Create: command.RemoteTimeout,
			Update: command.RemoteTimeout,
		}),
		pulumi.DependsOn([]pulumi.Resource{c.Instance}))
}

func GetOutputHost(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, "host")
}

func GetOutputUsername(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, "username")
}

func GetOutputPrivateKey(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, "privatekey")
}
