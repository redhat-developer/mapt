package pool

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	customResourceTypePool = "rh:rd:mapt:aws:mac:pool"
	awsMacPoolID           = "amp"
)

type PoolArgs struct {
	Region          string
	Name            string
	Arch            string
	OSVersion       string
	OfferedCapacity int
	MaxSize         int
}

type Pool struct {
	pulumi.ResourceState

	Region pulumi.StringOutput `pulumi:"region"`

	Name            pulumi.StringOutput `pulumi:"name"`
	Arch            pulumi.StringOutput `pulumi:"arch"`
	OSVersion       pulumi.StringOutput `pulumi:"osVersion"`
	OfferedCapacity pulumi.IntOutput    `pulumi:"offeredCapacity"`
	MaxSize         pulumi.IntOutput    `pulumi:"maxSize"`

	AzId pulumi.StringOutput `pulumi:"azId"`

	Vpc    ec2.VpcOutput           `pulumi:"vpc"`
	Subnet ec2.SubnetOutput        `pulumi:"subnet"`
	SshSG  ec2.SecurityGroupOutput `pulumi:"sshSG"`

	TaskHouseKeep awsxecs.FargateTaskDefinitionOutput `pulumi:"taskHouseKeep"`
	TaskRequest   awsxecs.FargateTaskDefinitionOutput `pulumi:"taskRequest"`
	TaskRelease   awsxecs.FargateTaskDefinitionOutput `pulumi:"taskRelease"`

	ClientAK pulumi.StringOutput `pulumi:"clientAK"`
	ClientSK pulumi.StringOutput `pulumi:"clientSK"`
}

func NewPool(ctx *pulumi.Context, name string, p *PoolArgs, opts ...pulumi.ResourceOption) (*Pool, error) {
	var err error
	res := &Pool{
		Region:          pulumi.String(p.Region).ToStringOutput(),
		Name:            pulumi.String(p.Name).ToStringOutput(),
		Arch:            pulumi.String(p.Arch).ToStringOutput(),
		OSVersion:       pulumi.String(p.OSVersion).ToStringOutput(),
		OfferedCapacity: pulumi.Int(p.OfferedCapacity).ToIntOutput(),
		MaxSize:         pulumi.Int(p.MaxSize).ToIntOutput(),
	}
	if err = ctx.RegisterComponentResource(
		customResourceTypePool,
		name,
		res,
		opts...); err != nil {
		return nil, err
	}
	azID, err := data.GetRandomAvailabilityZone(p.Region, nil)
	if err != nil {
		return nil, err
	}
	res.AzId = pulumi.String(*azID).ToStringOutput()
	nonLB := false
	nr := network.NetworkRequest{
		Prefix:                  p.Name,
		ID:                      awsMacPoolID,
		Region:                  p.Region,
		AZ:                      *azID,
		CreateLoadBalancer:      &nonLB,
		Airgap:                  false,
		AirgapPhaseConnectivity: network.ON,
	}
	vpc, subnet, _, _, _, err := nr.Network(ctx)
	if err != nil {
		return nil, err
	}
	sshRunnerSG, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(p.Name, awsMacPoolID, "ssh-runner-sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsMacPoolID),
	}.Create(ctx)
	if err != nil {
		return nil, err
	}
	// Resoruces Outputs
	res.Vpc = vpc.ToVpcOutput()
	res.Subnet = subnet.ToSubnetOutput()
	res.SshSG = sshRunnerSG.SG.ToSecurityGroupOutput()
	// Tasks
	res.TaskHouseKeep = pulumi.All(vpc.ID(), subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return houseKeeperTaskSpecScheduler(ctx, p, &vpcID, azID, &subnetID, &sgID)
		}).(awsxecs.FargateTaskDefinitionOutput)
	res.TaskRequest = pulumi.All(vpc.ID(), subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return requestTaskSpec(ctx, p, &vpcID, azID, &subnetID, &sgID)
		}).(awsxecs.FargateTaskDefinitionOutput)
	res.TaskRelease = pulumi.All(vpc.ID(), subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return releaseTaskSpec(ctx, p, &vpcID, azID, &subnetID, &sgID)
		}).(awsxecs.FargateTaskDefinitionOutput)

	_, ak, err := clientAccount(ctx, p.Name, p.Arch, p.OSVersion, nil)
	if err != nil {
		return nil, err
	}
	return res, ctx.RegisterResourceOutputs(res, pulumi.Map{
		"region":          res.Region,
		"name":            res.Name,
		"arch":            res.Arch,
		"osVersion":       res.OSVersion,
		"offeredCapacity": res.OfferedCapacity,
		"maxSize":         res.MaxSize,
		"azId":            res.AzId,
		"vpc":             res.Vpc,
		"subnet":          res.Subnet,
		"sshSG":           res.SshSG,
		"taskHouseKeep":   res.TaskHouseKeep,
		"taskRequest":     res.TaskRequest,
		"taskRelease":     res.TaskRelease,
		"clientAK":        ak.ID(),
		"clientSK":        ak.Secret,
	})
}

