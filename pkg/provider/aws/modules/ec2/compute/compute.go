package compute

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	diskSize            int    = 200
	rootBlockDeviceName string = "/dev/sda1"

	// Delay health check due to baremetal + userdata otherwise it will kill hosts consntantly
	// Probably move this to compute asset as each can have different value depending on userdata
	defaultHealthCheckGracePeriod int = 1200

	LoggingCmdStd   = true
	NoLoggingCmdStd = false
)

type ComputeRequest struct {
	MCtx   *mc.Context
	Prefix string
	ID     string
	VPC    *ec2.Vpc
	Subnet *ec2.Subnet
	LB     *lb.LoadBalancer
	LBEIP  *ec2.Eip
	// Array of TCP ports to be
	// created as tg for the LB
	LBTargetGroups  []int
	AMI             *ec2.LookupAmiResult
	KeyResources    *keypair.KeyPairResources
	SecurityGroups  pulumi.StringArray
	InstaceTypes    []string
	InstanceProfile *iam.InstanceProfile
	DiskSize        *int
	Airgap          bool
	Spot            bool
	// Only required if Spot is true
	SpotPrice float64
	// Only required if we need to set userdata
	UserDataAsBase64 pulumi.StringPtrInput
	// If we need to add explicit dependecies
	DependsOn []pulumi.Resource
}

type Compute struct {
	// Non spot return a on-demand insntace
	Instance *ec2.Instance
	// Spot instance is created through a mixedPolicy on a as group
	// in case of asg it is accessed through the LB
	AutoscalingGroup *autoscaling.Group
	LB               *lb.LoadBalancer
	LBEIP            *ec2.Eip
	// This can be used in case explicit
	// dependencies on the Compute resources
	Dependencies []pulumi.Resource
}

// Create compute resource based on requested args

// main logic differs based on Spot arg, if non spot it will create an on-demand
// otherwise it will create an asg with mixed policiy forcing only spot

// TODO on-demand could be changed to to policy trying spot
func (r *ComputeRequest) NewCompute(ctx *pulumi.Context) (*Compute, error) {
	if r.Spot {
		asg, err := r.spotInstance(ctx)
		return &Compute{
			AutoscalingGroup: asg,
			LB:               r.LB,
			LBEIP:            r.LBEIP,
			Dependencies:     []pulumi.Resource{asg, r.LB, r.LBEIP}}, err
	}
	i, err := r.onDemandInstance(ctx)
	return &Compute{Instance: i,
		Dependencies: []pulumi.Resource{i}}, err
}

// Create on demand instance
func (r *ComputeRequest) onDemandInstance(ctx *pulumi.Context) (*ec2.Instance, error) {
	args := ec2.InstanceArgs{
		SubnetId:                 r.Subnet.ID(),
		Ami:                      pulumi.String(r.AMI.Id),
		InstanceType:             pulumi.String(util.RandomItemFromArray(r.InstaceTypes)),
		KeyName:                  r.KeyResources.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(true),
		VpcSecurityGroupIds:      r.SecurityGroups,
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(diskSize),
		},
		Tags: r.MCtx.ResourceTags(),
	}
	if r.InstanceProfile != nil {
		args.IamInstanceProfile = r.InstanceProfile
	}
	if r.UserDataAsBase64 != nil {
		args.UserData = r.UserDataAsBase64
	}
	if r.Airgap {
		args.AssociatePublicIpAddress = pulumi.Bool(false)
	}
	return ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "instance"),
		&args,
		pulumi.DependsOn(r.DependsOn))
}

// create asg with 1 instance forced by spot
func (r ComputeRequest) spotInstance(ctx *pulumi.Context) (*autoscaling.Group, error) {
	// Logging information
	r.Subnet.AvailabilityZone.ApplyT(func(az string) error {
		logging.Debugf("Requesting a spot instance of types: %s at %s paying: %f",
			strings.Join(r.InstaceTypes, ", "), az, r.SpotPrice)
		return nil
	})
	args := &ec2.LaunchTemplateArgs{
		NamePrefix: pulumi.String(r.ID),
		ImageId:    pulumi.String(r.AMI.Id),
		KeyName:    r.KeyResources.AWSKeyPair.KeyName,
		EbsOptimized: pulumi.String("true"),
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
		Tags: r.MCtx.ResourceTags(),
		TagSpecifications: ec2.LaunchTemplateTagSpecificationArray{
			&ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String(constants.PulumiAwsResourceInstance),
				Tags:         r.MCtx.ResourceTags(),
			},
			&ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String(constants.PulumiAwsResourceVolume),
				Tags:         r.MCtx.ResourceTags(),
			},
			&ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String(constants.PulumiAwsResourceNetworkInterface),
				Tags:         r.MCtx.ResourceTags(),
			},
			&ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String(constants.PulumiAwsResourceSpotInstanceRequest),
				Tags:         r.MCtx.ResourceTags(),
			},
		},
	}
	if r.InstanceProfile != nil {
		args.IamInstanceProfile = ec2.LaunchTemplateIamInstanceProfileArgs{
			Arn: r.InstanceProfile.Arn,
		}
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
	spotMaxPrice := strconv.FormatFloat(r.SpotPrice, 'f', -1, 64)
	mixedInstancesPolicy := &autoscaling.GroupMixedInstancesPolicyArgs{
		InstancesDistribution: &autoscaling.GroupMixedInstancesPolicyInstancesDistributionArgs{
			OnDemandBaseCapacity:                pulumi.Int(0),
			OnDemandPercentageAboveBaseCapacity: pulumi.Int(0),
			SpotAllocationStrategy:              pulumi.String("capacity-optimized"),
			SpotMaxPrice:                        pulumi.String(spotMaxPrice),
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
			Delete: "30m"}),
		pulumi.DependsOn(r.DependsOn))
}

// function returns the ip to access the target host
func (c *Compute) GetHostIP(public bool) (ip pulumi.StringInput) {
	if c.Instance != nil {
		if public {
			return c.Instance.PublicDns
		}
		return c.Instance.PrivateDns
	}
	if c.LBEIP != nil {
		return c.LBEIP.PublicIp
	}
	return c.LB.DnsName
}

// Check if compute is healthy based on running a remote cmd
func (compute *Compute) Readiness(ctx *pulumi.Context,
	cmd string,
	prefix, id string,
	mk *tls.PrivateKey, username string,
	b *bastion.Bastion,
	dependecies []pulumi.Resource) error {
	_, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(prefix, id, "readiness-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(compute, mk, username, b),
			Create:     pulumi.String(cmd),
			Update:     pulumi.String(cmd),
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: command.RemoteTimeout,
				Update: command.RemoteTimeout}),
		pulumi.DependsOn(dependecies))
	return err
}

// Check if compute is healthy based on running a remote cmd
func (compute *Compute) RunCommand(ctx *pulumi.Context,
	cmd string,
	loggingCmdStd bool,
	prefix, id string,
	mk *tls.PrivateKey, username string,
	b *bastion.Bastion,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	ca := &remote.CommandArgs{
		Connection: remoteCommandArgs(compute, mk, username, b),
		Create:     pulumi.String(cmd),
		Update:     pulumi.String(cmd),
	}
	if !loggingCmdStd {
		ca.Logging = remote.LoggingNone
	}
	return remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(prefix, id, "cmd"),
		ca,
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: command.RemoteTimeout,
				Update: command.RemoteTimeout}),
		pulumi.DependsOn(dependecies))
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
