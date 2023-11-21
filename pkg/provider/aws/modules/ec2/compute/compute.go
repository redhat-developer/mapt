package compute

import (
	"fmt"
	"strconv"

	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/bastion"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/keypair"
	"github.com/adrianriobo/qenvs/pkg/provider/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	diskSize            int    = 200
	rootBlockDeviceName string = "/dev/sda1"

	// Delay health check due to baremetal + userdata otherwise it will kill hosts consntantly
	// Probably move this to compute asset as each can have different value depending on userdata
	defaultHealthCheckGracePeriod int = 1200
)

type ComputeRequest struct {
	Prefix string
	ID     string
	VPC    *ec2.Vpc
	Subnet *ec2.Subnet
	LB     *lb.LoadBalancer
	// Array of TCP ports to be
	// created as tg for the LB
	LBTargetGroups []int
	AMI            *ec2.LookupAmiResult
	KeyResources   *keypair.KeyPairResources
	SecurityGroups pulumi.StringArray
	InstaceTypes   []string
	DiskSize       *int
	Airgap         bool
	Spot           bool
	// Only required if Spot is true
	SpotPrice        string
	UserDataAsBase64 pulumi.StringPtrInput
}

type Compute struct {
	// Non spot return a on-demand insntace
	Instance *ec2.Instance
	// Spot instance is created through a mixedPolicy on a as group
	// in case of asg it is accessed through the LB
	AutoscalingGroup *autoscaling.Group
	LB               *lb.LoadBalancer
}

// Create compute resource based on requested args

// main logic differs based on Spot arg, if non spot it will create an on-demand
// otherwise it will create an asg with mixed policiy forcing only spot

// TODO on-demand could be changed to to policy trying spot
func (r *ComputeRequest) NewCompute(ctx *pulumi.Context) (*Compute, error) {
	if r.Spot {
		asg, err := r.spotInstance(ctx)
		return &Compute{AutoscalingGroup: asg, LB: r.LB}, err
	}
	i, err := r.onDemandInstance(ctx)
	return &Compute{Instance: i}, err
}

// Create on demand instance
func (r *ComputeRequest) onDemandInstance(ctx *pulumi.Context) (*ec2.Instance, error) {
	args := ec2.InstanceArgs{
		SubnetId:                 r.Subnet.ID(),
		Ami:                      pulumi.String(r.AMI.Id),
		InstanceType:             pulumi.String(util.RandomItemFromArray[string](r.InstaceTypes)),
		KeyName:                  r.KeyResources.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(true),
		VpcSecurityGroupIds:      r.SecurityGroups,
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(diskSize),
		},
		Tags: qenvsContext.ResourceTags(),
	}
	if r.UserDataAsBase64 != nil {
		args.UserData = r.UserDataAsBase64
	}
	if r.Airgap {
		args.AssociatePublicIpAddress = pulumi.Bool(false)
	}
	return ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "instance"),
		&args)
}

