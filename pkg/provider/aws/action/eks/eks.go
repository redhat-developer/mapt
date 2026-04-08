package eks

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-playground/validator/v10"
	awsProvider "github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumix"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	network "github.com/redhat-developer/mapt/pkg/provider/aws/modules/network/standard"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	subnet "github.com/redhat-developer/mapt/pkg/provider/aws/services/vpc/subnet"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type EKSArgs struct {
	Prefix                 string
	KubernetesVersion      string
	ComputeRequest         *cr.ComputeRequestArgs
	ScalingDesiredSize     int
	ScalingMaxSize         int
	ScalingMinSize         int
	Spot                   *spotTypes.SpotArgs
	Addons                 []string
	LoadBalancerController bool
	ExcludedZoneIDs        []string
	ServiceEndpoints       []string
}

type eksRequest struct {
	mCtx                   *mc.Context `validate:"required"`
	prefix                 *string
	kubernetesVersion      *string
	arch                   *string
	scalingDesiredSize     *int
	scalingMaxSize         *int
	scalingMinSize         *int
	spot                   bool
	addons                 []string
	loadBalancerController *bool
	allocationData         *allocation.AllocationResult
	availabilityZones      []string
	excludedZoneIDs        []string
	serviceEndpoints       []string
}

func (r *eksRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

func Create(mCtxArgs *mc.ContextArgs, args *EKSArgs) (err error) {
	logging.Debug("Creating EKS")
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	arch := archFromComputeRequest(args.ComputeRequest.Arch)
	r := eksRequest{
		mCtx:                   mCtx,
		prefix:                 &prefix,
		kubernetesVersion:      &args.KubernetesVersion,
		arch:                   &arch,
		scalingDesiredSize:     &args.ScalingDesiredSize,
		scalingMaxSize:         &args.ScalingMaxSize,
		scalingMinSize:         &args.ScalingMinSize,
		loadBalancerController: &args.LoadBalancerController,
		addons:                 args.Addons,
		excludedZoneIDs:        args.ExcludedZoneIDs,
		serviceEndpoints:       args.ServiceEndpoints,
	}
	if args.Spot != nil {
		r.spot = args.Spot.Spot
	}
	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			Prefix:                &args.Prefix,
			ComputeRequest:        args.ComputeRequest,
			AMIProductDescription: &amiProduct,
			Spot:                  args.Spot,
		})
	if err != nil {
		return err
	}
	r.availabilityZones = r.getAvailabilityZonesForEKS(*r.allocationData.Region, r.excludedZoneIDs)
	cs := manager.Stack{
		StackName:   mCtx.StackNameByProject(stackName),
		ProjectName: mCtx.ProjectName(),
		BackedURL:   mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.Region,
			}),
		DeployFunc: r.deployer,
	}
	sr, err := manager.UpStack(mCtx, cs)
	if err != nil {
		return err
	}
	return r.manageResults(mCtx, sr)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Destroy EKS")
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	if err := aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			BackedURL: mCtx.BackedURL(),
			Stackname: stackName,
		}); err != nil {
		return err
	}

	// Cleanup S3 state after all stacks have been destroyed
	return aws.CleanupState(mCtx)
}

