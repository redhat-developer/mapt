package eks

import (
	"testing"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
)

type chartMocks struct {
	chartDeletedWith string
}

func (m *chartMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if args.TypeToken == "kubernetes:helm.sh/v4:Chart" {
		m.chartDeletedWith = args.RegisterRPC.GetDeletedWith()
	}
	return args.Name + "_id", args.Inputs, nil
}

func (m *chartMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{
		resource.PropertyKey("token"): resource.NewStringProperty("test-token"),
	}, nil
}

// The Chart component must carry the cluster as its DeletedWith resource.
// Children inherit it from the component, so deleting the cluster also removes
// the chart and its resources.
func TestLoadBalancerChartInheritsClusterLifecycle(t *testing.T) {
	mocks := &chartMocks{}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		region := "us-east-2"
		cluster, err := eks.NewCluster(ctx, "eks-cluster", &eks.ClusterArgs{
			RoleArn:   pulumi.String("arn:aws:iam::123456789012:role/eks-cluster"),
			VpcConfig: &eks.ClusterVpcConfigArgs{},
		})
		if err != nil {
			return err
		}
		vpc, err := ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{CidrBlock: pulumi.String("10.0.0.0/16")})
		if err != nil {
			return err
		}
		nodeGroup, err := ec2.NewVpc(ctx, "nodes", &ec2.VpcArgs{})
		if err != nil {
			return err
		}
		providerARN := pulumi.String("arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE").ToStringOutput()
		r := &eksRequest{
			mCtx:           mc.InitNoState(),
			allocationData: &allocation.AllocationResult{Region: &region},
		}
		return r.installAwsLoadBalancerController(ctx, providerARN, cluster, vpc, nodeGroup, nil)
	}, pulumi.WithMocks("project", "stack", mocks))
	if err != nil {
		t.Fatal(err)
	}
	// urn:pulumi:<stack>::<project>::<type>::<name>
	want := "urn:pulumi:stack::project::aws:eks/cluster:Cluster::eks-cluster"
	if mocks.chartDeletedWith != want {
		t.Fatalf("chart deletedWith=%q, want %q", mocks.chartDeletedWith, want)
	}
}
