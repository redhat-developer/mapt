package eks

import (
	"encoding/json"
	"fmt"
	"net/url"

	awsProvider "github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	network "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	subnet "github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/subnet"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
)

type EKSRequest struct {
	Prefix                 string
	InstanceRequest        *instancetypes.AwsInstanceRequest
	KubernetesVersion      string
	ScalingDesiredSize     int
	ScalingMaxSize         int
	ScalingMinSize         int
	Spot                   bool
	Addons                 []string
	LoadBalancerController bool
	AvailabilityZones      []string
	AllocationData         *allocation.AllocationData
}

func Create(ctx *maptContext.ContextArgs, r *EKSRequest) (err error) {
	logging.Debug("Creating EKS")
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}

	// Get instance types matching requirements
	instanceTypes, err := r.InstanceRequest.GetMachineTypes()
	if err != nil {
		return err
	}
	if len(instanceTypes) == 0 {
		return fmt.Errorf("no instances matching criteria")
	}

	// Get allocation data based on spot flag
	projectName := maptContext.ProjectName()
	r.AllocationData, err = util.IfWithError(r.Spot,
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnSpot(
				&projectName,
				&amiProduct,
				nil,
				instanceTypes)
		},
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnDemand()
		})
	if err != nil {
		return err
	}
	r.AvailabilityZones = data.GetAvailabilityZones(*r.AllocationData.Region)

	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackName),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.AllocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.AllocationData.Region,
			}),
		DeployFunc:          r.deployer,
	}

	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

func Destroy(ctx *maptContext.ContextArgs) error {
	// Create mapt Context
	logging.Debug("Destroy EKS")
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			BackedURL: maptContext.BackedURL(),
			Stackname: stackName,
		})
}

