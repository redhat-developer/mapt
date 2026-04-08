package eks

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	eksaws "github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	ami "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	eksNodeDiskSize = 200
)

type selfManagedNodeGroupArgs struct {
	prefix            string
	kubernetesVersion string
	arch              string
	eksCluster        *eksaws.Cluster
	nodeGroupRole     *iam.Role
	securityGroups    pulumi.StringArray
	subnetIds         pulumi.StringArray
	instanceTypes     []string
	scalingDesired    int
	scalingMax        int
	scalingMin        int
	spotPrice         *float64
	accessEntry       *eksaws.AccessEntry
	tags              pulumi.StringMap
}

// createSelfManagedNodeGroup creates an ASG-based self-managed EKS node group.
// For spot mode (spotPrice != nil), it sets SpotMaxPrice on the ASG mixed instances policy.
// For on-demand mode, it uses 100% on-demand capacity.
func createSelfManagedNodeGroup(ctx *pulumi.Context, args *selfManagedNodeGroupArgs) (*autoscaling.Group, error) {
	// Look up EKS-optimized AL2023 AMI
	eksAMI, err := ami.GetAMIByName(ctx,
		fmt.Sprintf("amazon-eks-node-al2023-%s-standard-%s-*", args.arch, args.kubernetesVersion),
		[]string{"amazon"},
		map[string]string{"architecture": args.arch})
	if err != nil {
		return nil, fmt.Errorf("failed to look up EKS-optimized AMI: %w", err)
	}

	// Create IAM instance profile from the node role
	instanceProfile, err := iam.NewInstanceProfile(ctx, "eks-node-instance-profile", &iam.InstanceProfileArgs{
		Role: args.nodeGroupRole.Name,
		Tags: args.tags,
	})
	if err != nil {
		return nil, err
	}

	// Build nodeadm userdata using cluster outputs
	userData := generateNodeadmUserData(
		args.eksCluster.Name,
		args.eksCluster.Endpoint,
		args.eksCluster.CertificateAuthority.Data().Elem(),
		args.eksCluster.KubernetesNetworkConfig.ServiceIpv4Cidr().Elem(),
	)

	// Merge custom SGs with the cluster's managed SG
	// ClusterSecurityGroupId() returns StringPtrOutput; .Elem() converts to StringOutput (implements StringInput)
	allSecurityGroups := append(
		args.securityGroups,
		args.eksCluster.VpcConfig.ClusterSecurityGroupId().Elem(),
	)

	// Build cluster name tag for EKS discovery
	clusterTag := args.eksCluster.Name.ApplyT(func(name string) map[string]string {
		return map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", name): "owned",
		}
	}).(pulumi.StringMapOutput)

	// Merge resource tags with the EKS cluster tag
	instanceTags := pulumi.All(args.tags, clusterTag).ApplyT(
		func(all []interface{}) map[string]string {
			merged := make(map[string]string)
			if baseTags, ok := all[0].(map[string]string); ok {
				for k, v := range baseTags {
					merged[k] = v
				}
			}
			if eksTags, ok := all[1].(map[string]string); ok {
				for k, v := range eksTags {
					merged[k] = v
				}
			}
			return merged
		},
	).(pulumi.StringMapOutput)

	// Create launch template
	lt, err := ec2.NewLaunchTemplate(ctx,
		resourcesUtil.GetResourceName(args.prefix, awsEKSID, "lt"),
		&ec2.LaunchTemplateArgs{
			NamePrefix: pulumi.String(awsEKSID),
			ImageId:    pulumi.String(eksAMI.Id),
			IamInstanceProfile: ec2.LaunchTemplateIamInstanceProfileArgs{
				Arn: instanceProfile.Arn,
			},
			UserData: userData,
			NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
				&ec2.LaunchTemplateNetworkInterfaceArgs{
					SecurityGroups:           allSecurityGroups,
					AssociatePublicIpAddress: pulumi.String("true"),
				},
			},
			BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
				&ec2.LaunchTemplateBlockDeviceMappingArgs{
					DeviceName: pulumi.String("/dev/xvda"),
					Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
						VolumeSize: pulumi.Int(eksNodeDiskSize),
					},
				},
			},
			Tags: instanceTags,
			TagSpecifications: ec2.LaunchTemplateTagSpecificationArray{
				&ec2.LaunchTemplateTagSpecificationArgs{
					ResourceType: pulumi.String(constants.PulumiAwsResourceInstance),
					Tags:         instanceTags,
				},
				&ec2.LaunchTemplateTagSpecificationArgs{
					ResourceType: pulumi.String(constants.PulumiAwsResourceVolume),
					Tags:         instanceTags,
				},
				&ec2.LaunchTemplateTagSpecificationArgs{
					ResourceType: pulumi.String(constants.PulumiAwsResourceNetworkInterface),
					Tags:         instanceTags,
				},
			},
		})
	if err != nil {
		return nil, err
	}

	// Build instance type overrides
	overrides := autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArray{}
	for _, instanceType := range args.instanceTypes {
		overrides = append(overrides, &autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
			InstanceType: pulumi.String(instanceType),
		})
	}

	// Build instances distribution (spot vs on-demand)
	distribution := &autoscaling.GroupMixedInstancesPolicyInstancesDistributionArgs{}
	if args.spotPrice != nil {
		spotMaxPrice := strconv.FormatFloat(*args.spotPrice, 'f', -1, 64)
		distribution.OnDemandBaseCapacity = pulumi.Int(0)
		distribution.OnDemandPercentageAboveBaseCapacity = pulumi.Int(0)
		distribution.SpotAllocationStrategy = pulumi.String("capacity-optimized")
		distribution.SpotMaxPrice = pulumi.String(spotMaxPrice)
	} else {
		distribution.OnDemandPercentageAboveBaseCapacity = pulumi.Int(100)
	}

	mixedInstancesPolicy := &autoscaling.GroupMixedInstancesPolicyArgs{
		InstancesDistribution: distribution,
		LaunchTemplate: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateArgs{
			LaunchTemplateSpecification: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationArgs{
				LaunchTemplateId: lt.ID(),
			},
			Overrides: overrides,
		},
	}

	// Build ASG tags: Name + resource tags
	asgName := resourcesUtil.GetResourceName(args.prefix, awsEKSID, "asg")
	asgTags := autoscaling.GroupTagArray{
		&autoscaling.GroupTagArgs{
			Key:               pulumi.String("Name"),
			Value:             pulumi.String(asgName),
			PropagateAtLaunch: pulumi.Bool(true),
		},
	}
	for k, v := range args.tags {
		asgTags = append(asgTags, &autoscaling.GroupTagArgs{
			Key:               pulumi.String(k),
			Value:             v,
			PropagateAtLaunch: pulumi.Bool(true),
		})
	}

	// Dependencies: access entry must exist before nodes try to join
	dependsOn := []pulumi.Resource{args.eksCluster, args.nodeGroupRole, instanceProfile}
	if args.accessEntry != nil {
		dependsOn = append(dependsOn, args.accessEntry)
	}

	return autoscaling.NewGroup(ctx, asgName,
		&autoscaling.GroupArgs{
			DesiredCapacity:        pulumi.Int(args.scalingDesired),
			MaxSize:                pulumi.Int(args.scalingMax),
			MinSize:                pulumi.Int(args.scalingMin),
			VpcZoneIdentifiers:     args.subnetIds,
			MixedInstancesPolicy:   mixedInstancesPolicy,
			Tags:                   asgTags,
			WaitForCapacityTimeout: pulumi.String("15m"),
			HealthCheckType:        pulumi.String("EC2"),
			HealthCheckGracePeriod: pulumi.Int(300),
		},
		pulumi.DependsOn(dependsOn))
}

