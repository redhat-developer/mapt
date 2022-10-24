package rhel

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	supportmatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r RHELRequest) Create(ctx *pulumi.Context) (*RHELResources, error) {
	awsKeyPair, privateKey, err := compute.ManageKeypair(ctx, r.keyPair, r.Name, OutputPrivateKey)
	if err != nil {
		return nil, err
	}
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

	amiNameRegex := fmt.Sprintf(defaultAMIPattern, r.VersionMajor)
	ami, err := ami.GetAMIByName(ctx, amiNameRegex)
	if err != nil {
		return nil, err
	}
	var sir *ec2.SpotInstanceRequest
	var i *ec2.Instance
	if len(r.SpotPrice) > 0 {
		lt, err := ec2.NewLaunchTemplate(ctx,
			r.Name,
			&ec2.LaunchTemplateArgs{
				NamePrefix:   pulumi.String(r.Name),
				ImageId:      pulumi.String(ami.Id),
				InstanceType: pulumi.String(supportmatrix.OL_RHEL.InstaceTypes[0]),
				KeyName:      awsKeyPair.KeyName,
				// VpcSecurityGroupIds: pulumi.StringArray{sg.SG.ID()},
				NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
					&ec2.LaunchTemplateNetworkInterfaceArgs{
						AssociatePublicIpAddress: pulumi.String("true"),
						SecurityGroups:           pulumi.StringArray{sg.SG.ID()},
					},
				},
			})
		if err != nil {
			return nil, err
		}
		_, err = autoscaling.NewGroup(ctx,
			r.Name,
			&autoscaling.GroupArgs{
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
			})
		if err != nil {
			return nil, err
		}
		// ctx.Export(OutputPrivateIP,
		// 	util.If(r.Public,
		// 		sir.PublicIp,
		// 		sir.PrivateIp))
		ctx.Export(OutputPrivateIP, pulumi.String("asdasd"))
	} else {
		i, err = ec2.NewInstance(ctx,
			r.Name,
			&ec2.InstanceArgs{
				SubnetId:                 r.Subnets[0].ID(),
				Ami:                      pulumi.String(ami.Id),
				InstanceType:             pulumi.String(defaultInstanceType),
				KeyName:                  awsKeyPair.KeyName,
				AssociatePublicIpAddress: pulumi.Bool(r.Public),
				VpcSecurityGroupIds:      pulumi.StringArray{sg.SG.ID()},
				Tags: pulumi.StringMap{
					"Name": pulumi.String(r.Name),
				},
			})
		if err != nil {
			return nil, err
		}
		ctx.Export(OutputPrivateIP,
			util.If(r.Public,
				i.PublicIp,
				i.PrivateIp))
	}
	ctx.Export(OutputUsername, pulumi.String(defaultAMIUser))
	rhel := RHELResources{
		AWSKeyPair:          awsKeyPair,
		PrivateKey:          privateKey,
		Instance:            i,
		SpotInstanceRequest: sir,
	}
	// if r.Public {
	// 	return &rhel, rhel.waitForInit(ctx)
	// }
	// for private we need bastion support on commands
	// https://github.com/pulumi/pulumi-command/pull/132
	return &rhel, nil
}

// func (c RHELResources) waitForInit(ctx *pulumi.Context) error {
// 	instance := command.RemoteInstance{
// 		Instace:             c.Instance,
// 		SpotInstanceRequest: c.SpotInstanceRequest,
// 		Username:            defaultAMIUser,
// 		PrivateKey:          c.PrivateKey}
// 	return instance.RemoteExec(
// 		ctx,
// 		command.CommandPing,
// 		"rhel-WaitForConnect")
// }
