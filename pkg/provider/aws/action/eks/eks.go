package eks

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	network "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	subnet "github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/subnet"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type EKSRequest struct {
	Prefix                 string
	Region                 string
	VMSize                 string
	KubernetesVersion      string
	ScalingDesiredSize     int
	ScalingMaxSize         int
	ScalingMinSize         int
	Addons                 []string
	LoadBalancerController bool
	AvailabilityZones      []string
}

func Create(ctx *maptContext.ContextArgs, r *EKSRequest) (err error) {
	logging.Debug("Creating EKS")
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackName),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	r.Region = os.Getenv("AWS_DEFAULT_REGION")
	az := data.GetAvailabilityZones(r.Region)
	r.AvailabilityZones = az
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
		PublicSubnetsCIDRs: network.DefaultCIDRPublicSubnets[:],
		Region:             r.Region,
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

	eksRole, err := iam.NewRole(ctx, "eks-iam-eksRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2008-10-17",
			"Statement": [{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
							"Service": "eks.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
			}]
	}`),
	})
	if err != nil {
		return err
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
			return err
		}
	}
	// Create the EC2 NodeGroup Role
	nodeGroupRole, err := iam.NewRole(ctx, "nodegroup-iam-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
							"Service": "ec2.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
			}]
	}`),
	})
	if err != nil {
		return err
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
		})
		if err != nil {
			return err
		}
	}

	// Security groups
	securityGroups, err := r.securityGroups(ctx, vpc)
	if err != nil {
		return err
	}
	// Create EKS Cluster
	eksCluster, err := eks.NewCluster(ctx, "eks-cluster", &eks.ClusterArgs{
		RoleArn: pulumi.StringInput(eksRole.Arn),
		VpcConfig: &eks.ClusterVpcConfigArgs{
			PublicAccessCidrs: pulumi.StringArray{
				pulumi.String("0.0.0.0/0"),
			},
			SecurityGroupIds: securityGroups,
			SubnetIds:        subnetIds,
		},
		Version: pulumi.String(r.KubernetesVersion),
	})
	if err != nil {
		return err
	}

	for _, addon := range r.Addons {
		_, err = eks.NewAddon(ctx, addon, &eks.AddonArgs{
			ClusterName: eksCluster.Name,
			AddonName:   pulumi.String(addon),
		})
		if err != nil {
			return err
		}
	}

	_, err = eks.NewNodeGroup(ctx, "node-group-0", &eks.NodeGroupArgs{
		ClusterName:   eksCluster.Name,
		NodeGroupName: pulumi.String("eks-nodegroup-0"),
		NodeRoleArn:   pulumi.StringInput(nodeGroupRole.Arn),
		SubnetIds:     subnetIds,
		InstanceTypes: pulumi.StringArray{
			pulumi.String(r.VMSize),
		},
		ScalingConfig: &eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(r.ScalingDesiredSize),
			MaxSize:     pulumi.Int(r.ScalingMaxSize),
			MinSize:     pulumi.Int(r.ScalingMinSize),
		},
	})
	if err != nil {
		return err
	}

	kubeconfig := generateKubeconfig(eksCluster.Endpoint, eksCluster.CertificateAuthority.Data().Elem(), eksCluster.Name)

	if r.LoadBalancerController {
		// Create a Kubernetes provider instance
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8sProvider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeconfig,
		})
		if err != nil {
			return err
		}

		// IAM Policy Document
		policyDocumentJSON := getAwsLoadBalancerControllerIamPolicy()

		// Create IAM policy
		policy, err := iam.NewPolicy(ctx, "loadBalancerControllerPolicy", &iam.PolicyArgs{
			Policy: pulumi.String(policyDocumentJSON),
		})
		if err != nil {
			return err
		}

		// Create IAM role
		role, err := iam.NewRole(ctx, "loadBalancerControllerRole", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2008-10-17",
				"Statement": [{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
								"Service": "ec2.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
				}]
		}`),
		})
		if err != nil {
			return err
		}

		// Attach policy to role
		_, err = iam.NewRolePolicyAttachment(ctx, "loadBalancerControllerPolicyAttachment", &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: policy.Arn,
		})
		if err != nil {
			return err
		}

		loadBalancerControllerNamespace := pulumi.String("kube-system")
		serviceAccountName := pulumi.String("aws-load-balancer-controller")

		// Create the Kubernetes service account with the IAM role annotation
		_, err = corev1.NewServiceAccount(ctx, "k8sServiceAccount", &corev1.ServiceAccountArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      serviceAccountName,
				Namespace: loadBalancerControllerNamespace,
				Annotations: pulumi.StringMap{
					"eks.amazonaws.com/role-arn": role.Arn,
				},
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		_, err = rbacv1.NewClusterRoleBinding(ctx, "aws-lb-controller-rolebinding", &rbacv1.ClusterRoleBindingArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.Sprintf("aws-load-balancer-controller-rolebinding-%s", eksCluster.Name),
			},
			RoleRef: &rbacv1.RoleRefArgs{
				ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
				Kind:     pulumi.String("ClusterRole"),
				Name:     pulumi.String("cluster-admin"),
			},
			Subjects: rbacv1.SubjectArray{
				&rbacv1.SubjectArgs{
					Kind:      pulumi.String("ServiceAccount"),
					Name:      serviceAccountName,
					Namespace: loadBalancerControllerNamespace,
				},
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		// Deploy the AWS Load Balancer Controller as a Helm chart
		_, err = helmv3.NewChart(ctx, "aws-load-balancer-controller", helmv3.ChartArgs{
			Chart: pulumi.String("aws-load-balancer-controller"),
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String("https://aws.github.io/eks-charts"),
			},
			Namespace: loadBalancerControllerNamespace,
			Values: pulumi.Map{
				"clusterName": eksCluster.Name,
				"serviceAccount": pulumi.Map{
					"create": pulumi.Bool(false),
					"name":   serviceAccountName,
				},
				"region": pulumi.String(r.Region),
				"vpcId":  vpc.ID(),
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}
	}

	// ctx.Export("clusterName", eksCluster.Name)
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig), kubeconfig)
	return nil
}

// Create the KubeConfig Structure as per https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
func generateKubeconfig(clusterEndpoint pulumi.StringOutput, certData pulumi.StringOutput, clusterName pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.Sprintf(`{
        "apiVersion": "v1",
        "clusters": [{
            "cluster": {
                "server": "%s",
                "certificate-authority-data": "%s"
            },
            "name": "kubernetes",
        }],
        "contexts": [{
            "context": {
                "cluster": "kubernetes",
                "user": "aws",
            },
            "name": "aws",
        }],
        "current-context": "aws",
        "kind": "Config",
        "users": [{
            "name": "aws",
            "user": {
                "exec": {
                    "apiVersion": "client.authentication.k8s.io/v1beta1",
                    "command": "aws",
                    "args": [
                        "eks",
                        "get-token",
                        "--cluster-name",
                        "%s",
                    ],
                },
            },
        }],
    }`, clusterEndpoint, certData, clusterName)
}

// Write exported values in context to files o a selected target folder
func (r *EKSRequest) manageResults(stackResult auto.UpResult) error {

	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig): "kubeconfig",
	})
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

func getAwsLoadBalancerControllerIamPolicy() json.RawMessage {
	// Source: https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.12.0/docs/install/iam_policy.json
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