func (p *Pool) RunRemoteHouseKeep(name, arch, osVersion *string, offeredCapacity, maxSize *int) error {
	_ = pulumi.All(p.TaskHouseKeep, p.Vpc, p.Subnet, p.SshSG, p.AzId,
		p.Name, p.Arch, p.OSVersion, p.OfferedCapacity, p.MaxSize).ApplyT(
		func(args []any) error {
			task := args[0].(*awsxecs.FargateTaskDefinition)
			vpc := args[1].(*ec2.Vpc)
			subnet := args[2].(*ec2.Subnet)
			sshSG := args[3].(*ec2.SecurityGroup)
			azID := args[4].(string)
			pulumi.All(vpc.ID(), subnet.ID(), sshSG.ID(), task.TaskDefinition.Arn()).ApplyT(
				func(args []any) (err error) {
					vpcId := string(args[0].(pulumi.ID))
					subnetId := string(args[1].(pulumi.ID))
					sgId := string(args[2].(pulumi.ID))
					tArn := args[3].(string)
					return houseKeeperRemote(&tArn,
						name, arch, osVersion,
						offeredCapacity, maxSize,
						&vpcId, &azID, &subnetId, &sgId)
				})
			return nil
		})
	return nil
}

// func getRandomAZ(ctx *pulumi.Context, name, region *string) (pulumi.StringOutput, error) {
// 	azs, err := data.DescribeAvailabilityZones(*region)
// 	if err != nil {
// 		return pulumi.StringOutput{}, err
// 	}
// 	azIdx, err := random.NewRandomInteger(ctx,
// 		resourcesUtil.GetResourceName(*name, awsMacPoolID, "az-idx"),
// 		&random.RandomIntegerArgs{
// 			Min: pulumi.Int(0),
// 			Max: pulumi.Int(len(azs) - 1),
// 		})
// 	if err != nil {
// 		return pulumi.StringOutput{}, err
// 	}
// 	azId := azIdx.Result.ApplyT(func(idx int) string {
// 		return *azs[idx].ZoneId
// 	}).(pulumi.StringOutput)
// 	return azId, nil
// }

// type taskDef func(ctx *pulumi.Context, p *PoolArgs, vpcID, azID, subnetID, sgID *string) (*awsxecs.FargateTaskDefinition, error)

// func (p *Pool) task(ctx *pulumi.Context, azID *string, a *PoolArgs, tdf taskDef) awsxecs.FargateTaskDefinitionOutput {
// 	return pulumi.All(p.Vpc, p.Subnet, p.SSHSG).ApplyT(
// 		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
// 			vpc := args[0].(*ec2.Vpc)
// 			subnet := args[1].(*ec2.Subnet)
// 			sshSG := args[2].(*ec2.SecurityGroup)
// 			var td *awsxecs.FargateTaskDefinition
// 			pulumi.All(vpc.ID(), subnet.ID(), sshSG.ID()).ApplyT(
// 				func(args []any) (err error) {
// 					vpcId := string(args[0].(pulumi.ID))
// 					subnetId := string(args[1].(pulumi.ID))
// 					sgId := string(args[2].(pulumi.ID))
// 					td, err = tdf(ctx, a, &vpcId, azID, &subnetId, &sgId)
// 					return err
// 				})

// 			return td, nil
// 		}).(awsxecs.FargateTaskDefinitionOutput)
// }
