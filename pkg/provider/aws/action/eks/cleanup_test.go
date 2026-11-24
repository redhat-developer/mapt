package eks

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	classictypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go"
)

func testIdentity() cleanupIdentity {
	return cleanupIdentity{Version: 1, AccountID: "123456789012", Region: "us-east-2", ClusterName: "test-cluster", VPCID: "vpc-test"}
}

func tags(values map[string]string) []ec2types.Tag {
	result := []ec2types.Tag{}
	for key, value := range values {
		result = append(result, ec2types.Tag{Key: sdkaws.String(key), Value: sdkaws.String(value)})
	}
	return result
}

type fakeEC2Cleanup struct {
	ec2CleanupAPI
	interfaces []ec2types.NetworkInterface
	volumes    []ec2types.Volume
	groups     []ec2types.SecurityGroup
	deleted    []string
	revoked    []string
	pages      int
}

func (f *fakeEC2Cleanup) DescribeNetworkInterfaces(_ context.Context, in *ec2.DescribeNetworkInterfacesInput, _ ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	f.pages++
	if in.NextToken == nil {
		return &ec2.DescribeNetworkInterfacesOutput{NextToken: sdkaws.String("next")}, nil
	}
	return &ec2.DescribeNetworkInterfacesOutput{NetworkInterfaces: f.interfaces}, nil
}

func (f *fakeEC2Cleanup) DeleteNetworkInterface(_ context.Context, in *ec2.DeleteNetworkInterfaceInput, _ ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error) {
	f.deleted = append(f.deleted, *in.NetworkInterfaceId)
	// Another actor may delete the same interface between listing and deletion.
	return nil, &smithy.GenericAPIError{Code: "InvalidNetworkInterfaceID.NotFound"}
}

func TestNetworkInterfaceCleanupOwnership(t *testing.T) {
	owned := tags(map[string]string{"cluster.k8s.amazonaws.com/name": "test-cluster", "eks:eni:owner": "amazon-vpc-cni"})
	base := ec2types.NetworkInterface{VpcId: sdkaws.String("vpc-test"), Status: ec2types.NetworkInterfaceStatusAvailable, TagSet: owned}
	var interfaces []ec2types.NetworkInterface
	for _, name := range []string{"owned", "attached", "requester", "foreign-cluster", "foreign-vpc", "untagged", "in-use", "endpoint", "nat", "nlb"} {
		eni := base
		eni.NetworkInterfaceId = sdkaws.String(name)
		switch name {
		case "attached":
			eni.Attachment = &ec2types.NetworkInterfaceAttachment{}
		case "requester":
			eni.RequesterManaged = sdkaws.Bool(true)
		case "foreign-cluster":
			eni.TagSet = tags(map[string]string{"cluster.k8s.amazonaws.com/name": "other", "eks:eni:owner": "amazon-vpc-cni"})
		case "foreign-vpc":
			eni.VpcId = sdkaws.String("vpc-other")
		case "untagged":
			eni.TagSet = nil
		case "in-use":
			eni.Status = ec2types.NetworkInterfaceStatusInUse
		case "endpoint":
			// Interface VPC endpoints are deleted by the final destroy phase.
			eni.InterfaceType = ec2types.NetworkInterfaceTypeVpcEndpoint
			eni.RequesterManaged = sdkaws.Bool(true)
		case "nat":
			eni.InterfaceType = ec2types.NetworkInterfaceTypeNatGateway
		case "nlb":
			// Network load balancer interfaces are released after the ELB delete.
			eni.InterfaceType = ec2types.NetworkInterfaceTypeNetworkLoadBalancer
			eni.RequesterManaged = sdkaws.Bool(true)
			eni.TagSet = nil
		}
		interfaces = append(interfaces, eni)
	}
	f := &fakeEC2Cleanup{interfaces: interfaces}
	c := awsCleaner{identity: testIdentity(), ec2: f}
	pending, err := c.networkInterfaces(context.Background())
	// The owned interface is deleted; endpoint and NAT gateway interfaces belong
	// to the final network phase and are neither reported nor waited on.
	if err != nil || !reflect.DeepEqual(f.deleted, []string{"owned"}) || len(pending) != 7 || f.pages != 2 {
		t.Fatalf("err=%v deleted=%v pending=%v pages=%d", err, f.deleted, pending, f.pages)
	}
}

