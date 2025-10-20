package compute

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	lb "github.com/pulumi/pulumi-aws-native/sdk/go/aws/elasticloadbalancingv2"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
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
	Eip    *ec2.Eip
	// If LB is nill EIP should be associated to the machine
	// to allow make use of it before creating the instance
	LB *lb.LoadBalancer
	// Array of TCP ports to be
	// created as tg for the LB
	LBTargetGroups  []int
	AMI             *ami.AMIResult
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
	AutoscalingGroup *autoscaling.AutoScalingGroup
	// If LB is nil Eip is used for machine otherwise for the instance
	Eip *ec2.Eip
	LB  *lb.LoadBalancer
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
			Eip:              r.Eip,
			Dependencies:     []pulumi.Resource{asg, r.LB, r.Eip}}, err
	}
	i, err := r.onDemandInstance(ctx)
	return &Compute{
		Instance:     i,
		Eip:          r.Eip,
		Dependencies: []pulumi.Resource{i, r.Eip}}, err
}

// Create on demand instance
func (r *ComputeRequest) onDemandInstance(ctx *pulumi.Context) (*ec2.Instance, error) {
	args := ec2.InstanceArgs{
		SubnetId:         r.Subnet.ID(),
		ImageId:          pulumi.String(r.AMI.ImageId),
		InstanceType:     pulumi.String(util.RandomItemFromArray(r.InstaceTypes)),
		KeyName:          r.KeyResources.AWSKeyPair.KeyName,
		SecurityGroupIds: r.SecurityGroups,
		BlockDeviceMappings: ec2.InstanceBlockDeviceMappingArray{
			&ec2.InstanceBlockDeviceMappingArgs{
				DeviceName: pulumi.String(rootBlockDeviceName),
				Ebs: &ec2.InstanceEbsArgs{
					VolumeSize: pulumi.Int(diskSize),
				},
			},
		},
		// Tags: r.MCtx.ResourceTags(), // TODO: Convert to AWS Native tag format
	}
	if r.InstanceProfile != nil {
		args.IamInstanceProfile = r.InstanceProfile.Arn
	}
	if r.UserDataAsBase64 != nil {
		args.UserData = r.UserDataAsBase64
	}
	// Note: AWS Native doesn't support AssociatePublicIpAddress directly
	// Public IP association is typically handled through subnet configuration
	instance, err := ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "instance"),
		&args,
		pulumi.DependsOn(r.DependsOn))
	if err != nil {
		return nil, err
	}
	_, err = ec2.NewEipAssociation(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "instance-eip"),
		&ec2.EipAssociationArgs{
			InstanceId:   instance.ID(),
			AllocationId: r.Eip.AllocationId,
		})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// create asg with 1 instance forced by spot