// Main function to deploy all requried resources to AWS
func (r *EKSRequest) deployer(ctx *pulumi.Context) error {
	// Networking
	nr, err := network.NetworkRequest{
		Name:               resourcesUtil.GetResourceName(r.Prefix, awsEKSID, "net"),
		CIDR:               network.DefaultCIDRNetwork,
		AvailabilityZones:  r.AvailabilityZones,
		PublicSubnetsCIDRs: network.GeneratePublicSubnetCIDRs(len(r.AvailabilityZones)),
		Region:             *r.AllocationData.Region,
		SingleNatGateway:   true,
		MapPublicIp:        true,
	}.CreateNetwork(ctx)
	if err != nil {
		return err
	}
	vpc := nr.VPCResources.VPC
	subnetResources := nr.PublicSNResources
	subnetIds := pulumi.StringArray(util.ArrayConvert(subnetResources, func(s *subnet.PublicSubnetResources) pulumi.StringInput { return s.Subnet.ID() }))
	if err != nil {
		return err
	}

	eksRole, err := r.createEksRole(ctx)
	if err != nil {
		return err
	}
	// Create the EC2 NodeGroup Role
	nodeGroupRole, err := r.createNodeGroupRole(ctx)
	if err != nil {
		return err
	}

	// Security groups
	securityGroups, err := r.securityGroups(ctx, vpc)
	if err != nil {
		return err
	}
	// Create EKS Cluster
	eksCluster, err := eks.NewCluster(ctx, "eks-cluster", &eks.ClusterArgs{
		RoleArn: eksRole.Arn,
		VpcConfig: &eks.ClusterVpcConfigArgs{
			PublicAccessCidrs: pulumi.StringArray{
				pulumi.String("0.0.0.0/0"),
			},
			SecurityGroupIds: securityGroups,
			SubnetIds:        subnetIds,
		},
		Version: pulumi.String(r.KubernetesVersion),
	}, pulumi.DependsOn([]pulumi.Resource{eksRole}))
	if err != nil {
		return err
	}

	kubeconfig := generateKubeconfig(eksCluster.Endpoint, eksCluster.CertificateAuthority.Data().Elem(), eksCluster.Name)
	// Create a Kubernetes provider instance
	k8sProvider, err := kubernetes.NewProvider(ctx, "k8sProvider", &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster}))
	if err != nil {
		return err
	}

	currentAws, err := awsProvider.GetCallerIdentity(ctx, &awsProvider.GetCallerIdentityArgs{}, nil)
	if err != nil {
		return err
	}
	accountId := currentAws.AccountId

	oidcIssuerUrl := eksCluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer().Elem()
	_, err = iam.NewOpenIdConnectProvider(ctx, "my-oidc-provider", &iam.OpenIdConnectProviderArgs{
		ClientIdLists: pulumi.StringArray{
			pulumi.String("sts.amazonaws.com"),
		},
		Url: oidcIssuerUrl,
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster}))
	if err != nil {
		return err
	}
	oidcIssuerHostPath := oidcIssuerUrl.ApplyT(func(urlStr string) (string, error) {
		parsedUrl, err := url.Parse(urlStr)
		if err != nil {
			return "", err
		}
		// This is the format required for the OIDC provider ARN path segment and condition keys.
		if parsedUrl.Path == "" || parsedUrl.Path == "/" {
			return parsedUrl.Host, nil
		}
		return parsedUrl.Host + parsedUrl.Path, nil
	}).(pulumi.StringOutput)

	err = deployAddons(r, oidcIssuerHostPath, accountId, ctx, eksCluster)
	if err != nil {
		return err
	}

	nodeGroup0, err := eks.NewNodeGroup(ctx, "node-group-0", &eks.NodeGroupArgs{
		ClusterName:   eksCluster.Name,
		NodeGroupName: pulumi.String("eks-nodegroup-0"),
		NodeRoleArn:   nodeGroupRole.Arn,
		SubnetIds:     subnetIds,
		InstanceTypes: pulumi.StringArray(util.ArrayConvert(r.AllocationData.InstanceTypes, func(s string) pulumi.StringInput { return pulumi.String(s) })),
		ScalingConfig: &eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(r.ScalingDesiredSize),
			MaxSize:     pulumi.Int(r.ScalingMaxSize),
			MinSize:     pulumi.Int(r.ScalingMinSize),
		},
		CapacityType: util.If(r.AllocationData != nil && r.AllocationData.SpotPrice != nil,
			pulumi.String("SPOT"),
			pulumi.String("ON_DEMAND")),
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster, nodeGroupRole}))
	if err != nil {
		return err
	}

	// Install AWS Load Balancer Controller
	if r.LoadBalancerController {
		// IAM Policy Document
		err := r.installAwsLoadBalancerController(ctx, oidcIssuerHostPath, accountId, k8sProvider, eksCluster, vpc, nodeGroup0)
		if err != nil {
			return err
		}
	}

	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig), kubeconfig)
	return nil
}

// security group with ingress rules for ssh and vnc
func (r *EKSRequest) securityGroups(ctx *pulumi.Context,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// ingress for ssh access from 0.0.0.0
	var ingressRules []securityGroup.IngressRules
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	ingressRules = []securityGroup.IngressRules{sshIngressRule}
	// Integration ports
	cirrusPort, err := cirrus.CirrusPort()
	if err != nil {
		return nil, err
	}
	if cirrusPort != nil {
		ingressRules = append(ingressRules,
			securityGroup.IngressRules{
				Description: fmt.Sprintf("Cirrus port for %s", awsEKSID),
				FromPort:    *cirrusPort,
				ToPort:      *cirrusPort,
				Protocol:    "tcp",
				CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
			})
	}

	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(r.Prefix, awsEKSID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", awsEKSID),
		IngressRules: ingressRules,
	}.Create(ctx)
	if err != nil {
		return nil, err
	}
	// Convert to an array of IDs
	sgs := util.ArrayConvert([]*ec2.SecurityGroup{sg.SG},
		func(sg *ec2.SecurityGroup) pulumi.StringInput {
			return sg.ID()
		})
	return pulumi.StringArray(sgs[:]), nil
}