func (f *fakeEC2Cleanup) DescribeVolumes(context.Context, *ec2.DescribeVolumesInput, ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{Volumes: f.volumes}, nil
}

func (f *fakeEC2Cleanup) DeleteVolume(_ context.Context, in *ec2.DeleteVolumeInput, _ ...func(*ec2.Options)) (*ec2.DeleteVolumeOutput, error) {
	f.deleted = append(f.deleted, *in.VolumeId)
	return &ec2.DeleteVolumeOutput{}, nil
}

func TestVolumeCleanupRequiresExactOwnershipAndDetachment(t *testing.T) {
	owned := tags(map[string]string{clusterOwnershipTag: "test-cluster", "ebs.csi.aws.com/cluster": "true"})
	f := &fakeEC2Cleanup{volumes: []ec2types.Volume{
		{VolumeId: sdkaws.String("owned"), Tags: owned, State: ec2types.VolumeStateAvailable},
		{VolumeId: sdkaws.String("attached"), Tags: owned, State: ec2types.VolumeStateInUse},
		{VolumeId: sdkaws.String("other"), Tags: tags(map[string]string{"projectName": "eks"}), State: ec2types.VolumeStateAvailable},
	}}
	c := awsCleaner{identity: testIdentity(), ec2: f}
	pending, err := c.volumes(context.Background())
	if err != nil || !reflect.DeepEqual(f.deleted, []string{"owned"}) || len(pending) != 2 {
		t.Fatalf("err=%v deleted=%v pending=%v", err, f.deleted, pending)
	}
}

func (f *fakeEC2Cleanup) DescribeSecurityGroups(context.Context, *ec2.DescribeSecurityGroupsInput, ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{SecurityGroups: f.groups}, nil
}

func (f *fakeEC2Cleanup) RevokeSecurityGroupIngress(_ context.Context, in *ec2.RevokeSecurityGroupIngressInput, _ ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	for _, rule := range in.IpPermissions {
		for _, pair := range rule.UserIdGroupPairs {
			f.revoked = append(f.revoked, *in.GroupId+"->"+*pair.GroupId)
		}
	}
	return &ec2.RevokeSecurityGroupIngressOutput{}, nil
}

func (f *fakeEC2Cleanup) DeleteSecurityGroup(_ context.Context, in *ec2.DeleteSecurityGroupInput, _ ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	f.deleted = append(f.deleted, *in.GroupId)
	return &ec2.DeleteSecurityGroupOutput{}, nil
}

func TestSecurityGroupCleanupRevokesForeignReferences(t *testing.T) {
	owned := tags(map[string]string{"elbv2.k8s.aws/cluster": "test-cluster"})
	f := &fakeEC2Cleanup{groups: []ec2types.SecurityGroup{
		{GroupId: sdkaws.String("sg-a"), VpcId: sdkaws.String("vpc-test"), Tags: owned, IpPermissions: []ec2types.IpPermission{{IpProtocol: sdkaws.String("tcp"), UserIdGroupPairs: []ec2types.UserIdGroupPair{{GroupId: sdkaws.String("sg-b")}, {GroupId: sdkaws.String("sg-foreign")}}}}},
		{GroupId: sdkaws.String("sg-b"), VpcId: sdkaws.String("vpc-test"), Tags: owned},
		{GroupId: sdkaws.String("sg-elb"), VpcId: sdkaws.String("vpc-test"), Tags: tags(map[string]string{"kubernetes.io/cluster/test-cluster": "owned"})},
		{GroupId: sdkaws.String("sg-foreign"), VpcId: sdkaws.String("vpc-test")},
		// A node group (managed by Pulumi) references the load balancer security group.
		{GroupId: sdkaws.String("sg-node"), VpcId: sdkaws.String("vpc-test"), IpPermissions: []ec2types.IpPermission{{IpProtocol: sdkaws.String("tcp"), FromPort: sdkaws.Int32(30000), ToPort: sdkaws.Int32(32767), UserIdGroupPairs: []ec2types.UserIdGroupPair{{GroupId: sdkaws.String("sg-elb")}}}}},
		{GroupId: sdkaws.String("sg-other-vpc"), VpcId: sdkaws.String("vpc-other"), Tags: owned},
	}}
	c := awsCleaner{identity: testIdentity(), ec2: f}
	_, err := c.securityGroups(context.Background())
	if err != nil || !reflect.DeepEqual(f.deleted, []string{"sg-a", "sg-b", "sg-elb"}) ||
		!reflect.DeepEqual(f.revoked, []string{"sg-a->sg-b", "sg-node->sg-elb"}) {
		t.Fatalf("err=%v deleted=%v revoked=%v", err, f.deleted, f.revoked)
	}
}

