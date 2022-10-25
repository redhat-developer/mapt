package rhel

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r RHELRequest) Create(ctx *pulumi.Context) (*RHELResources, error) {
	var rhel RHELResources
	awsKeyPair, privateKey, err := compute.ManageKeypair(ctx, r.keyPair, r.Name, OutputPrivateKey)
	if err != nil {
		return nil, err
	}
	rhel.AWSKeyPair = awsKeyPair
	rhel.PrivateKey = privateKey
	rhelIngressRule := securityGroup.SSH_TCP
	if r.Public {
		rhelIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	} else {
		rhelIngressRule.SG = r.BastionSG
	}
	sg, err := securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  "rhel sg group",
		IngressRules: []securityGroup.IngressRules{rhelIngressRule}}.Create(ctx)
	if err != nil {
		return nil, err
	}

	amiNameRegex := fmt.Sprintf(r.Specs.AMI.RegexPattern, r.VersionMajor)
	ami, err := ami.GetAMIByName(ctx, amiNameRegex, r.Specs.AMI.Filters)
	if err != nil {
		return nil, err
	}
	if len(r.SpotPrice) > 0 {
		err = r.spotInstance(ctx, ami.Id, awsKeyPair, sg.SG, &rhel)
		if err != nil {
			return nil, err
		}
	} else {
		err = r.onDemandInstance(ctx, ami.Id, awsKeyPair, sg.SG, &rhel)
		if err != nil {
			return nil, err
		}
	}
	ctx.Export(OutputUsername, pulumi.String(r.Specs.AMI.DefaultUser))
	if r.Public {
		return &rhel, rhel.waitForInit(ctx)
	}
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &rhel, nil
}

func (c RHELResources) waitForInit(ctx *pulumi.Context) error {
	instance := command.RemoteInstance{
		Instance:   c.Instance,
		InstanceIP: &c.InstanceIP,
		Username:   c.Username,
		PrivateKey: c.PrivateKey}
	return instance.RemoteExec(
		ctx,
		command.CommandPing,
		"rhel-WaitForConnect")
}

func (r RHELRequest) spotInstance(ctx *pulumi.Context,
	amiID string, keyPair *ec2.KeyPair, sg *ec2.SecurityGroup,
	rhel *RHELResources) error {
	lt, err := ec2.NewLaunchTemplate(ctx,
		r.Name,
		&ec2.LaunchTemplateArgs{
			NamePrefix: pulumi.String(r.Name),
			ImageId:    pulumi.String(amiID),
			// InstanceType: pulumi.String(supportmatrix.OL_RHEL.InstaceTypes[0]),
			KeyName: keyPair.KeyName,
			// VpcSecurityGroupIds: pulumi.StringArray{sg.SG.ID()},
			NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
				&ec2.LaunchTemplateNetworkInterfaceArgs{
					// AssociatePublicIpAddress: pulumi.String(r.Public),
					SecurityGroups: pulumi.StringArray{sg.ID()},
				},
			},
		})
	if err != nil {
		return err
	}

	// if r.Public {
	nlb, err := lb.NewLoadBalancer(ctx,
		r.Name,
		&lb.LoadBalancerArgs{
			LoadBalancerType: pulumi.String("network"),
			Subnets:          pulumi.StringArray{r.Subnets[0].ID()},
		})
	if err != nil {
		return err
	}
	rhelTargetGroup, err := lb.NewTargetGroup(ctx, r.Name,
		&lb.TargetGroupArgs{
			Port:     pulumi.Int(22),
			Protocol: pulumi.String("TCP"),
			VpcId:    r.VPC.ID(),
		})
	if err != nil {
		return err
	}
	_, err = lb.NewListener(ctx,
		r.Name,
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
	rhel.InstanceIP = nlb.DnsName
	rhel.Username = r.Specs.AMI.DefaultUser
	_, err = autoscaling.NewGroup(ctx,
		r.Name,
		&autoscaling.GroupArgs{
			TargetGroupArns:    pulumi.ToStringArrayOutput([]pulumi.StringOutput{rhelTargetGroup.Arn}),
			CapacityRebalance:  pulumi.Bool(true),
			DesiredCapacity:    pulumi.Int(1),
			MaxSize:            pulumi.Int(1),
			MinSize:            pulumi.Int(1),
			VpcZoneIdentifiers: pulumi.StringArray{r.Subnets[0].ID()},
			MixedInstancesPolicy: &autoscaling.GroupMixedInstancesPolicyArgs{
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
					Overrides: autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArray{
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String(supportmatrix.OL_RHEL.InstaceTypes[0]),
						},
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String(supportmatrix.OL_RHEL.InstaceTypes[1]),
						},
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String(supportmatrix.OL_RHEL.InstaceTypes[2]),
						},
					},
				},
			},
			Tags: autoscaling.GroupTagArray{
				&autoscaling.GroupTagArgs{
					Key:               pulumi.String("Name"),
					Value:             pulumi.String(r.Name),
					PropagateAtLaunch: pulumi.Bool(true),
				},
				&autoscaling.GroupTagArgs{
					Key:               pulumi.String("HOST_ID"),
					Value:             pulumi.String(supportmatrix.OL_RHEL.ID),
					PropagateAtLaunch: pulumi.Bool(true),
				},
			},
		})
	if err != nil {
		return err
	}
	ctx.Export(OutputPrivateIP, rhel.InstanceIP)
	return nil
}

func (r RHELRequest) onDemandInstance(ctx *pulumi.Context,
	amiID string, keyPair *ec2.KeyPair, sg *ec2.SecurityGroup,
	rhel *RHELResources) error {
	i, err := ec2.NewInstance(ctx,
		r.Name,
		&ec2.InstanceArgs{
			SubnetId:                 r.Subnets[0].ID(),
			Ami:                      pulumi.String(amiID),
			InstanceType:             pulumi.String(r.Specs.InstaceTypes[0]),
			KeyName:                  keyPair.KeyName,
			AssociatePublicIpAddress: pulumi.Bool(r.Public),
			VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
			Tags: pulumi.StringMap{
				"Name":    pulumi.String(r.Name),
				"HOST_ID": pulumi.String(supportmatrix.OL_RHEL.ID),
			},
		})
	if err != nil {
		return err
	}
	rhel.Instance = i
	rhel.Username = r.Specs.AMI.DefaultUser
	ctx.Export(OutputPrivateIP,
		util.If(r.Public,
			i.PublicIp,
			i.PrivateIp))
	return nil
}