// create asg with 1 instance forced by spot
func (r ComputeRequest) spotInstance(ctx *pulumi.Context) (*autoscaling.Group, error) {
	args := &ec2.LaunchTemplateArgs{
		NamePrefix: pulumi.String(r.ID),
		ImageId:    pulumi.String(r.AMI.Id),
		KeyName:    r.KeyResources.AWSKeyPair.KeyName,
		NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
			&ec2.LaunchTemplateNetworkInterfaceArgs{
				SecurityGroups:           r.SecurityGroups,
				AssociatePublicIpAddress: pulumi.String(strconv.FormatBool(!r.Airgap)),
				SubnetId:                 r.Subnet.ID(),
			},
		},
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String(rootBlockDeviceName),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeSize: pulumi.Int(diskSize),
				},
			},
		},
		Tags: qenvsContext.ResourceTags(),
	}
	if r.UserDataAsBase64 != nil {
		args.UserData = r.UserDataAsBase64
	}
	lt, err := ec2.NewLaunchTemplate(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "lt"),
		args)
	if err != nil {
		return nil, err
	}
	// Create target groups
	var tgGroupsARNs []pulumi.StringOutput
	for _, tgPort := range r.LBTargetGroups {
		tg, err := r.createForwardTargetGRoups(ctx, tgPort)
		if err != nil {
			return nil, err
		}
		tgGroupsARNs = append(tgGroupsARNs, tg.Arn)
	}
	overrides := autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArray{}
	for _, instanceType := range r.InstaceTypes {
		overrides = append(overrides, &autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
			InstanceType: pulumi.String(instanceType),
		})
	}
	mixedInstancesPolicy := &autoscaling.GroupMixedInstancesPolicyArgs{
		InstancesDistribution: &autoscaling.GroupMixedInstancesPolicyInstancesDistributionArgs{
			OnDemandBaseCapacity:                pulumi.Int(0),
			OnDemandPercentageAboveBaseCapacity: pulumi.Int(0),
			SpotAllocationStrategy:              pulumi.String("capacity-optimized"),
			SpotMaxPrice:                        pulumi.String(r.SpotPrice),
		},
		LaunchTemplate: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateArgs{
			LaunchTemplateSpecification: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationArgs{
				LaunchTemplateId: lt.ID(),
			},
			Overrides: overrides,
		},
	}
	return autoscaling.NewGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "asg"),
		&autoscaling.GroupArgs{
			TargetGroupArns:      pulumi.ToStringArrayOutput(tgGroupsARNs),
			CapacityRebalance:    pulumi.Bool(true),
			DesiredCapacity:      pulumi.Int(1),
			MaxSize:              pulumi.Int(1),
			MinSize:              pulumi.Int(1),
			VpcZoneIdentifiers:   pulumi.StringArray{r.Subnet.ID()},
			MixedInstancesPolicy: mixedInstancesPolicy,
			// Check if this is needed now
			HealthCheckGracePeriod: pulumi.Int(defaultHealthCheckGracePeriod),
			// Suspend healthcheck to allow restart computer
			// required on windows hosts for Openshift local installation
			SuspendedProcesses: pulumi.StringArray{
				pulumi.String("HealthCheck")},
			Tags: autoscaling.GroupTagArray{
				&autoscaling.GroupTagArgs{
					Key:               pulumi.String("Name"),
					Value:             pulumi.String(resourcesUtil.GetResourceName(r.Prefix, r.ID, "asg")),
					PropagateAtLaunch: pulumi.Bool(true),
				},
			},
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{
			Delete: "30m"}))
}

// function returns the ip to access the target host
func (c *Compute) GetHostIP(public bool) (ip pulumi.StringInput) {
	if c.Instance != nil {
		if public {
			return c.Instance.PublicDns
		}
		return c.Instance.PrivateDns
	}
	return c.LB.DnsName
}

// Check if compute is healthy (ping on ssh)
func (compute *Compute) Readiness(ctx *pulumi.Context,
	prefix, id string,
	mk *tls.PrivateKey, username string,
	b *bastion.Bastion,
	dependecies []pulumi.Resource) error {
	_, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(prefix, id, "readiness-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(compute, mk, username, b),
			Create:     pulumi.String(command.CommandPing),
			Update:     pulumi.String(command.CommandPing),
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: command.RemoteTimeout,
				Update: command.RemoteTimeout}),
		pulumi.DependsOn(dependecies))
	return err
}

// helper function to set the connection args
// based on bastion or direct connection to target host
func remoteCommandArgs(
	c *Compute,
	mk *tls.PrivateKey, username string,
	b *bastion.Bastion) remote.ConnectionArgs {
	ca := remote.ConnectionArgs{
		Host:           c.GetHostIP(b == nil),
		PrivateKey:     mk.PrivateKeyOpenssh,
		User:           pulumi.String(username),
		Port:           pulumi.Float64(22),
		DialErrorLimit: pulumi.Int(-1)}
	if b != nil {
		ca.Proxy = remote.ProxyConnectionArgs{
			Host:           b.Instance.PublicIp,
			PrivateKey:     b.PrivateKey.PrivateKeyOpenssh,
			User:           pulumi.String(b.Usarname),
			Port:           pulumi.Float64(b.Port),
			DialErrorLimit: pulumi.Int(-1)}

	}
	return ca
}

func (r ComputeRequest) createForwardTargetGRoups(ctx *pulumi.Context, port int) (*lb.TargetGroup, error) {
	tg, err := lb.NewTargetGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, fmt.Sprintf("tg-%d", port)),
		&lb.TargetGroupArgs{
			Port:     pulumi.Int(port),
			Protocol: pulumi.String("TCP"),
			VpcId:    r.VPC.ID(),
		})
	if err != nil {
		return nil, err
	}
	if _, err := lb.NewListener(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, fmt.Sprintf("listener-%d", port)),
		&lb.ListenerArgs{
			LoadBalancerArn: r.LB.Arn,
			Port:            pulumi.Int(port),
			Protocol:        pulumi.String("TCP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: tg.Arn,
				},
			},
		}); err != nil {
		return nil, err
	}
	return tg, nil
}