// generateNodeadmUserData builds base64-encoded MIME userdata for nodeadm bootstrap.
// Cluster name, endpoint, CA, and service CIDR are pulumi.StringOutput values resolved at deploy time.
// The service CIDR (cidr field) is required for AL2023 nodeadm to properly configure
// pod networking without needing to call the DescribeCluster API.
func generateNodeadmUserData(
	clusterName, endpoint, certificateAuthority, serviceCIDR pulumi.StringOutput,
) pulumi.StringPtrInput {
	return pulumi.All(clusterName, endpoint, certificateAuthority, serviceCIDR).ApplyT(
		func(args []interface{}) *string {
			name := args[0].(string)
			ep := args[1].(string)
			ca := args[2].(string)
			cidr := args[3].(string)

			nodeConfig := fmt.Sprintf(`MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="BOUNDARY"

--BOUNDARY
Content-Type: application/node.eks.aws

---
apiVersion: node.eks.aws/v1alpha1
kind: NodeConfig
spec:
  cluster:
    name: %s
    apiServerEndpoint: %s
    certificateAuthority: %s
    cidr: %s

--BOUNDARY--`, name, ep, ca, cidr)

			encoded := base64.StdEncoding.EncodeToString([]byte(nodeConfig))
			return &encoded
		},
	).(pulumi.StringPtrOutput)
}
