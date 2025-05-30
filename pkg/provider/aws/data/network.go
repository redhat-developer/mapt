package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	filterVPCID               = "vpc-id"
	filterAssociationSubnetID = "association.subnet-id"
)

var (
	ErrNoDefaultVPC           = fmt.Errorf("no VPC marked as default")
	ErrNoRouteTableBySubnetID = fmt.Errorf("no Route table by association.subnet-id")
)

func GetRandomPublicSubnet(region string) (*string, error) {
	cfg, err := getConfig(region)
	if err != nil {
		return nil, err
	}
	ec2Client := ec2.NewFromConfig(cfg)
	vpcsOutput, err := ec2Client.DescribeVpcs(
		context.TODO(), &ec2.DescribeVpcsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("isDefault"),
					Values: []string{"true"},
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}
	if len(vpcsOutput.Vpcs) == 0 {
		return nil, ErrNoDefaultVPC
	}
	for _, v := range vpcsOutput.Vpcs {
		ps, err := getPublicSubnets(ec2Client, *v.VpcId)
		if err != nil {
			logging.Error(err)
			break
		}
		if len(ps) > 0 {
			return util.RandomItemFromArray(ps), nil
		}
	}
	return nil, fmt.Errorf("no public subnet can be found on a default VPC")
}

func getPublicSubnets(client *ec2.Client, vpcID string) (subnets []*string, err error) {
	subnetsOutput, err := client.DescribeSubnets(
		context.TODO(), &ec2.DescribeSubnetsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String(filterVPCID),
					Values: []string{vpcID},
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}
	noSubnetRoutetableCounter := 0
	for _, s := range subnetsOutput.Subnets {
		err := isPublic(client, *s.SubnetId)
		if err == nil {
			subnets = append(subnets, s.SubnetId)
			break
		} else if err == ErrNoRouteTableBySubnetID {
			noSubnetRoutetableCounter++
		}
	}
	// If none route table for any subnet we assume there is only main route table
	// and so any subnet should be public
	if noSubnetRoutetableCounter == len(subnetsOutput.Subnets) {
		subnets = util.ArrayConvert(
			subnetsOutput.Subnets,
			func(s ec2types.Subnet) *string {
				return s.SubnetId
			})
	}
	return
}

func isPublic(client *ec2.Client, subnetID string) error {
	routeTablesOutput, err := client.DescribeRouteTables(
		context.TODO(),
		&ec2.DescribeRouteTablesInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String(filterAssociationSubnetID),
					Values: []string{subnetID},
				},
			},
		})
	if err != nil {
		return fmt.Errorf("failed to describe route tables: %w", err)
	}
	if len(routeTablesOutput.RouteTables) == 0 {
		return ErrNoRouteTableBySubnetID
	}
	for _, routeTable := range routeTablesOutput.RouteTables {
		for _, route := range routeTable.Routes {
			gwID := *route.GatewayId
			if route.GatewayId != nil && len(gwID) > 0 && gwID[:2] == "igw" {
				return nil
			}
		}
	}
	return fmt.Errorf("no public subnet setup found")
}

type SubnetRequestArgs struct {
	Region, VpcId, AzId *string
}

// Get first subnet if azid is pass it will pick from it
// If vpc id is pass it will pick first from vpc id subnets
func GetSubnetID(args *SubnetRequestArgs) (*string, error) {
	cfg, err := getConfig(*args.Region)
	if err != nil {
		return nil, err
	}
	ec2Client := ec2.NewFromConfig(cfg)
	var filters []ec2types.Filter
	if args.VpcId != nil {
		filters = append(filters, ec2types.Filter{
			Name:   aws.String("vpc-id"),
			Values: []string{*args.VpcId}})
	}
	if args.AzId != nil {
		filters = append(filters, ec2types.Filter{
			Name:   aws.String("availability-zone-id"),
			Values: []string{*args.AzId}})
	}
	output, err := ec2Client.DescribeSubnets(
		context.TODO(),
		&ec2.DescribeSubnetsInput{
			Filters: filters,
		})
	if err != nil {
		return nil, err
	}
	if len(output.Subnets) == 1 {
		return output.Subnets[0].SubnetId, nil
	}
	// If we got several subnets (all for vpcId we get subnet random)
	return output.Subnets[util.Random(len(output.Subnets), 0)].SubnetId, nil
}