func (*EKSRequest) createEksRole(ctx *pulumi.Context) (*iam.Role, error) {
	eksRolePolicyJSON, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": "eks.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	eksRole, err := iam.NewRole(ctx, "eks-iam-eksRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(eksRolePolicyJSON),
	})
	if err != nil {
		return nil, err
	}
	eksPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}
	for i, eksPolicy := range eksPolicies {
		_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("rpa-%d", i), &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(eksPolicy),
			Role:      eksRole.Name,
		})
		if err != nil {
			return nil, err
		}
	}
	return eksRole, nil
}

func (*EKSRequest) createNodeGroupRole(ctx *pulumi.Context) (*iam.Role, error) {
	nodeGroupAssumeRolePolicyJSON, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": "ec2.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	nodeGroupRole, err := iam.NewRole(ctx, "nodegroup-iam-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(nodeGroupAssumeRolePolicyJSON),
	})
	if err != nil {
		return nil, err
	}
	nodeGroupPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
	}
	for i, nodeGroupPolicy := range nodeGroupPolicies {
		_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("ngpa-%d", i), &iam.RolePolicyAttachmentArgs{
			Role:      nodeGroupRole.Name,
			PolicyArn: pulumi.String(nodeGroupPolicy),
		}, pulumi.DependsOn([]pulumi.Resource{nodeGroupRole}))
		if err != nil {
			return nil, err
		}
	}
	return nodeGroupRole, nil
}

