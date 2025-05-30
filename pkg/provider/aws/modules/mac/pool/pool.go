package pool

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	awsxecs "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util"

	network "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	// "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/subnet"
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

	Vpc     ec2.VpcOutput           `pulumi:"vpc"`
	Subnets []ec2.SubnetOutput      `pulumi:"subnets"`
	SshSG   ec2.SecurityGroupOutput `pulumi:"sshSG"`

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
	azs := data.GetAvailabilityZones(p.Region)
	nr, err := network.NetworkRequest{
		Name:               p.Name,
		Region:             p.Region,
		AvailabilityZones:  azs,
		CIDR:               network.DefaultCIDRNetwork,
		PublicSubnetsCIDRs: network.DefaultCIDRPublicSubnets[:len(azs)],
		NatGatewayType:     network.NONE,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, err
	}
	sshRunnerSG, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(p.Name, awsMacPoolID, "ssh-runner-sg"),
		VPC:         nr.VPCResources.VPC,
		Description: fmt.Sprintf("sg for %s", awsMacPoolID),
	}.Create(ctx)
	if err != nil {
		return nil, err
	}
	// Outputs
	res.Vpc = nr.VPCResources.VPC.ToVpcOutput()
	res.Subnets = util.ArrayConvert(nr.PublicSNResources,
		func(sn *subnet.PublicSubnetResources) ec2.SubnetOutput {
			return sn.Subnet.ToSubnetOutput()
		})
	res.SshSG = sshRunnerSG.SG.ToSecurityGroupOutput()
	// Tasks
	res.TaskHouseKeep = pulumi.All(nr.VPCResources.VPC.ID(), nr.PublicSNResources[0].Subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return houseKeeperTaskSpecScheduler(ctx, p, &vpcID, &subnetID, &sgID)
		}).(awsxecs.FargateTaskDefinitionOutput)
	res.TaskRequest = pulumi.All(nr.VPCResources.VPC.ID(), nr.PublicSNResources[0].Subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return requestTaskSpec(ctx, p, &vpcID, &subnetID, &sgID)
		}).(awsxecs.FargateTaskDefinitionOutput)
	res.TaskRelease = pulumi.All(nr.VPCResources.VPC.ID(), nr.PublicSNResources[0].Subnet.ID(), sshRunnerSG.SG.ID()).ApplyT(
		func(args []any) (*awsxecs.FargateTaskDefinition, error) {
			vpcID := string(args[0].(pulumi.ID))
			subnetID := string(args[1].(pulumi.ID))
			sgID := string(args[2].(pulumi.ID))
			return releaseTaskSpec(ctx, p, &vpcID, &subnetID, &sgID)
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
		"vpc":             res.Vpc,
		"sshSG":           res.SshSG,
		"taskHouseKeep":   res.TaskHouseKeep,
		"taskRequest":     res.TaskRequest,
		"taskRelease":     res.TaskRelease,
		"clientAK":        ak.ID(),
		"clientSK":        ak.Secret,
	})
}
