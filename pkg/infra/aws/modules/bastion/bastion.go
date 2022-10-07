package modules

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BastionRequest struct {
	ProjectName string
	Name        string
	keyPair     *ec2.KeyPair
	// loadBalancer *lb.LoadBalancer
}

type BastionResources struct {
	LaunchTemplate ec2.LaunchTemplate
}

const (
	bastionDefaultAMI          string = "amzn2-ami-hvm-*-x86_64-gp2"
	bastionDefaultInstanceType string = "t2.small"
	// bastionDefaultDeviceType   string = "gp2"
	// bastionDefaultDeviceSize   int    = 10
)

func (r BastionRequest) Create(ctx *pulumi.Context) (*BastionResources, error) {

	// k, err := ec2.NewKeyPair(ctx, "deployer", &ec2.KeyPairArgs{
	// 	PublicKey: pulumi.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"),
	// })

	// if err != nil {
	// 	return nil, err
	// }
	ami, err := ami.GetAMIByName(ctx, bastionDefaultAMI)
	if err != nil {
		return nil, err
	}
	ltName := fmt.Sprintf("%s-%s", r.ProjectName, r.Name)
	lt, err := ec2.NewLaunchTemplate(ctx,
		ltName,
		&ec2.LaunchTemplateArgs{
			// BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			// 	ec2.LaunchTemplateBlockDeviceMappingArgs{
			// 		Ebs: ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
			// 			VolumeType: pulumi.String(bastionDefaultDeviceType),
			// 			VolumeSize: pulumi.Int(bastionDefaultDeviceSize)}},
			// },
			// InstanceMarketOptions: ec2.LaunchTemplateInstanceMarketOptionsArgs{SpotOptions: ec2.LaunchTemplateInstanceMarketOptionsSpotOptionsArgs{}}
			NamePrefix:   pulumi.String(ltName),
			ImageId:      pulumi.String(ami.Id),
			InstanceType: pulumi.String(bastionDefaultInstanceType),
			KeyName:      r.keyPair.KeyName,
		})
	if err != nil {
		return nil, err
	}
	_, err = autoscaling.NewGroup(ctx, "bar", &autoscaling.GroupArgs{
		DesiredCapacity: pulumi.Int(1),
		MaxSize:         pulumi.Int(1),
		MinSize:         pulumi.Int(1),
		LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
			Id:      lt.ID(),
			Version: pulumi.String("$Latest"),
		},
	})
	if err != nil {
		return nil, err
	}
	// _, err = lb.NewLoadBalancer(ctx, "example", &lb.LoadBalancerArgs{
	// 	LoadBalancerType: pulumi.String("network"),
	// 	SubnetMappings: lb.LoadBalancerSubnetMappingArray{
	// 		&lb.LoadBalancerSubnetMappingArgs{
	// 			SubnetId:           pulumi.Any(aws_subnet.Example1.Id),
	// 			PrivateIpv4Address: pulumi.String("10.0.1.15"),
	// 		},
	// 		&lb.LoadBalancerSubnetMappingArgs{
	// 			SubnetId:           pulumi.Any(aws_subnet.Example2.Id),
	// 			PrivateIpv4Address: pulumi.String("10.0.2.15"),
	// 		},
	// 	},
	// })
	// _, err := elb.NewLoadBalancer(ctx, "bar", &elb.LoadBalancerArgs{})
	// return &BastionResources{
	// 	LaunchTemplate: *lt,
	// }, nil
	return nil, nil
}
