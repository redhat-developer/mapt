package subnet

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var validEndpoints = map[string]bool{"s3": true, "ecr": true, "ssm": true}

type EndpointsRequest struct {
	VPC              *ec2.Vpc
	Subnets          []*ec2.Subnet
	RouteTables      []*ec2.RouteTable
	Region           string
	Name             string
	ServiceEndpoints []string
}

func (r EndpointsRequest) Create(ctx *pulumi.Context) error {
	if len(r.ServiceEndpoints) == 0 {
		return nil
	}
	for _, e := range r.ServiceEndpoints {
		if !validEndpoints[e] {
			return fmt.Errorf("unknown VPC endpoint %q: accepted values are s3, ecr, ssm", e)
		}
	}

	// Create interface-endpoint security group only when needed
	needInterfaceSG := false
	for _, e := range r.ServiceEndpoints {
		if e == "ecr" || e == "ssm" {
			needInterfaceSG = true
			break
		}
	}
	var sg *ec2.SecurityGroup
	if needInterfaceSG {
		var err error
		sg, err = ec2.NewSecurityGroup(ctx,
			fmt.Sprintf("%s-%s", "endpoints", r.Name),
			&ec2.SecurityGroupArgs{
				VpcId: r.VPC.ID(),
				Ingress: ec2.SecurityGroupIngressArray{
					&ec2.SecurityGroupIngressArgs{
						Protocol:   pulumi.String("tcp"),
						FromPort:   pulumi.Int(443),
						ToPort:     pulumi.Int(443),
						CidrBlocks: pulumi.StringArray{r.VPC.CidrBlock},
					},
				},
			})
		if err != nil {
			return err
		}
	}

	routeTableIds := make(pulumi.StringArray, len(r.RouteTables))
	for i, rt := range r.RouteTables {
		routeTableIds[i] = rt.ID()
	}
	subnetIds := make(pulumi.StringArray, len(r.Subnets))
	for i, sn := range r.Subnets {
		subnetIds[i] = sn.ID()
	}

	for _, e := range r.ServiceEndpoints {
		switch e {
		case "s3":
			_, err := ec2.NewVpcEndpoint(ctx,
				fmt.Sprintf("%s-%s", "endpoint-s3", r.Name),
				&ec2.VpcEndpointArgs{
					VpcId:           r.VPC.ID(),
					ServiceName:     pulumi.Sprintf("com.amazonaws.%s.s3", r.Region),
					VpcEndpointType: pulumi.String("Gateway"),
					RouteTableIds:   routeTableIds,
				})
			if err != nil {
				return err
			}
		case "ecr":
			_, err := ec2.NewVpcEndpoint(ctx,
				fmt.Sprintf("%s-%s", "endpoint-ecr", r.Name),
				&ec2.VpcEndpointArgs{
					VpcId:            r.VPC.ID(),
					ServiceName:      pulumi.Sprintf("com.amazonaws.%s.ecr.dkr", r.Region),
					VpcEndpointType:  pulumi.String("Interface"),
					SubnetIds:        subnetIds,
					SecurityGroupIds: pulumi.StringArray{sg.ID()},
				})
			if err != nil {
				return err
			}
		case "ssm":
			_, err := ec2.NewVpcEndpoint(ctx,
				fmt.Sprintf("%s-%s", "endpoint-ssm", r.Name),
				&ec2.VpcEndpointArgs{
					VpcId:            r.VPC.ID(),
					ServiceName:      pulumi.Sprintf("com.amazonaws.%s.ssm", r.Region),
					VpcEndpointType:  pulumi.String("Interface"),
					SubnetIds:        subnetIds,
					SecurityGroupIds: pulumi.StringArray{sg.ID()},
				})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
