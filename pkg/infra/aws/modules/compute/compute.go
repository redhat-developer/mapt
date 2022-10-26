package compute

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/keypair"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *Request) GetName() string {
	return fmt.Sprintf("%s-%s", r.ProjecName, r.Specs.ID)
}

func (r *Request) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, "", r.Specs.AMI.Filters)
}

func (r *Request) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *Request) GetPostScript() (string, error) {
	return "", nil
}

func (r *Request) Create(ctx *pulumi.Context, computeType ComputeRequestType) (*Resources, error) {
	// Manage keypairs for requested host
	compute := Resources{
		Name:  r.GetName(),
		Specs: r.Specs,
	}
	if err := r.manageKeypair(ctx, &compute); err != nil {
		return nil, err
	}
	// Create sg according to request params
	if err := r.manageSecurityGroup(ctx, &compute); err != nil {
		return nil, err
	}
	ami, err := computeType.GetAMI(ctx)
	if err != nil {
		return nil, err
	}

	if len(r.SpotPrice) > 0 {
		err = r.createSpotInstance(ctx, ami.Id, &compute)
		if err != nil {
			return nil, err
		}
	} else {
		dh, err := computeType.GetDedicatedHost(ctx)
		if err != nil {
			return nil, err
		}
		err = r.createOnDemand(ctx, ami.Id, dh, &compute)
		if err != nil {
			return nil, err
		}
	}
	ctx.Export(compute.OutputUsername(), pulumi.String(r.Specs.AMI.DefaultUser))
	if r.Public {
		postScript, err := computeType.GetPostScript()
		if err != nil {
			return nil, err
		}
		waitCmddependencies := []pulumi.Resource{}
		if len(postScript) > 0 {
			rc, err := compute.remoteExec(ctx,
				fmt.Sprintf("%s-%s", r.Specs.ID, "postscript"), postScript, nil)
			if err != nil {
				return nil, err
			}
			waitCmddependencies = append(waitCmddependencies, rc)
		}
		_, err = compute.remoteExec(ctx,
			fmt.Sprintf("%s-%s", r.Specs.ID, "wait"), command.CommandPing, waitCmddependencies)
		return &compute, err
	}
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &compute, nil
}

func (r *Resources) OutputPrivateKey() string {
	return fmt.Sprintf("%s-%s", OutputPrivateKey, r.Specs.ID)
}

func (r *Resources) OutputHost() string {
	return fmt.Sprintf("%s-%s", OutputHost, r.Specs.ID)
}

func (r *Resources) OutputUsername() string {
	return fmt.Sprintf("%s-%s", OutputUsername, r.Specs.ID)
}

func (r *Request) manageKeypair(ctx *pulumi.Context, result *Resources) error {
	if r.KeyPair == nil {
		// create key
		keyResources, err := keypair.KeyPairRequest{
			Name: r.GetName()}.Create(ctx)
		if err != nil {
			return err
		}
		result.AWSKeyPair = keyResources.AWSKeyPair
		result.PrivateKey = keyResources.PrivateKey
		ctx.Export(result.OutputPrivateKey(), keyResources.PrivateKey.PrivateKeyPem)
		return nil
	}
	result.AWSKeyPair = r.KeyPair

	return nil
}

func (r *Request) manageSecurityGroup(ctx *pulumi.Context, result *Resources) error {
	ingressRule := securityGroup.SSH_TCP
	if r.Public {
		ingressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		ingressRule.SG = r.BastionSG
	}
	sg, err := securityGroup.SGRequest{
		Name:         r.GetName(),
		VPC:          r.VPC,
		Description:  fmt.Sprintf("sg for %s", r.GetName()),
		IngressRules: []securityGroup.IngressRules{ingressRule}}.Create(ctx)
	if err != nil {
		return err
	}
	result.SG = sg.SG
	return nil
}

