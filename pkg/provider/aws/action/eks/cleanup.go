package eks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkeks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// cleanupIdentity is non-secret and must outlive the cluster during teardown.
type cleanupIdentity struct {
	Version     int    `json:"version"`
	AccountID   string `json:"accountId"`
	Region      string `json:"region"`
	ClusterName string `json:"clusterName"`
	VPCID       string `json:"vpcId"`
}

func (i cleanupIdentity) validate() error {
	if i.Version != 1 || len(i.AccountID) != 12 || i.Region == "" || i.ClusterName == "" || !strings.HasPrefix(i.VPCID, "vpc-") {
		return fmt.Errorf("incomplete or unsupported EKS cleanup identity: %+v", i)
	}
	return nil
}

type ec2CleanupAPI interface {
	DescribeNetworkInterfaces(context.Context, *ec2.DescribeNetworkInterfacesInput, ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DeleteNetworkInterface(context.Context, *ec2.DeleteNetworkInterfaceInput, ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error)
	DescribeSecurityGroups(context.Context, *ec2.DescribeSecurityGroupsInput, ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	RevokeSecurityGroupIngress(context.Context, *ec2.RevokeSecurityGroupIngressInput, ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupEgress(context.Context, *ec2.RevokeSecurityGroupEgressInput, ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error)
	DeleteSecurityGroup(context.Context, *ec2.DeleteSecurityGroupInput, ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	DescribeVolumes(context.Context, *ec2.DescribeVolumesInput, ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DeleteVolume(context.Context, *ec2.DeleteVolumeInput, ...func(*ec2.Options)) (*ec2.DeleteVolumeOutput, error)
}

type elbCleanupAPI interface {
	DescribeLoadBalancers(context.Context, *elasticloadbalancingv2.DescribeLoadBalancersInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(context.Context, *elasticloadbalancingv2.DeleteLoadBalancerInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DeleteLoadBalancerOutput, error)
	DescribeTargetGroups(context.Context, *elasticloadbalancingv2.DescribeTargetGroupsInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DeleteTargetGroup(context.Context, *elasticloadbalancingv2.DeleteTargetGroupInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DeleteTargetGroupOutput, error)
	DescribeTags(context.Context, *elasticloadbalancingv2.DescribeTagsInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
}

type classicELBCleanupAPI interface {
	DescribeLoadBalancers(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)
	DescribeTags(context.Context, *elasticloadbalancing.DescribeTagsInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeTagsOutput, error)
	DeleteLoadBalancer(context.Context, *elasticloadbalancing.DeleteLoadBalancerInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DeleteLoadBalancerOutput, error)
}

type awsCleaner struct {
	identity   cleanupIdentity
	ec2        ec2CleanupAPI
	elb        elbCleanupAPI
	classicELB classicELBCleanupAPI
}

func newAWSCleaner(ctx context.Context, identity cleanupIdentity) (*awsCleaner, error) {
	if err := identity.validate(); err != nil {
		return nil, err
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(identity.Region))
	if err != nil {
		return nil, err
	}
	caller, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	if sdkaws.ToString(caller.Account) != identity.AccountID {
		return nil, fmt.Errorf("cleanup account mismatch: expected %s, got %s", identity.AccountID, sdkaws.ToString(caller.Account))
	}
	// A running controller could recreate resources deleted directly through AWS.
	_, err = sdkeks.NewFromConfig(cfg).DescribeCluster(ctx, &sdkeks.DescribeClusterInput{Name: &identity.ClusterName})
	if err == nil {
		return nil, fmt.Errorf("cluster %s still exists; finish cluster teardown before cleanup", identity.ClusterName)
	}
	if !awsErrorCode(err, "ResourceNotFoundException") {
		return nil, fmt.Errorf("checking cluster absence: %w", err)
	}
	return &awsCleaner{
		identity:   identity,
		ec2:        ec2.NewFromConfig(cfg),
		elb:        elasticloadbalancingv2.NewFromConfig(cfg),
		classicELB: elasticloadbalancing.NewFromConfig(cfg),
	}, nil
}

func (c *awsCleaner) cleanup(ctx context.Context) error {
	for _, phase := range []struct {
		name string
		run  func(context.Context) ([]string, error)
	}{
		{"load balancers", c.loadBalancers},
		{"classic load balancers", c.classicLoadBalancers},
		{"target groups", c.targetGroups},
		{"network interfaces", c.networkInterfaces},
		{"security groups", c.securityGroups},
		{"CSI volumes", c.volumes},
	} {
		logging.Infof("Cleaning up %s for cluster %s in %s", phase.name, c.identity.ClusterName, c.identity.Region)
		if err := pollCleanup(ctx, 10*time.Minute, 5*time.Second, phase.run); err != nil {
			return fmt.Errorf("cleanup %s: %w", phase.name, err)
		}
	}
	return nil
}

func pollCleanup(ctx context.Context, timeout, interval time.Duration, run func(context.Context) ([]string, error)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var last error
	for {
		if err := ctx.Err(); err != nil {
			return errors.Join(last, err)
		}
		pending, err := run(ctx)
		if err != nil && !retryCleanup(err) {
			if len(pending) > 0 {
				return fmt.Errorf("remaining resources %v: %w", pending, err)
			}
			return err
		}
		if err == nil && len(pending) == 0 {
			return nil
		}
		last = fmt.Errorf("remaining resources: %v", pending)
		if err != nil {
			last = errors.Join(last, err)
		}
		// Log only the pending IDs while retrying. Retryable AWS errors are
		// expected while dependencies are released and would read like failures;
		// the joined error is returned with the timeout if the phase never
		// completes.
		logging.Debugf("remaining resources: %v", pending)
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(last, ctx.Err())
		case <-timer.C:
		}
	}
}

func awsErrorCode(err error, codes ...string) bool {
	var apiError smithy.APIError
	if !errors.As(err, &apiError) {
		return false
	}
	for _, code := range codes {
		if apiError.ErrorCode() == code {
			return true
		}
	}
	return false
}

func missingResource(err error) bool {
	return awsErrorCode(err, "InvalidNetworkInterfaceID.NotFound", "InvalidGroup.NotFound", "InvalidVolume.NotFound", "LoadBalancerNotFound", "TargetGroupNotFound")
}

func retryCleanup(err error) bool {
	return awsErrorCode(err, "DependencyViolation", "ResourceInUse", "InvalidGroup.InUse", "VolumeInUse", "IncorrectState", "InvalidNetworkInterface.InUse") ||
		(awsErrorCode(err, "InvalidParameterValue") && strings.Contains(err.Error(), "in use"))
}

// ownsClusterTag reports whether a tag marks the resource as owned by this
// cluster. The load balancer controller uses elbv2.k8s.aws/cluster, while the
// in-tree cloud provider (classic ELBs and NLBs created from Service
// annotations) uses kubernetes.io/cluster/<name>.
func (c *awsCleaner) ownsClusterTag(key, value string) bool {
	return (key == "elbv2.k8s.aws/cluster" && value == c.identity.ClusterName) ||
		(key == "kubernetes.io/cluster/"+c.identity.ClusterName && value == "owned")
}

func (c *awsCleaner) elbOwned(ctx context.Context, arn string) (bool, error) {
	result, err := c.elb.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: []string{arn}})
	if missingResource(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, description := range result.TagDescriptions {
		if sdkaws.ToString(description.ResourceArn) != arn {
			continue
		}
		for _, tag := range description.Tags {
			if c.ownsClusterTag(sdkaws.ToString(tag.Key), sdkaws.ToString(tag.Value)) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *awsCleaner) classicLoadBalancers(ctx context.Context) ([]string, error) {
	var pending []string
	var names []string
	paginator := elasticloadbalancing.NewDescribeLoadBalancersPaginator(c.classicELB, &elasticloadbalancing.DescribeLoadBalancersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pending, err
		}
		for _, lb := range page.LoadBalancerDescriptions {
			if sdkaws.ToString(lb.VPCId) != c.identity.VPCID {
				continue
			}
			names = append(names, sdkaws.ToString(lb.LoadBalancerName))
		}
	}
	owned, err := c.classicOwned(ctx, names)
	if err != nil {
		return pending, err
	}
	for _, name := range names {
		if !owned[name] {
			continue
		}
		pending = append(pending, name)
		_, err := c.classicELB.DeleteLoadBalancer(ctx, &elasticloadbalancing.DeleteLoadBalancerInput{LoadBalancerName: &name})
		if err != nil && !missingResource(err) {
			return pending, err
		}
	}
	return pending, nil
}

// classicOwned reads tags in API-sized batches. A load balancer that disappears
// between listing and tagging is picked up by the next poll.
func (c *awsCleaner) classicOwned(ctx context.Context, names []string) (map[string]bool, error) {
	const batchSize = 20
	owned := map[string]bool{}
	for start := 0; start < len(names); start += batchSize {
		end := min(start+batchSize, len(names))
		result, err := c.classicELB.DescribeTags(ctx, &elasticloadbalancing.DescribeTagsInput{LoadBalancerNames: names[start:end]})
		if missingResource(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, description := range result.TagDescriptions {
			name := sdkaws.ToString(description.LoadBalancerName)
			for _, tag := range description.Tags {
				if c.ownsClusterTag(sdkaws.ToString(tag.Key), sdkaws.ToString(tag.Value)) {
					owned[name] = true
				}
			}
		}
	}
	return owned, nil
}

func (c *awsCleaner) loadBalancers(ctx context.Context) ([]string, error) {
	var pending []string
	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(c.elb, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pending, err
		}
		for _, lb := range page.LoadBalancers {
			if sdkaws.ToString(lb.VpcId) != c.identity.VPCID {
				continue
			}
			arn := sdkaws.ToString(lb.LoadBalancerArn)
			owned, err := c.elbOwned(ctx, arn)
			if err != nil {
				return pending, err
			}
			if !owned {
				continue
			}
			pending = append(pending, arn)
			_, err = c.elb.DeleteLoadBalancer(ctx, &elasticloadbalancingv2.DeleteLoadBalancerInput{LoadBalancerArn: &arn})
			if err != nil && !missingResource(err) {
				return pending, err
			}
		}
	}
	return pending, nil
}

func (c *awsCleaner) targetGroups(ctx context.Context) ([]string, error) {
	var pending []string
	paginator := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(c.elb, &elasticloadbalancingv2.DescribeTargetGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pending, err
		}
		for _, group := range page.TargetGroups {
			if sdkaws.ToString(group.VpcId) != c.identity.VPCID {
				continue
			}
			arn := sdkaws.ToString(group.TargetGroupArn)
			owned, err := c.elbOwned(ctx, arn)
			if err != nil {
				return pending, err
			}
			if !owned {
				continue
			}
			pending = append(pending, arn)
			_, err = c.elb.DeleteTargetGroup(ctx, &elasticloadbalancingv2.DeleteTargetGroupInput{TargetGroupArn: &arn})
			if err != nil && !missingResource(err) {
				return pending, err
			}
		}
	}
	return pending, nil
}

func ec2Tag(tags []ec2types.Tag, key string) string {
	for _, tag := range tags {
		if sdkaws.ToString(tag.Key) == key {
			return sdkaws.ToString(tag.Value)
		}
	}
	return ""
}

// cleanupPhaseManaged reports whether Pulumi deletes the interface in the final
// destroy phase. Interface VPC endpoints and NAT gateways are stack resources,
// so waiting for them here would always time out.
func cleanupPhaseManaged(eni ec2types.NetworkInterface) bool {
	return eni.InterfaceType == ec2types.NetworkInterfaceTypeVpcEndpoint ||
		eni.InterfaceType == ec2types.NetworkInterfaceTypeNatGateway
}

func (c *awsCleaner) networkInterfaces(ctx context.Context) ([]string, error) {
	var pending []string
	paginator := ec2.NewDescribeNetworkInterfacesPaginator(c.ec2, &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2types.Filter{{Name: sdkaws.String("vpc-id"), Values: []string{c.identity.VPCID}}},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pending, err
		}
		for _, eni := range page.NetworkInterfaces {
			if sdkaws.ToString(eni.VpcId) != c.identity.VPCID || cleanupPhaseManaged(eni) {
				continue
			}
			id := sdkaws.ToString(eni.NetworkInterfaceId)
			// Even an unowned interface blocks VPC teardown. Report it, never detach it.
			pending = append(pending, fmt.Sprintf("%s (%s, %s)", id, eni.Status, sdkaws.ToString(eni.Description)))
			if eni.Status != ec2types.NetworkInterfaceStatusAvailable || eni.Attachment != nil || sdkaws.ToBool(eni.RequesterManaged) ||
				ec2Tag(eni.TagSet, "cluster.k8s.amazonaws.com/name") != c.identity.ClusterName || ec2Tag(eni.TagSet, "eks:eni:owner") != "amazon-vpc-cni" {
				continue
			}
			_, err := c.ec2.DeleteNetworkInterface(ctx, &ec2.DeleteNetworkInterfaceInput{NetworkInterfaceId: &id})
			if err != nil && !missingResource(err) {
				return pending, err
			}
		}
	}
	return pending, nil
}

func (c *awsCleaner) groupOwned(group ec2types.SecurityGroup) bool {
	return sdkaws.ToString(group.VpcId) == c.identity.VPCID &&
		(ec2Tag(group.Tags, "elbv2.k8s.aws/cluster") == c.identity.ClusterName ||
			ec2Tag(group.Tags, "kubernetes.io/cluster/"+c.identity.ClusterName) == "owned")
}

func (c *awsCleaner) securityGroups(ctx context.Context) ([]string, error) {
	var groups []ec2types.SecurityGroup
	paginator := ec2.NewDescribeSecurityGroupsPaginator(c.ec2, &ec2.DescribeSecurityGroupsInput{
		Filters: []ec2types.Filter{{Name: sdkaws.String("vpc-id"), Values: []string{c.identity.VPCID}}},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		groups = append(groups, page.SecurityGroups...)
	}
	owned := map[string]bool{}
	var ownedGroups []ec2types.SecurityGroup
	for _, group := range groups {
		if c.groupOwned(group) {
			owned[sdkaws.ToString(group.GroupId)] = true
			ownedGroups = append(ownedGroups, group)
		}
	}
	var pending []string
	for _, group := range ownedGroups {
		pending = append(pending, sdkaws.ToString(group.GroupId))
	}
	// Break references to owned groups from any group in the VPC. The load
	// balancer controller and the in-tree cloud provider add such rules to node
	// groups, which must be revoked before the referenced group can be deleted.
	for _, group := range groups {
		id := sdkaws.ToString(group.GroupId)
		for _, ingress := range []bool{true, false} {
			permissions := group.IpPermissions
			if !ingress {
				permissions = group.IpPermissionsEgress
			}
			for _, permission := range permissions {
				var pairs []ec2types.UserIdGroupPair
				for _, pair := range permission.UserIdGroupPairs {
					if owned[sdkaws.ToString(pair.GroupId)] {
						pairs = append(pairs, pair)
					}
				}
				if len(pairs) == 0 {
					continue
				}
				rule := ec2types.IpPermission{IpProtocol: permission.IpProtocol, FromPort: permission.FromPort, ToPort: permission.ToPort, UserIdGroupPairs: pairs}
				var err error
				if ingress {
					_, err = c.ec2.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{GroupId: &id, IpPermissions: []ec2types.IpPermission{rule}})
				} else {
					_, err = c.ec2.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{GroupId: &id, IpPermissions: []ec2types.IpPermission{rule}})
				}
				if err != nil && !missingResource(err) && !awsErrorCode(err, "InvalidPermission.NotFound") {
					return pending, err
				}
			}
		}
	}
	for _, group := range ownedGroups {
		_, err := c.ec2.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{GroupId: group.GroupId})
		if err != nil && !missingResource(err) {
			return pending, err
		}
	}
	return pending, nil
}

func volumeClaim(tags []ec2types.Tag) string {
	claim := ec2Tag(tags, "kubernetes.io/created-for/pvc/name")
	if claim == "" {
		return ""
	}
	if namespace := ec2Tag(tags, "kubernetes.io/created-for/pvc/namespace"); namespace != "" {
		return fmt.Sprintf(" (pvc %s/%s)", namespace, claim)
	}
	return fmt.Sprintf(" (pvc %s)", claim)
}

func (c *awsCleaner) volumes(ctx context.Context) ([]string, error) {
	var pending []string
	paginator := ec2.NewDescribeVolumesPaginator(c.ec2, &ec2.DescribeVolumesInput{
		Filters: []ec2types.Filter{
			{Name: sdkaws.String("tag:" + clusterOwnershipTag), Values: []string{c.identity.ClusterName}},
			{Name: sdkaws.String("tag:ebs.csi.aws.com/cluster"), Values: []string{"true"}},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pending, err
		}
		for _, volume := range page.Volumes {
			if ec2Tag(volume.Tags, clusterOwnershipTag) != c.identity.ClusterName || ec2Tag(volume.Tags, "ebs.csi.aws.com/cluster") != "true" {
				continue
			}
			id := sdkaws.ToString(volume.VolumeId)
			pending = append(pending, fmt.Sprintf("%s (%s)", id, volume.State))
			if volume.State != ec2types.VolumeStateAvailable || len(volume.Attachments) > 0 {
				continue
			}
			logging.Infof("Deleting CSI volume %s for cluster %s%s", id, c.identity.ClusterName, volumeClaim(volume.Tags))
			_, err := c.ec2.DeleteVolume(ctx, &ec2.DeleteVolumeInput{VolumeId: &id})
			if err != nil && !missingResource(err) {
				return pending, err
			}
		}
	}
	return pending, nil
}