// Main function to deploy all requried resources to AWS
func (r *eksRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Networking
	nr, err := network.NetworkRequest{
		MCtx:               r.mCtx,
		Name:               resourcesUtil.GetResourceName(*r.prefix, awsEKSID, "net"),
		CIDR:               network.DefaultCIDRNetwork,
		AvailabilityZones:  r.availabilityZones,
		PublicSubnetsCIDRs: network.GeneratePublicSubnetCIDRs(len(r.availabilityZones)),
		Region:             *r.allocationData.Region,
		NatGatewayMode:     &network.NatGatewayModeNone,
		MapPublicIp:        true,
		ServiceEndpoints:   r.serviceEndpoints,
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
		AccessConfig: &eks.ClusterAccessConfigArgs{
			AuthenticationMode:                      pulumi.String("API_AND_CONFIG_MAP"),
			BootstrapClusterCreatorAdminPermissions: pulumi.Bool(true),
		},
		VpcConfig: &eks.ClusterVpcConfigArgs{
			PublicAccessCidrs: pulumi.StringArray{
				pulumi.String("0.0.0.0/0"),
			},
			SecurityGroupIds: securityGroups,
			SubnetIds:        subnetIds,
		},
		Version: pulumi.String(*r.kubernetesVersion),
		Tags:    r.mCtx.ResourceTags(),
	}, pulumi.DependsOn([]pulumi.Resource{eksRole}))
	if err != nil {
		return err
	}

	kubeconfig := generateKubeconfig(eksCluster.Endpoint, eksCluster.CertificateAuthority.Data().Elem(), eksCluster.Name)

	// Get cluster auth token using the AWS SDK (avoids dependency on the aws CLI binary)
	clusterAuth := eks.GetClusterAuthOutput(ctx, eks.GetClusterAuthOutputArgs{
		Name: eksCluster.Name,
	})
	// Generate a token-based kubeconfig for the Pulumi k8s provider
	providerKubeconfig := generateTokenKubeconfig(eksCluster.Endpoint, eksCluster.CertificateAuthority.Data().Elem(), clusterAuth.Token())

	// Create a Kubernetes provider instance using token-based auth
	k8sProvider, err := kubernetes.NewProvider(ctx, "k8sProvider", &kubernetes.ProviderArgs{
		Kubeconfig: providerKubeconfig,
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
		Url:  oidcIssuerUrl,
		Tags: r.mCtx.ResourceTags(),
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

	// Create access entry for self-managed nodes
	accessEntry, err := eks.NewAccessEntry(ctx, "node-access", &eks.AccessEntryArgs{
		ClusterName:  eksCluster.Name,
		PrincipalArn: nodeGroupRole.Arn,
		Type:         pulumi.String("EC2_LINUX"),
		Tags:         r.mCtx.ResourceTags(),
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster, nodeGroupRole}))
	if err != nil {
		return err
	}

	// Create self-managed node group
	nodeGroup0, err := createSelfManagedNodeGroup(ctx, &selfManagedNodeGroupArgs{
		prefix:            *r.prefix,
		kubernetesVersion: *r.kubernetesVersion,
		arch:              *r.arch,
		eksCluster:        eksCluster,
		nodeGroupRole:     nodeGroupRole,
		securityGroups:    securityGroups,
		subnetIds:         subnetIds,
		instanceTypes:     r.allocationData.InstanceTypes,
		scalingDesired:    *r.scalingDesiredSize,
		scalingMax:        *r.scalingMaxSize,
		scalingMin:        *r.scalingMinSize,
		spotPrice:         r.allocationData.SpotPrice,
		accessEntry:       accessEntry,
		tags:              r.mCtx.ResourceTags(),
	})
	if err != nil {
		return err
	}

	// Deploy addons after node group so nodes are available for addon pods to schedule
	corednsAddon, err := deployAddons(r, oidcIssuerHostPath, accountId, ctx, eksCluster, nodeGroup0)
	if err != nil {
		return err
	}

	// Install AWS Load Balancer Controller (depends on coredns for DNS resolution)
	if *r.loadBalancerController {
		var lbcDeps []pulumi.Resource
		if corednsAddon != nil {
			lbcDeps = append(lbcDeps, corednsAddon)
		}
		err := r.installAwsLoadBalancerController(ctx, oidcIssuerHostPath, accountId, k8sProvider, eksCluster, vpc, nodeGroup0, lbcDeps)
		if err != nil {
			return err
		}
	}

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig), kubeconfig)
	return nil
}

// getAvailabilityZonesForEKS gets availability zones for EKS with default excluded zone IDs
func (r *eksRequest) getAvailabilityZonesForEKS(region string, excludedZoneIDs []string) []string {
	logging.Debugf("Getting availability zones for region region: %s, excludedZoneIDs: %v", *r.allocationData.Region, excludedZoneIDs)
	if len(excludedZoneIDs) == 0 {
		// When no excluded zone IDs are specified, apply default EKS-incompatible availability zone IDs
		// These zone IDs are known to be unsupported by EKS as documented at https://repost.aws/knowledge-center/eks-cluster-creation-errors
		excludedZoneIDs = []string{"use1-az3", "usw1-az2", "cac1-az3"}
	}
	azs := data.GetAvailabilityZones(r.mCtx.Context(), *r.allocationData.Region, excludedZoneIDs)
	logging.Debugf("Got availability zones: %v", azs)
	return azs
}

