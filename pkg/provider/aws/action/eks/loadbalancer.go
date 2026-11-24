package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Review the controller IAM policy whenever this chart is updated.
// renovate: datasource=helm depName=aws-load-balancer-controller registryUrl=https://aws.github.io/eks-charts
const loadBalancerChartVersion = "3.6.0"

const loadBalancerServiceAccount = "aws-load-balancer-controller-sa"

func (r *eksRequest) installAwsLoadBalancerController(ctx *pulumi.Context, oidcProviderArn pulumi.StringOutput, eksCluster *eks.Cluster, vpc *ec2.Vpc, nodeGroup pulumi.Resource, extraDeps []pulumi.Resource) error {
	// Mint the short-lived provider token after nodes and core addons are ready,
	// rather than allowing a slow infrastructure rollout to consume its lifetime.
	deps := append([]pulumi.Resource{eksCluster, nodeGroup}, extraDeps...)
	auth := eks.GetClusterAuthOutput(ctx, eks.GetClusterAuthOutputArgs{Name: eksCluster.Name}, pulumi.DependsOn(deps))
	k8sProvider, err := kubernetes.NewProvider(ctx, "k8sProvider", &kubernetes.ProviderArgs{
		Kubeconfig: generateTokenKubeconfig(eksCluster.Endpoint, eksCluster.CertificateAuthority.Data().Elem(), auth.Token()),
	})
	if err != nil {
		return err
	}
	policy, err := iam.NewPolicy(ctx, "loadBalancerControllerPolicy", &iam.PolicyArgs{
		Policy: pulumi.String(getAwsLoadBalancerControllerIamPolicy()),
		Tags:   r.mCtx.ResourceTags(),
	})
	if err != nil {
		return err
	}
	serviceAccountName := pulumi.String(loadBalancerServiceAccount)
	role, err := iam.NewRole(ctx, "loadBalancerControllerRole", &iam.RoleArgs{
		NamePrefix:       pulumi.String("MaptLBCRole-"),
		AssumeRolePolicy: irsaTrustPolicy(oidcProviderArn, loadBalancerServiceAccount),
		Tags:             r.mCtx.ResourceTags(),
	})
	if err != nil {
		return err
	}
	attachment, err := iam.NewRolePolicyAttachment(ctx, "loadBalancerControllerPolicyAttachment", &iam.RolePolicyAttachmentArgs{
		Role: role.Name, PolicyArn: policy.Arn,
	})
	if err != nil {
		return err
	}
	serviceAccount, err := corev1.NewServiceAccount(ctx, "lbcK8sServiceAccount", &corev1.ServiceAccountArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: serviceAccountName, Namespace: pulumi.String("kube-system"),
			Annotations: pulumi.StringMap{"eks.amazonaws.com/role-arn": role.Arn},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DeletedWith(eksCluster))
	if err != nil {
		return err
	}
	// Helm v4 orders CRDs before their custom resources and awaits establishment.
	// Children inherit DeletedWith from this component, so the chart is removed
	// together with the cluster rather than ahead of it.
	_, err = helmv4.NewChart(ctx, "aws-load-balancer-controller", &helmv4.ChartArgs{
		Chart: pulumi.String("aws-load-balancer-controller"), Version: pulumi.String(loadBalancerChartVersion),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{Repo: pulumi.String("https://aws.github.io/eks-charts")},
		Namespace:      pulumi.String("kube-system"),
		Values: pulumi.Map{
			"clusterName": eksCluster.Name,
			"defaultTags": eksCluster.Name.ApplyT(func(name string) map[string]string {
				return controllerTags(r.mCtx.GetTags(), name)
			}),
			"serviceAccount": pulumi.Map{"create": pulumi.Bool(false), "name": serviceAccountName},
			"region":         pulumi.String(*r.allocationData.Region), "vpcId": vpc.ID(),
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn(append([]pulumi.Resource{eksCluster, nodeGroup, attachment, serviceAccount}, extraDeps...)),
		pulumi.DeletedWith(eksCluster))
	return err
}
