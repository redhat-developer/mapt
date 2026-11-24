package eks

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
)

type securityGroupTagMocks struct {
	tags         map[string]string
	resourceIDs  map[string]string
	deletedWiths map[string]string
}

func (m *securityGroupTagMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if args.TypeToken == "aws:ec2/tag:Tag" {
		key := args.Inputs["key"].StringValue()
		m.tags[key] = args.Inputs["value"].StringValue()
		m.resourceIDs[key] = args.Inputs["resourceId"].StringValue()
		m.deletedWiths[key] = args.RegisterRPC.GetDeletedWith()
	}
	return args.Name + "_id", args.Inputs, nil
}

func (m *securityGroupTagMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

// EKS does not propagate cluster tags to the security group it creates, so mapt
// tags the group for every tag, including the ones passed through the CLI.
func TestClusterSecurityGroupTags(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "us-east-2")
	mCtx, err := mc.Init(&mc.ContextArgs{
		Context:     context.Background(),
		ProjectName: "eks-tags-test",
		Tags:        map[string]string{"build-id": "123"},
	}, aws.Provider())
	if err != nil {
		t.Fatal(err)
	}
	mocks := &securityGroupTagMocks{
		tags:         map[string]string{},
		resourceIDs:  map[string]string{},
		deletedWiths: map[string]string{},
	}
	err = pulumi.RunErr(func(ctx *pulumi.Context) error {
		cluster, err := eks.NewCluster(ctx, "eks-cluster", &eks.ClusterArgs{
			Name:    pulumi.String("eks-cluster-test"),
			RoleArn: pulumi.String("arn:aws:iam::123456789012:role/eks-cluster"),
			VpcConfig: &eks.ClusterVpcConfigArgs{
				ClusterSecurityGroupId: pulumi.String("sg-cluster"),
			},
		})
		if err != nil {
			return err
		}
		prefix := "mapt"
		return (&eksRequest{mCtx: mCtx, prefix: &prefix}).tagClusterSecurityGroup(ctx, cluster)
	}, pulumi.WithMocks("project", "stack", mocks))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"origin":            "mapt",
		"projectName":       "eks-tags-test",
		"build-id":          "123",
		"mapt-cluster-name": "eks-cluster-test",
	}
	wantDeletedWith := "urn:pulumi:stack::project::aws:eks/cluster:Cluster::eks-cluster"
	for key, value := range want {
		if mocks.tags[key] != value {
			t.Fatalf("tag %s=%q, want %q", key, mocks.tags[key], value)
		}
		if mocks.resourceIDs[key] != "sg-cluster" {
			t.Fatalf("tag %s targets %q, want sg-cluster", key, mocks.resourceIDs[key])
		}
		if mocks.deletedWiths[key] != wantDeletedWith {
			t.Fatalf("tag %s deletedWith=%q, want %q", key, mocks.deletedWiths[key], wantDeletedWith)
		}
	}
	if len(mocks.tags) != len(want) {
		t.Fatalf("unexpected tags: %v", mocks.tags)
	}
}
