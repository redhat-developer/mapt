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
	filterAvailabilityZone    = "availability-zone"
	filterAssociationSubnetID = "association.subnet-id"
)

var (
	ErrNoDefaultVPC           = fmt.Errorf("no VPC marked as default")
	ErrNoRouteTableBySubnetID = fmt.Errorf("no Route table by association.subnet-id")
)

func GetRandomPublicSubnet(ctx context.Context, region string) (*string, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	ec2Client := ec2.NewFromConfig(cfg)
	vpcsOutput, err := ec2Client.DescribeVpcs(
		ctx, &ec2.DescribeVpcsInput{
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
		ps, err := getPublicSubnets(ctx, ec2Client, *v.VpcId)
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

func getPublicSubnets(ctx context.Context, client *ec2.Client, vpcID string) (subnets []*string, err error) {
	subnetsOutput, err := client.DescribeSubnets(
		ctx, &ec2.DescribeSubnetsInput{
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
		err := isPublic(ctx, client, *s.SubnetId)
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

// GetSubnetAZsForVPC returns the unique AZ names of all subnets in the specified VPC.
// Both public and private subnets are included; callers that need public-only access
// should use GetPublicSubnetIDInAZ and handle the private-subnet fallback themselves.
func GetSubnetAZsForVPC(ctx context.Context, region, vpcID string) ([]string, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	subnetsOutput, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String(filterVPCID), Values: []string{vpcID}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets for VPC %s: %w", vpcID, err)
	}
	seen := map[string]struct{}{}
	var azs []string
	for _, s := range subnetsOutput.Subnets {
		if s.AvailabilityZone != nil {
			az := *s.AvailabilityZone
			if _, ok := seen[az]; !ok {
				seen[az] = struct{}{}
				azs = append(azs, az)
			}
		}
	}
	if len(azs) == 0 {
		return nil, fmt.Errorf("no subnets found in VPC %s", vpcID)
	}
	return azs, nil
}

// GetAnySubnetIDInAZ returns the first available subnet (public or private) in the
// given AZ within the specified VPC. Used as a fallback when no public subnet exists.
func GetAnySubnetIDInAZ(ctx context.Context, region, vpcID, az string) (*string, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	subnetsOutput, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String(filterVPCID), Values: []string{vpcID}},
			{Name: aws.String(filterAvailabilityZone), Values: []string{az}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets in VPC %s AZ %s: %w", vpcID, az, err)
	}
	if len(subnetsOutput.Subnets) == 0 {
		return nil, fmt.Errorf("no subnet found in VPC %s AZ %s", vpcID, az)
	}
	return subnetsOutput.Subnets[0].SubnetId, nil
}

// GetPublicSubnetIDInAZ returns a public subnet ID in the given AZ within the specified VPC.
func GetPublicSubnetIDInAZ(ctx context.Context, region, vpcID, az string) (*string, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	subnetsOutput, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String(filterVPCID), Values: []string{vpcID}},
			{Name: aws.String(filterAvailabilityZone), Values: []string{az}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets in VPC %s AZ %s: %w", vpcID, az, err)
	}
	for _, s := range subnetsOutput.Subnets {
		if err := isPublic(ctx, client, *s.SubnetId); err == nil {
			return s.SubnetId, nil
		}
	}
	return nil, fmt.Errorf("no public subnet found in VPC %s AZ %s", vpcID, az)
}

func isPublic(ctx context.Context, client *ec2.Client, subnetID string) error {
	routeTablesOutput, err := client.DescribeRouteTables(
		ctx,
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
			if route.GatewayId != nil {
				gwID := *route.GatewayId
				if len(gwID) > 0 && gwID[:2] == "igw" {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no public subnet setup found")
}