type fakeELBCleanup struct {
	elbCleanupAPI
	deleted   []string
	pages     int
	ownership map[string]map[string]string
}

func (f *fakeELBCleanup) DescribeLoadBalancers(_ context.Context, in *elasticloadbalancingv2.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	f.pages++
	if in.Marker == nil {
		return &elasticloadbalancingv2.DescribeLoadBalancersOutput{NextMarker: sdkaws.String("next")}, nil
	}
	return &elasticloadbalancingv2.DescribeLoadBalancersOutput{LoadBalancers: []elbtypes.LoadBalancer{
		{LoadBalancerArn: sdkaws.String("owned"), VpcId: sdkaws.String("vpc-test")},
		{LoadBalancerArn: sdkaws.String("foreign-cluster"), VpcId: sdkaws.String("vpc-test")},
		{LoadBalancerArn: sdkaws.String("foreign-vpc"), VpcId: sdkaws.String("vpc-other")},
	}}, nil
}

func (f *fakeELBCleanup) DescribeTags(_ context.Context, in *elasticloadbalancingv2.DescribeTagsInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	output := &elasticloadbalancingv2.DescribeTagsOutput{}
	for _, arn := range in.ResourceArns {
		description := elbtypes.TagDescription{ResourceArn: sdkaws.String(arn)}
		for key, value := range f.ownership[arn] {
			description.Tags = append(description.Tags, elbtypes.Tag{Key: sdkaws.String(key), Value: sdkaws.String(value)})
		}
		output.TagDescriptions = append(output.TagDescriptions, description)
	}
	return output, nil
}

func (f *fakeELBCleanup) DeleteLoadBalancer(_ context.Context, in *elasticloadbalancingv2.DeleteLoadBalancerInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DeleteLoadBalancerOutput, error) {
	f.deleted = append(f.deleted, *in.LoadBalancerArn)
	return &elasticloadbalancingv2.DeleteLoadBalancerOutput{}, nil
}

func TestLoadBalancerCleanupRequiresVPCAndCluster(t *testing.T) {
	f := &fakeELBCleanup{ownership: map[string]map[string]string{
		"owned":           {"elbv2.k8s.aws/cluster": "test-cluster"},
		"foreign-cluster": {"elbv2.k8s.aws/cluster": "other"},
	}}
	c := awsCleaner{identity: testIdentity(), elb: f}
	_, err := c.loadBalancers(context.Background())
	if err != nil || !reflect.DeepEqual(f.deleted, []string{"owned"}) || f.pages != 2 {
		t.Fatalf("err=%v deleted=%v pages=%d", err, f.deleted, f.pages)
	}
}

func TestLoadBalancerCleanupAcceptsInTreeOwnership(t *testing.T) {
	f := &fakeELBCleanup{ownership: map[string]map[string]string{
		"owned": {"kubernetes.io/cluster/test-cluster": "owned"},
	}}
	c := awsCleaner{identity: testIdentity(), elb: f}
	_, err := c.loadBalancers(context.Background())
	if err != nil || !reflect.DeepEqual(f.deleted, []string{"owned"}) {
		t.Fatalf("err=%v deleted=%v", err, f.deleted)
	}
}

type classicLB struct {
	name string
	vpc  string
}

type fakeClassicELBCleanup struct {
	lbs     []classicLB
	owned   map[string]bool
	deleted []string
	batches int
}

