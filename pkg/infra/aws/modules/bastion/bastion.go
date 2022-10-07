package modules

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/keypair"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BastionRequest struct {
	ProjectName string
	Name        string
	HA          bool
	keyPair     *ec2.KeyPair
	// loadBalancer *lb.LoadBalancer
}

type BastionResources struct {
	LaunchTemplate *ec2.LaunchTemplate
	Instance       *ec2.Instance
	KeyPair        *ec2.KeyPair
	// contains value if key is created within this module
	KeyPEM []byte
}

const (
	bastionDefaultAMI          string = "amzn2-ami-hvm-*-x86_64-gp2"
	bastionDefaultInstanceType string = "t2.small"
	// bastionDefaultDeviceType   string = "gp2"
	// bastionDefaultDeviceSize   int    = 10
)

func (r BastionRequest) Create(ctx *pulumi.Context) (*BastionResources, error) {
	ami, err := ami.GetAMIByName(ctx, bastionDefaultAMI)
	if err != nil {
		return nil, err
	}
	keyPair, keyPem, err := r.manageKeypair(ctx)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s-%s", r.ProjectName, r.Name)
	var instance *ec2.Instance
	if !r.HA {
		instance, err = ec2.NewInstance(ctx,
			name,
			&ec2.InstanceArgs{
				Ami:          pulumi.String(ami.Id),
				InstanceType: pulumi.String(bastionDefaultInstanceType),
				Tags: pulumi.StringMap{
					"Name": pulumi.String(name),
				},
			})
		if err != nil {
			return nil, err
		}
	}

	return &BastionResources{
			KeyPair:  keyPair,
			KeyPEM:   keyPem,
			Instance: instance,
		},
		nil
}

func (r BastionRequest) manageKeypair(ctx *pulumi.Context) (*ec2.KeyPair, []byte, error) {
	if r.keyPair == nil {
		// create key
		keyResources, err := keypair.KeyPairRequest{
			ProjectName: r.ProjectName,
			Name:        r.Name}.Create(ctx)
		if err != nil {
			return nil, nil, err
		}
		return keyResources.KeyPair, keyResources.KeyPEM, nil
	}
	return r.keyPair, nil, nil
}

// func manageHA(ctx *pulumi.Context, name string, ami *ec2.LookupAmiResult, keypair *ec2.KeyPair) (*ec2.Instance, error) {
// 	lt, err := ec2.NewLaunchTemplate(ctx,
// 		name,
// 		&ec2.LaunchTemplateArgs{
// 			// BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
// 			// 	ec2.LaunchTemplateBlockDeviceMappingArgs{
// 			// 		Ebs: ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
// 			// 			VolumeType: pulumi.String(bastionDefaultDeviceType),
// 			// 			VolumeSize: pulumi.Int(bastionDefaultDeviceSize)}},
// 			// },
// 			// InstanceMarketOptions: ec2.LaunchTemplateInstanceMarketOptionsArgs{SpotOptions: ec2.LaunchTemplateInstanceMarketOptionsSpotOptionsArgs{}}
// 			NamePrefix:   pulumi.String(name),
// 			ImageId:      pulumi.String(ami.Id),
// 			InstanceType: pulumi.String(bastionDefaultInstanceType),
// 			KeyName:      keypair.KeyName,
// 		})
// 	if err != nil {
// 		return nil, err
// 	}
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
