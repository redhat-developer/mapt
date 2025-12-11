package compute

// import (
// 	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
// 	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
// 	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
// 	mc "github.com/redhat-developer/mapt/pkg/manager/context"
// )

// type ComputeArgs struct {
// 	MCtx   *mc.Context
// 	Prefix string
// 	ID     string
// 	VPC    *ibmcloud.IsVpc
// 	// Subnet *ec2.Subnet
// 	// Eip    *ec2.Eip
// 	LB *ibmcloud.IsLb
// 	// Array of TCP ports to be
// 	// created as tg for the LB
// 	LBTargetGroups []int
// 	// AMI             *ec2.LookupAmiResult

// 	Key *ibmcloud.PiKey
// 	// SecurityGroups  pulumi.StringArray
// 	// InstaceTypes    []string
// 	// InstanceProfile *iam.InstanceProfile
// 	// DiskSize        *int
// 	// Airgap          bool
// 	// Spot            bool
// 	// Only required if Spot is true
// 	// SpotPrice float64
// 	// Only required if we need to set userdata
// 	UserDataAsBase64 pulumi.StringPtrInput
// 	// If we need to add explicit dependecies
// 	DependsOn []pulumi.Resource
// }

// func Create() {
// 	sshKey, err := ibmcloud.NewPiKey(ctx, "power11-ssh-key", &ibmcloud.PiKeyArgs{
// 		PiKeyName:         pulumi.String("power11-access-key"),
// 		PiSshKey:          pulumi.String(sshPublicKey),
// 		PiCloudInstanceId: powerVsWorkspace.Guid,
// 	})

// 	power11Instance, err := ibmcloud.NewPiInstance(ctx,
// 		"power11-instance",
// 		&ibmcloud.PiInstanceArgs{
// 			PiInstanceName:    pulumi.String(instanceName),
// 			PiCloudInstanceId: powerVsWorkspace.Guid,

// 			// Power11 Specifications
// 			PiMemory:     pulumi.Float64(16),         // 16 GB RAM
// 			PiProcessors: pulumi.Float64(2),          // 2 cores
// 			PiProcType:   pulumi.String("dedicated"), // dedicated or shared
// 			PiSysType:    pulumi.String("s1022"),     // Power10/11 system type

// 			// Operating System - RHEL 9 for Power11
// 			PiImageId: pulumi.String("rhel-9-2"), // Replace with actual image ID

// 			// Network Configuration
// 			PiNetworks: ibmcloud.PiInstancePiNetworkArray{
// 				&ibmcloud.PiInstancePiNetworkArgs{
// 					NetworkId: powerVsNetwork.NetworkId,
// 				},
// 			},

// 			// SSH Access
// 			PiKeyPairName: sshKey.PiKeyName,

// 			// Storage - tier1 (NVMe) for best performance
// 			PiStorageType: pulumi.String("tier1"),
// 			PiStoragePool: pulumi.String("Tier1-Flash"),

// 			// Health
// 			PiHealthStatus: pulumi.String("OK"),
// 		}, pulumi.Timeouts(&pulumi.CustomTimeouts{
// 			Create: "30m",
// 			Update: "30m",
// 			Delete: "30m",
// 		}))
// 	if err != nil {
// 		return err
// 	}
// }

// func s390x(ctx *pulumi.Context, vpc *ibmcloud.IsVpc) (*ibmcloud.IsInstance, error) {
// 	instanceS390x, err := ibmcloud.NewIsInstance(ctx,
// 		"gitlab-runner-s390x",
// 		&ibmcloud.IsInstanceArgs{
// 			Name: pulumi.String("gitlab-runner-s390x"),
// 			// Image:         imageS390x,
// 			Profile:       pulumi.String("bz2-2x8"), // s390x profile: 2 vCPU, 8 GB RAM
// 			Vpc:           vpc.ID(),
// 			Zone:          pulumi.String("us-east-1"),
// 			ResourceGroup: rg.ID(),
// 			Keys:          pulumi.StringArray{sshKey.ID()},
// 			PrimaryNetworkInterface: &ibmcloud.IsInstancePrimaryNetworkInterfaceArgs{
// 				Subnet: subnetS390x.ID(),
// 				SecurityGroups: pulumi.StringArray{
// 					securityGroup.ID(),
// 				},
// 			},
// 			UserData: pulumi.String(fedoraUserData),
// 			Tags: pulumi.StringArray{
// 				pulumi.String("gitlab-runner"),
// 				pulumi.String("s390x"),
// 				pulumi.String("ibm-z"),
// 			},
// 		})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func createForwardTargetGRoups(ctx *pulumi.Context, port int) (*lb.TargetGroup, error) {
// 	// tg, err := lb.NewTargetGroup(ctx,
// 	// 	resourcesUtil.GetResourceName(r.Prefix, r.ID, fmt.Sprintf("tg-%d", port)),
// 	// 	&lb.TargetGroupArgs{
// 	// 		Port:     pulumi.Int(port),
// 	// 		Protocol: pulumi.String("TCP"),
// 	// 		VpcId:    r.VPC.ID(),
// 	// 	})
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// if _, err := lb.NewListener(ctx,
// 	// 	resourcesUtil.GetResourceName(r.Prefix, r.ID, fmt.Sprintf("listener-%d", port)),
// 	// 	&lb.ListenerArgs{
// 	// 		LoadBalancerArn: r.LB.Arn,
// 	// 		Port:            pulumi.Int(port),
// 	// 		Protocol:        pulumi.String("TCP"),
// 	// 		DefaultActions: lb.ListenerDefaultActionArray{
// 	// 			&lb.ListenerDefaultActionArgs{
// 	// 				Type:           pulumi.String("forward"),
// 	// 				TargetGroupArn: tg.Arn,
// 	// 			},
// 	// 		},
// 	// 	}); err != nil {
// 	// 	return nil, err
// 	// }
// 	// return tg, nil
// 	_, err = ibmcloud.NewIsLbPoolMember(ctx,
// 		"power11-member",
// 		&ibmcloud.IsLbPoolMemberArgs{
// 			Lb:            loadBalancer.ID(),
// 			Pool:          backendPool.ID(),
// 			Port:          pulumi.Int(80),
// 			TargetAddress: power11Instance.PiNetworks.Index(pulumi.Int(0)).IpAddress(),
// 			Weight:        pulumi.Int(100),
// 		})
// 	if err != nil {
// 		return err
// 	}

// 	// Step 13: Create HTTP Listener
// 	_, err = ibmcloud.NewIsLbListener(ctx, "http-listener", &ibmcloud.IsLbListenerArgs{
// 		Lb:              loadBalancer.ID(),
// 		DefaultPool:     backendPool.ID(),
// 		Port:            pulumi.Int(80),
// 		Protocol:        pulumi.String("http"),
// 		ConnectionLimit: pulumi.Int(2000),
// 	})
// 	if err != nil {
// 		return err
// 	}
// }