func (r *Request) createOnDemand(ctx *pulumi.Context, amiID string,
	dh *ec2.DedicatedHost, result *Resources) error {
	instanceArgs := ec2.InstanceArgs{
		SubnetId:                 r.Subnets[0].ID(),
		Ami:                      pulumi.String(amiID),
		InstanceType:             pulumi.String(r.Specs.InstaceTypes[0]),
		KeyName:                  result.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(r.Public),
		VpcSecurityGroupIds:      pulumi.StringArray{result.SG.ID()},
		Tags: pulumi.StringMap{
			"Name":    pulumi.String(r.GetName()),
			"HOST_ID": pulumi.String(r.Specs.ID),
		},
	}
	if dh != nil {
		instanceArgs.HostId = dh.ID()
	}
	i, err := ec2.NewInstance(ctx, r.GetName(), &instanceArgs)
	if err != nil {
		return err
	}
	result.Instance = i
	result.Username = r.Specs.AMI.DefaultUser
	ctx.Export(result.OutputHost(),
		util.If(r.Public,
			i.PublicIp,
			i.PrivateIp))
	return nil
}

func (r Request) createSpotInstance(ctx *pulumi.Context,
	amiID string, result *Resources) error {
	lt, err := ec2.NewLaunchTemplate(ctx,
		r.GetName(),
		&ec2.LaunchTemplateArgs{
			NamePrefix: pulumi.String(r.GetName()),
			ImageId:    pulumi.String(amiID),
			KeyName:    result.AWSKeyPair.KeyName,
			NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
				&ec2.LaunchTemplateNetworkInterfaceArgs{
					SecurityGroups: pulumi.StringArray{result.SG.ID()},
				},
			},
		})
	if err != nil {
		return err
	}
	nlb, err := lb.NewLoadBalancer(ctx,
		r.GetName(),
		&lb.LoadBalancerArgs{
			LoadBalancerType: pulumi.String("network"),
			Subnets:          pulumi.StringArray{r.Subnets[0].ID()},
		})
	if err != nil {
		return err
	}
	rhelTargetGroup, err := lb.NewTargetGroup(ctx, r.GetName(),
		&lb.TargetGroupArgs{
			Port:     pulumi.Int(22),
			Protocol: pulumi.String("TCP"),
			VpcId:    r.VPC.ID(),
		})
	if err != nil {
		return err
	}
	_, err = lb.NewListener(ctx,
		r.GetName(),
		&lb.ListenerArgs{
			LoadBalancerArn: nlb.Arn,
			Port:            pulumi.Int(22),
			Protocol:        pulumi.String("TCP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: rhelTargetGroup.Arn,
				},
			},
		})
	if err != nil {
		return err
	}
	result.InstanceIP = nlb.DnsName
	result.Username = r.Specs.AMI.DefaultUser
	overrides := autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArray{}
	for _, instanceType := range r.Specs.InstaceTypes {
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
	_, err = autoscaling.NewGroup(ctx,
		r.GetName(),
		&autoscaling.GroupArgs{
			TargetGroupArns:      pulumi.ToStringArrayOutput([]pulumi.StringOutput{rhelTargetGroup.Arn}),
			CapacityRebalance:    pulumi.Bool(true),
			DesiredCapacity:      pulumi.Int(1),
			MaxSize:              pulumi.Int(1),
			MinSize:              pulumi.Int(1),
			VpcZoneIdentifiers:   pulumi.StringArray{r.Subnets[0].ID()},
			MixedInstancesPolicy: mixedInstancesPolicy,
			Tags: autoscaling.GroupTagArray{
				&autoscaling.GroupTagArgs{
					Key:               pulumi.String("Name"),
					Value:             pulumi.String(r.GetName()),
					PropagateAtLaunch: pulumi.Bool(true),
				},
				&autoscaling.GroupTagArgs{
					Key:               pulumi.String("HOST_ID"),
					Value:             pulumi.String(r.Specs.ID),
					PropagateAtLaunch: pulumi.Bool(true),
				},
			},
		})
	if err != nil {
		return err
	}
	ctx.Export(result.OutputHost(), result.InstanceIP)
	return nil
}

func (c *Resources) remoteExec(ctx *pulumi.Context, cmdName, cmd string,
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
