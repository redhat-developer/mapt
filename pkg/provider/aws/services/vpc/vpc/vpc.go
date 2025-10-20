package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
)

type VPCRequest struct {
	CIDR string
	Name string
}

type VPCResources struct {
	VPC                  *ec2.Vpc
	InternetGateway      *ec2.InternetGateway
	VpcGatewayAttachment *ec2.VpcGatewayAttachment
	SecurityGroup        *ec2.SecurityGroup
}

func (s VPCRequest) CreateNetwork(ctx *pulumi.Context, mCtx *mc.Context) (*VPCResources, error) {
	vName := fmt.Sprintf("%s-%s", "vpc", s.Name)
	v, err := ec2.NewVpc(ctx, vName,
		&ec2.VpcArgs{
			CidrBlock: pulumi.String(s.CIDR),
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}
	iName := fmt.Sprintf("%s-%s", "igw", s.Name)
	i, err := ec2.NewInternetGateway(ctx,
		iName,
		&ec2.InternetGatewayArgs{
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}
	// Create VPC Gateway Attachment
	vpcGatewayAttachment, err := ec2.NewVpcGatewayAttachment(ctx,
		fmt.Sprintf("%s-attachment", iName),
		&ec2.VpcGatewayAttachmentArgs{
			VpcId:             v.ID(),
			InternetGatewayId: i.ID(),
		})
	if err != nil {
		return nil, err
	}
	sgName := fmt.Sprintf("%s-%s", "default", s.Name)
	sg, err := ec2.NewSecurityGroup(ctx,
		fmt.Sprintf("%s-%s", sgName, s.Name),
		&ec2.SecurityGroupArgs{
			GroupDescription: pulumi.String("Default"),
			VpcId:           v.ID(),
			// Tags: mCtx.ResourceTags() // TODO: Convert to AWS Native tag format,
		})
	if err != nil {
		return nil, err
	}
	// Create self-referencing ingress rule
	_, err = ec2.NewSecurityGroupIngress(ctx,
		fmt.Sprintf("%s-ingress-self", sgName),
		&ec2.SecurityGroupIngressArgs{
			GroupId:               sg.ID(),
			SourceSecurityGroupId: sg.ID(),
			FromPort:              pulumi.Int(0),
			ToPort:                pulumi.Int(0),
			IpProtocol:            pulumi.String("-1"),
		})
	if err != nil {
		return nil, err
	}
	// Create egress rule for all traffic
	_, err = ec2.NewSecurityGroupEgress(ctx,
		fmt.Sprintf("%s-egress-all", sgName),
		&ec2.SecurityGroupEgressArgs{
			GroupId:    sg.ID(),
			FromPort:   pulumi.Int(0),
			ToPort:     pulumi.Int(0),
			IpProtocol: pulumi.String("-1"),
			CidrIp:     pulumi.String(infra.NETWORKING_CIDR_ANY_IPV4),
		})
	if err != nil {
		return nil, err
	}
	return &VPCResources{
			VPC:                  v,
			InternetGateway:      i,
			VpcGatewayAttachment: vpcGatewayAttachment,
			SecurityGroup:        sg},
		nil
}
