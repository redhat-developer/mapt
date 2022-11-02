package compute

import (
	"fmt"
	"strconv"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/keypair"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *Request) GetName() string {
	//TODO review this to move fully to tags and avoid 32 limitation on resources limit
	name := fmt.Sprintf("%s-%s", r.ProjecName, r.Specs.ID)
	return util.If(len(name) > 15, name[:15], name)
}

func (r *Request) GetRequest() *Request {
	return r
}

func (r *Request) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, "", r.Specs.AMI.Filters)
}

func (r *Request) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *Request) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *Request) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *Request) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *Request) GetPostScript(ctx *pulumi.Context) (string, error) {
	return "", nil
}

func (r *Request) Create(ctx *pulumi.Context, computeRequested ComputeRequest) (*Compute, error) {
	// Manage keypairs for requested host
	compute := Compute{
		Name:  r.GetName(),
		Specs: r.Specs,
	}
	if err := r.manageKeypair(ctx, &compute); err != nil {
		return nil, err
	}
	// Create sg according to request params
	if err := r.manageSecurityGroup(ctx, computeRequested.CustomIngressRules(),
		&compute); err != nil {
		return nil, err
	}
	ami, err := computeRequested.GetAMI(ctx)
	if err != nil {
		return nil, err
	}
	userdataEncodedBase64, err := computeRequested.GetUserdata(ctx)
	if err != nil {
		return nil, err
	}
	if len(r.SpotPrice) > 0 {
		err = r.createSpotInstance(ctx, ami.Id, userdataEncodedBase64, &compute)
		if err != nil {
			return nil, err
		}
	} else {
		dh, err := computeRequested.GetDedicatedHost(ctx)
		if err != nil {
			return nil, err
		}
		err = r.createOnDemand(ctx, ami.Id, userdataEncodedBase64, dh, &compute)
		if err != nil {
			return nil, err
		}
	}
	ctx.Export(r.OutputUsername(), pulumi.String(r.Specs.AMI.DefaultUser))
	// if r.Public {
	// 	postScript, err := computeRequested.GetPostScript(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	waitCmddependencies := []pulumi.Resource{}
	// 	if len(postScript) > 0 {
	// 		rc, err := compute.remoteExec(ctx,
	// 			fmt.Sprintf("%s-%s", r.Specs.ID, "postscript"), postScript, nil)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		waitCmddependencies = append(waitCmddependencies, rc)
	// 	}
	// 	_, err = compute.remoteExec(ctx,
	// 		fmt.Sprintf("%s-%s", r.Specs.ID, "wait"), command.CommandPing, waitCmddependencies)
	// 	return &compute, err
	// }
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &compute, nil
}

func (r *Request) manageKeypair(ctx *pulumi.Context, result *Compute) error {
	if r.KeyPair == nil {
		// create key
		keyResources, err := keypair.KeyPairRequest{
			Name: r.GetName()}.Create(ctx)
		if err != nil {
			return err
		}
		result.AWSKeyPair = keyResources.AWSKeyPair
		result.PrivateKey = keyResources.PrivateKey
		r.PublicKeyOpenssh = keyResources.PrivateKey.PublicKeyOpenssh
		ctx.Export(r.OutputPrivateKey(), keyResources.PrivateKey.PrivateKeyPem)
		return nil
	}
	result.AWSKeyPair = r.KeyPair

	return nil
}

func (r *Request) manageSecurityGroup(ctx *pulumi.Context,
	customIngressRules []securityGroup.IngressRules, compute *Compute) error {
	ingressRule := securityGroup.SSH_TCP
	if r.Public {
		ingressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		ingressRule.SG = r.BastionSG
	}
	ingressRules := []securityGroup.IngressRules{ingressRule}
	if customIngressRules != nil {
		ingressRules = append(ingressRules, customIngressRules...)
	}

	sg, err := securityGroup.SGRequest{
		Name:         r.GetName(),
		VPC:          r.VPC,
		Description:  fmt.Sprintf("sg for %s", r.GetName()),
		IngressRules: ingressRules}.Create(ctx)
	if err != nil {
		return err
	}
	compute.SG = util.If(compute.SG == nil,
		[]*ec2.SecurityGroup{sg.SG},
		append(compute.SG, sg.SG))
	return nil
}

func (r *Request) createOnDemand(ctx *pulumi.Context, amiID string,
	udBase64 pulumi.StringPtrInput, dh *ec2.DedicatedHost, compute *Compute) error {
	instanceArgs := ec2.InstanceArgs{
		SubnetId:                 r.Subnets[0].ID(),
		Ami:                      pulumi.String(amiID),
		InstanceType:             pulumi.String(r.Specs.InstaceTypes[0]),
		KeyName:                  compute.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(r.Public),
		VpcSecurityGroupIds:      compute.getSecurityGroupsIDs(),
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(DefaultRootBlockDeviceSize),
		},
		Tags: pulumi.StringMap{
			"Name":    pulumi.String(r.GetName()),
			"HOST_ID": pulumi.String(r.Specs.ID),
		},
	}
	if dh != nil {
		instanceArgs.HostId = dh.ID()
	}
	if udBase64 != nil {
		instanceArgs.UserData = udBase64
	}
	i, err := ec2.NewInstance(ctx, r.GetName(), &instanceArgs)
	if err != nil {
		return err
	}
	compute.Instance = i
	compute.Username = r.Specs.AMI.DefaultUser
	ctx.Export(r.OutputHost(),
		util.If(r.Public,
			i.PublicIp,
			i.PrivateIp))
	return nil
}

func (r Request) createSpotInstance(ctx *pulumi.Context,
	amiID string, udBase64 pulumi.StringPtrInput, compute *Compute) error {
	args := &ec2.LaunchTemplateArgs{
		NamePrefix: pulumi.String(r.GetName()),
		ImageId:    pulumi.String(amiID),
		KeyName:    compute.AWSKeyPair.KeyName,
		NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
			&ec2.LaunchTemplateNetworkInterfaceArgs{
				SecurityGroups:           compute.getSecurityGroupsIDs(),
				AssociatePublicIpAddress: pulumi.String(strconv.FormatBool(r.Public)),
				SubnetId:                 r.Subnets[0].ID(),
			},
		},
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String(DefaultRootBlockDeviceName),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeSize: pulumi.Int(DefaultRootBlockDeviceSize),
				},
			},
		},
	}
	if udBase64 != nil {
		args.UserData = udBase64
	}
	lt, err := ec2.NewLaunchTemplate(ctx, r.GetName(), args)
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
	compute.InstanceIP = nlb.DnsName
	compute.Username = r.Specs.AMI.DefaultUser
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
	ctx.Export(r.OutputHost(), compute.InstanceIP)
	return nil
}

func (r *Request) OutputPrivateKey() string {
	return fmt.Sprintf("%s-%s", OutputPrivateKey, r.Specs.ID)
}

func (r *Request) OutputHost() string {
	return fmt.Sprintf("%s-%s", OutputHost, r.Specs.ID)
}

func (r *Request) OutputUsername() string {
	return fmt.Sprintf("%s-%s", OutputUsername, r.Specs.ID)
}

func (r *Request) OutputPassword() string {
	return fmt.Sprintf("%s-%s", OutputPasswordKey, r.Specs.ID)
}