// security group with ingress rules for ssh and vnc
func (r *eksRequest) securityGroups(ctx *pulumi.Context,
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
		Name:         resourcesUtil.GetResourceName(*r.prefix, awsEKSID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", awsEKSID),
		IngressRules: ingressRules,
	}.Create(ctx, r.mCtx)
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

func (r *eksRequest) createEksRole(ctx *pulumi.Context) (*iam.Role, error) {
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
		Tags:             r.mCtx.ResourceTags(),
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

func (r *eksRequest) createNodeGroupRole(ctx *pulumi.Context) (*iam.Role, error) {
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
		Tags:             r.mCtx.ResourceTags(),
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

func (r *eksRequest) installAwsLoadBalancerController(ctx *pulumi.Context, oidcIssuerHostPath pulumi.StringOutput, accountId string, k8sProvider *kubernetes.Provider, eksCluster *eks.Cluster, vpc *ec2.Vpc, nodeGroup0 pulumi.Resource, extraDeps []pulumi.Resource) error {
	policyDocumentJSON := getAwsLoadBalancerControllerIamPolicy()

	// Create IAM policy
	albControllerPolicyAttachment, err := iam.NewPolicy(ctx, "loadBalancerControllerPolicy", &iam.PolicyArgs{
		Policy: pulumi.String(policyDocumentJSON),
		Tags:   r.mCtx.ResourceTags(),
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
		Tags:             r.mCtx.ResourceTags(),
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
			"region": pulumi.String(*r.allocationData.Region),
			"vpcId":  vpc.ID(),
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn(append([]pulumi.Resource{eksCluster, nodeGroup0, iamRole, lbcK8sServiceAccount}, extraDeps...)))
	if err != nil {
		return err
	}
	return nil
}

// deployAddons deploys EKS addons in the correct dependency order:
//   - Phase 1 (infrastructure): vpc-cni, kube-proxy, eks-pod-identity-agent
//     These are DaemonSet-based addons that provide basic node networking and need only nodes.
//   - Phase 2 (core services): coredns
//     Requires vpc-cni to be running so pods can get IP addresses assigned.
//   - Phase 3 (storage/other): aws-ebs-csi-driver and any remaining addons
//     Requires coredns for DNS resolution of AWS API endpoints.
//
// It returns the coredns addon resource (if deployed) so the LB controller can depend on it.
func deployAddons(r *eksRequest, oidcIssuerHostPath pulumi.StringOutput, accountId string, ctx *pulumi.Context, eksCluster *eks.Cluster, nodeGroup pulumi.Resource) (*eks.Addon, error) {
	// Classify addons into deployment phases
	infraAddons := []string{}     // Phase 1: DaemonSet-based networking/identity addons
	var hasCoredns bool           // Phase 2: coredns
	remainingAddons := []string{} // Phase 3: everything else (aws-ebs-csi-driver, etc.)

	for _, addon := range r.addons {
		switch addon {
		case "vpc-cni", "kube-proxy", "eks-pod-identity-agent":
			infraAddons = append(infraAddons, addon)
		case "coredns":
			hasCoredns = true
		default:
			remainingAddons = append(remainingAddons, addon)
		}
	}

	// Phase 1: Deploy infrastructure addons (depend only on nodeGroup)
	infraAddonResources := []pulumi.Resource{}
	for _, addon := range infraAddons {
		a, err := eks.NewAddon(ctx, addon, &eks.AddonArgs{
			ClusterName:              eksCluster.Name,
			AddonName:                pulumi.String(addon),
			ResolveConflictsOnCreate: pulumi.String("OVERWRITE"),
			Tags:                     r.mCtx.ResourceTags(),
		}, pulumi.DeletedWith(eksCluster),
			pulumi.DependsOn([]pulumi.Resource{nodeGroup}),
			pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30m"}))
		if err != nil {
			return nil, err
		}
		infraAddonResources = append(infraAddonResources, a)
	}

	// Phase 2: Deploy coredns (depends on infrastructure addons being ready)
	var corednsAddon *eks.Addon
	if hasCoredns {
		corednsDeps := append([]pulumi.Resource{nodeGroup}, infraAddonResources...)
		var err error
		corednsAddon, err = eks.NewAddon(ctx, "coredns", &eks.AddonArgs{
			ClusterName:              eksCluster.Name,
			AddonName:                pulumi.String("coredns"),
			ResolveConflictsOnCreate: pulumi.String("OVERWRITE"),
			Tags:                     r.mCtx.ResourceTags(),
		}, pulumi.DeletedWith(eksCluster),
			pulumi.DependsOn(corednsDeps),
			pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30m"}))
		if err != nil {
			return nil, err
		}
	}

	// Phase 3: Deploy remaining addons (depend on coredns for DNS resolution)
	phase3Deps := append([]pulumi.Resource{nodeGroup}, infraAddonResources...)
	if corednsAddon != nil {
		phase3Deps = append(phase3Deps, corednsAddon)
	}
	for _, addon := range remainingAddons {
		if addon == "aws-ebs-csi-driver" {
			if err := deployEbsCsiDriver(ctx, r.mCtx, addon, oidcIssuerHostPath, accountId, eksCluster, phase3Deps); err != nil {
				return nil, err
			}
		} else {
			_, err := eks.NewAddon(ctx, addon, &eks.AddonArgs{
				ClusterName:              eksCluster.Name,
				AddonName:                pulumi.String(addon),
				ResolveConflictsOnCreate: pulumi.String("OVERWRITE"),
				Tags:                     r.mCtx.ResourceTags(),
			}, pulumi.DeletedWith(eksCluster),
				pulumi.DependsOn(phase3Deps),
				pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30m"}))
			if err != nil {
				return nil, err
			}
		}
	}
	return corednsAddon, nil
}

// deployEbsCsiDriver handles the special-case deployment of the aws-ebs-csi-driver addon
// which requires an IRSA role for accessing EBS APIs.
func deployEbsCsiDriver(ctx *pulumi.Context, mCtx *mc.Context, addon string, oidcIssuerHostPath pulumi.StringOutput, accountId string, eksCluster *eks.Cluster, deps []pulumi.Resource) error {
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
		Tags:             mCtx.ResourceTags(),
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

	configValues, err := json.Marshal(map[string]interface{}{
		"defaultStorageClass": map[string]interface{}{
			"enabled": true,
		},
	})
	if err != nil {
		return err
	}
	_, err = eks.NewAddon(ctx, addon, &eks.AddonArgs{
		ClusterName:              eksCluster.Name,
		AddonName:                pulumi.String(addon),
		ServiceAccountRoleArn:    awsEbsCsiDriverRole.Arn,
		ConfigurationValues:      pulumi.String(configValues),
		ResolveConflictsOnCreate: pulumi.String("OVERWRITE"),
		Tags:                     mCtx.ResourceTags(),
	}, pulumi.DeletedWith(eksCluster),
		pulumi.DependsOn(deps),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30m"}))
	return err
}

// buildKubeconfig creates a kubeconfig JSON string with the given endpoint, cert, and user auth config.
func buildKubeconfig(endpoint, cert string, userAuth map[string]interface{}) (string, error) {
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
				"user": userAuth,
			},
		},
	}
	kubeconfigJson, err := json.MarshalIndent(kubeconfigMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error generating kubeconfig: %w", err)
	}
	return string(kubeconfigJson), nil
}

// generateTokenKubeconfig creates a kubeconfig with an embedded token for use by the Pulumi k8s provider.
func generateTokenKubeconfig(clusterEndpoint, certData, token pulumi.StringOutput) pulumi.StringOutput {
	return pulumix.Cast[pulumi.StringOutput](
		pulumix.Apply3Err(clusterEndpoint, certData, token,
			func(endpoint, cert, t string) (string, error) {
				return buildKubeconfig(endpoint, cert, map[string]interface{}{
					"token": t,
				})
			}))
}

// Create the KubeConfig Structure as per https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
func generateKubeconfig(clusterEndpoint, certData, clusterName pulumi.StringOutput) pulumi.StringOutput {
	return pulumix.Cast[pulumi.StringOutput](
		pulumix.Apply3Err(clusterEndpoint, certData, clusterName,
			func(endpoint, cert, name string) (string, error) {
				return buildKubeconfig(endpoint, cert, map[string]interface{}{
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
				})
			}))
}

// Write exported values in context to files o a selected target folder
func (r *eksRequest) manageResults(c *mc.Context, stackResult auto.UpResult) error {

	return output.Write(stackResult, c.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig): "kubeconfig",
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

// archFromComputeRequest converts the compute-request Arch enum to
// the AWS architecture string used in AMI names and filters.
func archFromComputeRequest(a cr.Arch) string {
	if a == cr.Arm64 {
		return "arm64"
	}
	return "x86_64"
}