func (f *fakeClassicELBCleanup) DescribeLoadBalancers(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
	output := &elasticloadbalancing.DescribeLoadBalancersOutput{}
	for _, lb := range f.lbs {
		output.LoadBalancerDescriptions = append(output.LoadBalancerDescriptions, classictypes.LoadBalancerDescription{
			LoadBalancerName: sdkaws.String(lb.name),
			VPCId:            sdkaws.String(lb.vpc),
		})
	}
	return output, nil
}

func (f *fakeClassicELBCleanup) DescribeTags(_ context.Context, in *elasticloadbalancing.DescribeTagsInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeTagsOutput, error) {
	f.batches++
	output := &elasticloadbalancing.DescribeTagsOutput{}
	for _, name := range in.LoadBalancerNames {
		description := classictypes.TagDescription{LoadBalancerName: sdkaws.String(name)}
		if f.owned[name] {
			description.Tags = []classictypes.Tag{{Key: sdkaws.String("kubernetes.io/cluster/test-cluster"), Value: sdkaws.String("owned")}}
		}
		output.TagDescriptions = append(output.TagDescriptions, description)
	}
	return output, nil
}

func (f *fakeClassicELBCleanup) DeleteLoadBalancer(_ context.Context, in *elasticloadbalancing.DeleteLoadBalancerInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DeleteLoadBalancerOutput, error) {
	f.deleted = append(f.deleted, *in.LoadBalancerName)
	return &elasticloadbalancing.DeleteLoadBalancerOutput{}, nil
}

func TestClassicLoadBalancerCleanupRequiresVPCAndCluster(t *testing.T) {
	f := &fakeClassicELBCleanup{owned: map[string]bool{}}
	var want []string
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("owned-%02d", i)
		f.lbs = append(f.lbs, classicLB{name: name, vpc: "vpc-test"})
		f.owned[name] = true
		want = append(want, name)
	}
	f.lbs = append(f.lbs,
		classicLB{name: "foreign-cluster", vpc: "vpc-test"},
		classicLB{name: "foreign-vpc", vpc: "vpc-other"},
		classicLB{name: "other", vpc: "vpc-test"},
	)
	c := awsCleaner{identity: testIdentity(), classicELB: f}
	pending, err := c.classicLoadBalancers(context.Background())
	if err != nil || !reflect.DeepEqual(f.deleted, want) || len(pending) != 20 || f.batches != 2 {
		t.Fatalf("err=%v deleted=%v pending=%v batches=%d", err, f.deleted, pending, f.batches)
	}
}

func TestCleanupPolling(t *testing.T) {
	attempts := 0
	err := pollCleanup(context.Background(), time.Second, time.Millisecond, func(context.Context) ([]string, error) {
		attempts++
		if attempts == 1 {
			return []string{"eni-test"}, &smithy.GenericAPIError{Code: "DependencyViolation"}
		}
		if attempts == 2 {
			return []string{"eni-test"}, nil
		}
		return nil, nil
	})
	if err != nil || attempts != 3 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
	denied := &smithy.GenericAPIError{Code: "AccessDenied"}
	err = pollCleanup(context.Background(), time.Second, time.Millisecond, func(context.Context) ([]string, error) { return []string{"eni-test"}, denied })
	if !errors.Is(err, denied) || !strings.Contains(err.Error(), "eni-test") {
		t.Fatalf("permanent error lost: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = pollCleanup(ctx, time.Second, time.Millisecond, func(context.Context) ([]string, error) { t.Fatal("called after cancellation"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestCleanupPermissionErrorsNameTheRequiredPermissions(t *testing.T) {
	for _, code := range []string{"AccessDenied", "AccessDeniedException", "UnauthorizedOperation", "UnauthorizedAccess"} {
		denied := &smithy.GenericAPIError{Code: code, Message: "not authorized"}
		err := withPermissionsHint(denied)
		if !errors.Is(err, denied) || !strings.Contains(err.Error(), cleanupPermissionsHint) {
			t.Fatalf("code %s: hint or cause lost: %v", code, err)
		}
	}
	other := errors.New("dependency violation")
	if err := withPermissionsHint(other); err != other {
		t.Fatalf("unrelated error changed: %v", err)
	}
	if err := withPermissionsHint(nil); err != nil {
		t.Fatalf("nil error changed: %v", err)
	}
}
