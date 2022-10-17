package bastion

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/keypair"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r BastionRequest) Create(ctx *pulumi.Context) (*BastionResources, error) {
	awsKeyPair, privateKey, err := r.manageKeypair(ctx)
	if err != nil {
		return nil, err
	}
	sg, err := securityGroup.SGRequest{
		Name:         r.Name,
		VPC:          r.VPC,
		Description:  "bastion sg group",
		IngressRules: []securityGroup.IngressRules{securityGroup.SSH_TCP}}.Create(ctx)
	if err != nil {
		return nil, err
	}

	var instance *ec2.Instance
	if !r.HA {
		ami, err := ami.GetAMIByName(ctx, bastionDefaultAMI)
		if err != nil {
			return nil, err
		}
		instance, err = ec2.NewInstance(ctx,
			r.Name,
			&ec2.InstanceArgs{
				Tags: pulumi.StringMap{
					"Name": pulumi.String(r.Name),
				},
				SubnetId:                 r.PublicSubnets[0].ID(),
				Ami:                      pulumi.String(ami.Id),
				InstanceType:             pulumi.String(bastionDefaultInstanceType),
				KeyName:                  awsKeyPair.KeyName,
				VpcSecurityGroupIds:      pulumi.StringArray{sg.SG.ID()},
				AssociatePublicIpAddress: pulumi.Bool(true),
			})
		if err != nil {
			return nil, err
		}
		ctx.Export(OutputPublicIP, instance.PublicIp)
		ctx.Export(OutputUsername, pulumi.String(bastionDefaultAMIUser))
	}
	bastion := BastionResources{
		AWSKeyPair: awsKeyPair,
		PrivateKey: privateKey,
		Instance:   instance,
	}
	return &bastion, bastion.waitForInit(ctx)
	// return &bastion, nil
}

func (r BastionRequest) manageKeypair(ctx *pulumi.Context) (*ec2.KeyPair, *tls.PrivateKey, error) {
	if r.keyPair == nil {
		// create key
		keyResources, err := keypair.KeyPairRequest{
			Name: r.Name}.Create(ctx)
		if err != nil {
			return nil, nil, err
		}
		ctx.Export(OutputPrivateKey, keyResources.PrivateKey.PrivateKeyPem)
		return keyResources.AWSKeyPair, keyResources.PrivateKey, nil
	}
	return r.keyPair, nil, nil
}

func (b BastionResources) waitForInit(ctx *pulumi.Context) error {
	bastion := command.RemoteInstance{
		Instace:    b.Instance,
		Username:   bastionDefaultAMIUser,
		PrivateKey: b.PrivateKey}
	logging.Debug("Waiting for bastion to initialize")
	// return bastion.RemoteCommandAwait(bastion.RemoteCommand,
	// 	ctx, command.CommandPing)
	return bastion.RemoteCommand(ctx, command.CommandPing)
}

// func (r BastionRequest) getLaunchTemplate(ctx *pulumi.Context, sg *ec2.SecurityGroup, keyPair *ec2.KeyPair, ltName string) (*ec2.LaunchTemplate, error) {
// 	ami, err := ami.GetAMIByName(ctx, bastionDefaultAMI)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ec2.NewLaunchTemplate(ctx,
// 		ltName,
// 		&ec2.LaunchTemplateArgs{
// 			// BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
// 			// 	ec2.LaunchTemplateBlockDeviceMappingArgs{
// 			// 		Ebs: ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
// 			// 			VolumeType: pulumi.String(bastionDefaultDeviceType),
// 			// 			VolumeSize: pulumi.Int(bastionDefaultDeviceSize)}},
// 			// },
// 			// InstanceMarketOptions: ec2.LaunchTemplateInstanceMarketOptionsArgs{SpotOptions: ec2.LaunchTemplateInstanceMarketOptionsSpotOptionsArgs{}}
// 			NamePrefix:          pulumi.String(ltName),
// 			ImageId:             pulumi.String(ami.Id),
// 			InstanceType:        pulumi.String(bastionDefaultInstanceType),
// 			KeyName:             keyPair.KeyName,
// 			VpcSecurityGroupIds: pulumi.StringArray{sg.ID()},
// 		})
// }

// func bastionHA(ctx *pulumi.Context, name string, ami *ec2.LookupAmiResult, keypair *ec2.KeyPair) (*ec2.Instance, error) {
// lt, err := r.getLaunchTemplate(ctx, sg.SG, awsKeyPair, r.Name)
// if err != nil {
// 	return nil, err
// }
// 	_, err = autoscaling.NewGroup(ctx, "bar", &autoscaling.GroupArgs{
// 		DesiredCapacity: pulumi.Int(1),
// 		MaxSize:         pulumi.Int(1),
// 		MinSize:         pulumi.Int(1),
// 		LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
// 			Id:      lt.ID(),
// 			Version: pulumi.String("$Latest"),
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	// _, err = lb.NewLoadBalancer(ctx, "example", &lb.LoadBalancerArgs{
// 	// 	LoadBalancerType: pulumi.String("network"),
// 	// 	SubnetMappings: lb.LoadBalancerSubnetMappingArray{
// 	// 		&lb.LoadBalancerSubnetMappingArgs{
// 	// 			SubnetId:           pulumi.Any(aws_subnet.Example1.Id),
// 	// 			PrivateIpv4Address: pulumi.String("10.0.1.15"),
// 	// 		},
// 	// 		&lb.LoadBalancerSubnetMappingArgs{
// 	// 			SubnetId:           pulumi.Any(aws_subnet.Example2.Id),
// 	// 			PrivateIpv4Address: pulumi.String("10.0.2.15"),
// 	// 		},
// 	// 	},
// 	// })
// 	// _, err := elb.NewLoadBalancer(ctx, "bar", &elb.LoadBalancerArgs{})
// 	return nil, nil
// }