func (r *EKSRequest) installAwsLoadBalancerController(ctx *pulumi.Context, oidcIssuerHostPath pulumi.StringOutput, accountId string, k8sProvider *kubernetes.Provider, eksCluster *eks.Cluster, vpc *ec2.Vpc, nodeGroup0 *eks.NodeGroup) error {
	policyDocumentJSON := getAwsLoadBalancerControllerIamPolicy()

	// Create IAM policy
	albControllerPolicyAttachment, err := iam.NewPolicy(ctx, "loadBalancerControllerPolicy", &iam.PolicyArgs{
		Policy: pulumi.String(policyDocumentJSON),
	})
	if err != nil {
		return err
	}

	// Create IAM role
	lbcServiceAccountName := pulumi.String("aws-load-balancer-controller-sa")

	assumeRolePolicyJSON := oidcIssuerHostPath.ApplyT(func(hostPath string) (string, error) {
		policy, err := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountId, hostPath),
					},
					"Action": "sts:AssumeRoleWithWebIdentity",
					"Condition": map[string]interface{}{
						"StringEquals": map[string]interface{}{
							fmt.Sprintf("%s:aud", hostPath): "sts.amazonaws.com",
							fmt.Sprintf("%s:sub", hostPath): fmt.Sprintf("system:serviceaccount:kube-system:%s", lbcServiceAccountName),
						},
					},
				},
			},
		})
		if err != nil {
			return "", err
		}
		return string(policy), nil
	}).(pulumi.StringOutput)

	iamRole, err := iam.NewRole(ctx, "loadBalancerControllerRole", &iam.RoleArgs{
		NamePrefix:       pulumi.String("MaptLBCRole-"),
		AssumeRolePolicy: assumeRolePolicyJSON,
	})
	if err != nil {
		return err
	}

	// Attach policy to role
	_, err = iam.NewRolePolicyAttachment(ctx, "loadBalancerControllerPolicyAttachment", &iam.RolePolicyAttachmentArgs{
		Role:      iamRole.Name,
		PolicyArn: albControllerPolicyAttachment.Arn,
	})
	if err != nil {
		return err
	}

	// Create the Kubernetes service account with the IAM role annotation
	lbcK8sServiceAccount, err := corev1.NewServiceAccount(ctx, "lbcK8sServiceAccount", &corev1.ServiceAccountArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      lbcServiceAccountName,
			Namespace: pulumi.String("kube-system"),
			Annotations: pulumi.StringMap{
				"eks.amazonaws.com/role-arn": iamRole.Arn,
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DeletedWith(eksCluster))
	if err != nil {
		return err
	}

	// Deploy the AWS Load Balancer Controller as a Helm chart
	_, err = helmv3.NewChart(ctx, "aws-load-balancer-controller", helmv3.ChartArgs{
		Chart: pulumi.String("aws-load-balancer-controller"),
		FetchArgs: helmv3.FetchArgs{
			Repo: pulumi.String("https://aws.github.io/eks-charts"),
		},
		Namespace: pulumi.String("kube-system"),
		Values: pulumi.Map{
			"clusterName": eksCluster.Name,
			"serviceAccount": pulumi.Map{
				"create": pulumi.Bool(false),    // Tell Helm chart not to create SA
				"name":   lbcServiceAccountName, // Tell Helm chart to use the SA we created
			},
			"region": pulumi.String(*r.AllocationData.Region),
			"vpcId":  vpc.ID(),
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{eksCluster, nodeGroup0, iamRole, lbcK8sServiceAccount}))
	if err != nil {
		return err
	}
	return nil
}

func deployAddons(r *EKSRequest, oidcIssuerHostPath pulumi.StringOutput, accountId string, ctx *pulumi.Context, eksCluster *eks.Cluster) error {
	for _, addon := range r.Addons {
		if addon == "aws-ebs-csi-driver" {
			// Create the IAM role for the EBS CSI driver
			ebsCsiDriverServiceAccountName := pulumi.String("ebs-csi-controller-sa")

			assumeRolePolicyJSON := oidcIssuerHostPath.ApplyT(func(hostPath string) (string, error) {
				policy, err := json.Marshal(map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []map[string]interface{}{
						{
							"Effect": "Allow",
							"Principal": map[string]interface{}{
								"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountId, hostPath),
							},
							"Action": "sts:AssumeRoleWithWebIdentity",
							"Condition": map[string]interface{}{
								"StringEquals": map[string]interface{}{
									fmt.Sprintf("%s:aud", hostPath): "sts.amazonaws.com",
									fmt.Sprintf("%s:sub", hostPath): fmt.Sprintf("system:serviceaccount:kube-system:%s", ebsCsiDriverServiceAccountName),
								},
							},
						},
					},
				})
				if err != nil {
					return "", err
				}
				return string(policy), nil
			}).(pulumi.StringOutput)

			awsEbsCsiDriverRole, err := iam.NewRole(ctx, "AmazonEKS_EBS_CSI_DriverRole", &iam.RoleArgs{
				NamePrefix:       pulumi.String("MaptEBSCSIDriverRole-"),
				AssumeRolePolicy: assumeRolePolicyJSON,
			})
			if err != nil {
				return err
			}
			_, err = iam.NewRolePolicyAttachment(ctx, "AmazonEBSCSIDriverPolicyAttachment", &iam.RolePolicyAttachmentArgs{
				PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"),
				Role:      awsEbsCsiDriverRole.Name,
			})
			if err != nil {
				return err
			}

			// Enable addon and set as default storage class
			configValues, err := json.Marshal(map[string]interface{}{
				"defaultStorageClass": map[string]interface{}{
					"enabled": true,
				},
			})
			if err != nil {
				return err
			}
			_, err = eks.NewAddon(ctx, addon, &eks.AddonArgs{
				ClusterName:           eksCluster.Name,
				AddonName:             pulumi.String(addon),
				ServiceAccountRoleArn: awsEbsCsiDriverRole.Arn,
				ConfigurationValues:   pulumi.String(configValues),
			}, pulumi.DeletedWith(eksCluster))
			if err != nil {
				return err
			}

		} else {
			_, err := eks.NewAddon(ctx, addon, &eks.AddonArgs{
				ClusterName: eksCluster.Name,
				AddonName:   pulumi.String(addon),
			}, pulumi.DeletedWith(eksCluster))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Create the KubeConfig Structure as per https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
func generateKubeconfig(clusterEndpoint pulumi.StringOutput, certData pulumi.StringOutput, clusterName pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(clusterEndpoint, certData, clusterName).ApplyT(func(args []interface{}) (string, error) {
		endpoint := args[0].(string)
		cert := args[1].(string)
		name := args[2].(string)
		kubeconfigMap := map[string]interface{}{
			"apiVersion": "v1",
			"clusters": []map[string]interface{}{
				{
					"name": "kubernetes",
					"cluster": map[string]interface{}{
						"server":                     endpoint,
						"certificate-authority-data": cert,
					},
				},
			},
			"contexts": []map[string]interface{}{
				{
					"name": "aws",
					"context": map[string]interface{}{
						"cluster": "kubernetes",
						"user":    "aws",
					},
				},
			},
			"current-context": "aws",
			"kind":            "Config",
			"users": []map[string]interface{}{
				{
					"name": "aws",
					"user": map[string]interface{}{
						"exec": map[string]interface{}{
							"apiVersion": "client.authentication.k8s.io/v1beta1",
							"command":    "aws",
							"args": []string{
								"eks",
								"get-token",
								"--cluster-name",
								name,
							},
						},
					},
				},
			},
		}
		kubeconfigJson, err := json.MarshalIndent(kubeconfigMap, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error generating kubeconfig: %w", err)
		}
		return string(kubeconfigJson), nil
	}).(pulumi.StringOutput)
}

// Write exported values in context to files o a selected target folder
func (r *EKSRequest) manageResults(stackResult auto.UpResult) error {

	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig): "kubeconfig",
	})
}