func (r ComputeRequest) spotInstance(ctx *pulumi.Context) (*autoscaling.AutoScalingGroup, error) {
	// Logging information
	r.Subnet.AvailabilityZone.ApplyT(func(az *string) string {
		azVal := ""
		if az != nil {
			azVal = *az
		}
		logging.Debugf("Requesting a spot instance of types: %s at %s paying: %f",
			strings.Join(r.InstaceTypes, ", "), azVal, r.SpotPrice)
		return azVal
	})
	// Create launch template data structure for aws-native
	launchTemplateData := &ec2.LaunchTemplateDataArgs{
		ImageId:      pulumi.String(r.AMI.ImageId),
		KeyName:      r.KeyResources.AWSKeyPair.KeyName,
		EbsOptimized: pulumi.Bool(true),
		NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArrayInput(ec2.LaunchTemplateNetworkInterfaceArray{
			&ec2.LaunchTemplateNetworkInterfaceArgs{
				DeviceIndex:              pulumi.Int(0),
				Groups:                   r.SecurityGroups,
				SubnetId:                 r.Subnet.ID(),
				AssociatePublicIpAddress: pulumi.Bool(true),
				// Note: AssociatePublicIpAddress not supported in AWS Native LaunchTemplate
				// Public IP association is handled through subnet configuration
			},
		}),
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArrayInput(ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String(rootBlockDeviceName),
				Ebs: &ec2.LaunchTemplateEbsArgs{
					VolumeSize: pulumi.Int(diskSize),
				},
			},
		}),
		// TagSpecifications: ec2.TagSpecificationArrayInput(ec2.TagSpecificationArray{
		//	&ec2.TagSpecificationArgs{
		//		ResourceType: pulumi.String(constants.PulumiAwsResourceInstance),
		//		Tags:         r.MCtx.ResourceTags(),
		//	},
		//	&ec2.TagSpecificationArgs{
		//		ResourceType: pulumi.String(constants.PulumiAwsResourceVolume),
		//		Tags:         r.MCtx.ResourceTags(),
		//	},
		//	&ec2.TagSpecificationArgs{
		//		ResourceType: pulumi.String(constants.PulumiAwsResourceNetworkInterface),
		//		Tags:         r.MCtx.ResourceTags(),
		//	},
		//	&ec2.TagSpecificationArgs{
		//		ResourceType: pulumi.String(constants.PulumiAwsResourceSpotInstanceRequest),
		//		Tags:         r.MCtx.ResourceTags(),
		//	},
		// }), // TODO: Fix tags format for AWS Native
	}
	if r.InstanceProfile != nil {
		launchTemplateData.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileArgs{
			Arn: r.InstanceProfile.Arn,
		}
	}
	if r.UserDataAsBase64 != nil {
		launchTemplateData.UserData = r.UserDataAsBase64
	}
	args := &ec2.LaunchTemplateArgs{
		LaunchTemplateName: pulumi.String(r.ID),
		LaunchTemplateData: launchTemplateData,
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
		tgGroupsARNs = append(tgGroupsARNs, tg.TargetGroupArn)
	}
	overrides := autoscaling.AutoScalingGroupLaunchTemplateOverridesArray{}
	for _, instanceType := range r.InstaceTypes {
		overrides = append(overrides, &autoscaling.AutoScalingGroupLaunchTemplateOverridesArgs{
			InstanceType: pulumi.String(instanceType),
		})
	}
	spotMaxPrice := strconv.FormatFloat(r.SpotPrice, 'f', -1, 64)
	mixedInstancesPolicy := &autoscaling.AutoScalingGroupMixedInstancesPolicyArgs{
		InstancesDistribution: &autoscaling.AutoScalingGroupInstancesDistributionArgs{
			OnDemandBaseCapacity:                pulumi.Int(0),
			OnDemandPercentageAboveBaseCapacity: pulumi.Int(0),
			SpotAllocationStrategy:              pulumi.String("capacity-optimized"),
			SpotMaxPrice:                        pulumi.String(spotMaxPrice),
		},
		LaunchTemplate: &autoscaling.AutoScalingGroupLaunchTemplateArgs{
			LaunchTemplateSpecification: &autoscaling.AutoScalingGroupLaunchTemplateSpecificationArgs{
				LaunchTemplateId: lt.ID(),
				Version:          lt.LatestVersionNumber,
			},
			Overrides: overrides,
		},
	}
	return autoscaling.NewAutoScalingGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ID, "asg"),
		&autoscaling.AutoScalingGroupArgs{
			TargetGroupArns:      pulumi.ToStringArrayOutput(tgGroupsARNs),
			CapacityRebalance:    pulumi.Bool(true),
			DesiredCapacity:      pulumi.String("1"),
			MaxSize:              pulumi.String("1"),
			MinSize:              pulumi.String("1"),
			VpcZoneIdentifier:    pulumi.StringArray{r.Subnet.ID()},
			MixedInstancesPolicy: mixedInstancesPolicy,
			// Check if this is needed now
			HealthCheckGracePeriod: pulumi.Int(defaultHealthCheckGracePeriod),
			// Note: SuspendedProcesses not available in aws-native
			// This is an operational property not supported by CloudFormation
			// HealthCheck suspension would need to be managed separately via AWS API
			Tags: autoscaling.AutoScalingGroupTagPropertyArray{
				&autoscaling.AutoScalingGroupTagPropertyArgs{
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
	if c.LB != nil {
		return c.LB.DnsName
	}
	// Note: aws-native EIP doesn't have PublicDns/PrivateDns fields
	// Using PublicIp for both cases as DNS resolution is not directly available
	return c.Eip.PublicIp
}

// Check if compute is healthy based on running a remote cmd
func (compute *Compute) Readiness(ctx *pulumi.Context,
	cmd string,
	prefix, id string,
	mk *tls.PrivateKey, username string,
	b *bastion.BastionResult,
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
	b *bastion.BastionResult,
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
	b *bastion.BastionResult) remote.ConnectionArgs {
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
			LoadBalancerArn: r.LB.LoadBalancerArn,
			Port:            pulumi.Int(port),
			Protocol:        pulumi.String("TCP"),
			DefaultActions: lb.ListenerActionArray{
				&lb.ListenerActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: tg.TargetGroupArn,
				},
			},
		}); err != nil {
		return nil, err
	}
	return tg, nil
}