func getAwsLoadBalancerControllerIamPolicy() json.RawMessage {
	// Based on this with a few additions: https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.12.0/docs/install/iam_policy.json
	policyDocumentJSON := json.RawMessage(`{
			"Version": "2012-10-17",
			"Statement": [
					{
							"Effect": "Allow",
							"Action": [
									"iam:CreateServiceLinkedRole"
							],
							"Resource": "*",
							"Condition": {
									"StringEquals": {
											"iam:AWSServiceName": "elasticloadbalancing.amazonaws.com"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:DescribeAccountAttributes",
									"ec2:DescribeAddresses",
									"ec2:DescribeAvailabilityZones",
									"ec2:DescribeInternetGateways",
									"ec2:DescribeVpcs",
									"ec2:DescribeVpcPeeringConnections",
									"ec2:DescribeSubnets",
									"ec2:DescribeSecurityGroups",
									"ec2:DescribeInstances",
									"ec2:DescribeNetworkInterfaces",
									"ec2:DescribeTags",
									"ec2:GetCoipPoolUsage",
									"ec2:DescribeCoipPools",
									"ec2:GetSecurityGroupsForVpc",
									"ec2:DescribeIpamPools",
									"ec2:DescribeRouteTables",
									"elasticloadbalancing:DescribeLoadBalancers",
									"elasticloadbalancing:DescribeLoadBalancerAttributes",
									"elasticloadbalancing:DescribeListeners",
									"elasticloadbalancing:DescribeListenerCertificates",
									"elasticloadbalancing:DescribeSSLPolicies",
									"elasticloadbalancing:DescribeRules",
									"elasticloadbalancing:DescribeTargetGroups",
									"elasticloadbalancing:DescribeTargetGroupAttributes",
									"elasticloadbalancing:DescribeTargetHealth",
									"elasticloadbalancing:DescribeTags",
									"elasticloadbalancing:DescribeTrustStores",
									"elasticloadbalancing:DescribeListenerAttributes",
									"elasticloadbalancing:DescribeCapacityReservation"
							],
							"Resource": "*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"cognito-idp:DescribeUserPoolClient",
									"acm:ListCertificates",
									"acm:DescribeCertificate",
									"iam:ListServerCertificates",
									"iam:GetServerCertificate",
									"waf-regional:GetWebACL",
									"waf-regional:GetWebACLForResource",
									"waf-regional:AssociateWebACL",
									"waf-regional:DisassociateWebACL",
									"wafv2:GetWebACL",
									"wafv2:GetWebACLForResource",
									"wafv2:AssociateWebACL",
									"wafv2:DisassociateWebACL",
									"shield:GetSubscriptionState",
									"shield:DescribeProtection",
									"shield:CreateProtection",
									"shield:DeleteProtection"
							],
							"Resource": "*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:AuthorizeSecurityGroupIngress",
									"ec2:RevokeSecurityGroupIngress"
							],
							"Resource": "*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:CreateSecurityGroup"
							],
							"Resource": "*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:CreateTags"
							],
							"Resource": "arn:aws:ec2:*:*:security-group/*",
							"Condition": {
									"StringEquals": {
											"ec2:CreateAction": "CreateSecurityGroup"
									},
									"Null": {
											"aws:RequestTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:CreateTags",
									"ec2:DeleteTags"
							],
							"Resource": "arn:aws:ec2:*:*:security-group/*",
							"Condition": {
									"Null": {
											"aws:RequestTag/elbv2.k8s.aws/cluster": "true",
											"aws:ResourceTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"ec2:AuthorizeSecurityGroupIngress",
									"ec2:RevokeSecurityGroupIngress",
									"ec2:DeleteSecurityGroup"
							],
							"Resource": "*",
							"Condition": {
									"Null": {
											"aws:ResourceTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:CreateLoadBalancer",
									"elasticloadbalancing:CreateTargetGroup"
							],
							"Resource": "*",
							"Condition": {
									"Null": {
											"aws:RequestTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:CreateListener",
									"elasticloadbalancing:DeleteListener",
									"elasticloadbalancing:CreateRule",
									"elasticloadbalancing:DeleteRule"
							],
							"Resource": "*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:AddTags",
									"elasticloadbalancing:RemoveTags"
							],
							"Resource": [
									"arn:aws:elasticloadbalancing:*:*:targetgroup/*/*",
									"arn:aws:elasticloadbalancing:*:*:loadbalancer/net/*/*",
									"arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*"
							],
							"Condition": {
									"Null": {
											"aws:RequestTag/elbv2.k8s.aws/cluster": "true",
											"aws:ResourceTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:AddTags",
									"elasticloadbalancing:RemoveTags"
							],
							"Resource": [
									"arn:aws:elasticloadbalancing:*:*:listener/net/*/*/*",
									"arn:aws:elasticloadbalancing:*:*:listener/app/*/*/*",
									"arn:aws:elasticloadbalancing:*:*:listener-rule/net/*/*/*",
									"arn:aws:elasticloadbalancing:*:*:listener-rule/app/*/*/*"
							]
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:ModifyLoadBalancerAttributes",
									"elasticloadbalancing:SetIpAddressType",
									"elasticloadbalancing:SetSecurityGroups",
									"elasticloadbalancing:SetSubnets",
									"elasticloadbalancing:DeleteLoadBalancer",
									"elasticloadbalancing:ModifyTargetGroup",
									"elasticloadbalancing:ModifyTargetGroupAttributes",
									"elasticloadbalancing:DeleteTargetGroup",
									"elasticloadbalancing:ModifyListenerAttributes",
									"elasticloadbalancing:ModifyCapacityReservation",
									"elasticloadbalancing:ModifyIpPools"
							],
							"Resource": "*",
							"Condition": {
									"Null": {
											"aws:ResourceTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:AddTags"
							],
							"Resource": [
									"arn:aws:elasticloadbalancing:*:*:targetgroup/*/*",
									"arn:aws:elasticloadbalancing:*:*:loadbalancer/net/*/*",
									"arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*"
							],
							"Condition": {
									"StringEquals": {
											"elasticloadbalancing:CreateAction": [
													"CreateTargetGroup",
													"CreateLoadBalancer"
											]
									},
									"Null": {
											"aws:RequestTag/elbv2.k8s.aws/cluster": "false"
									}
							}
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:RegisterTargets",
									"elasticloadbalancing:DeregisterTargets"
							],
							"Resource": "arn:aws:elasticloadbalancing:*:*:targetgroup/*/*"
					},
					{
							"Effect": "Allow",
							"Action": [
									"elasticloadbalancing:SetWebAcl",
									"elasticloadbalancing:ModifyListener",
									"elasticloadbalancing:AddListenerCertificates",
									"elasticloadbalancing:RemoveListenerCertificates",
									"elasticloadbalancing:ModifyRule",
									"elasticloadbalancing:SetRulePriorities"
							],
							"Resource": "*"
					}
			]
	}`)
	return policyDocumentJSON
}
