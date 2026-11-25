package aks

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v3"

	containerservice "github.com/pulumi/pulumi-azure-native-sdk/containerservice/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/managedidentity/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type AKSArgs struct {
	Prefix   string
	Location string
	// VMSizes are the VM sizes the user pool can run on. When spot is enabled
	// the best option across the sizes is selected, otherwise the first one is
	// used.
	VMSizes          []string
	SystemPoolVMSize string
	WorkersDesired   int
	// "1.26.3"
	KubernetesVersion string
	OnlySystemPool    bool
	EnableAppRouting  bool
	Spot              *spotTypes.SpotArgs
}

type aksRequest struct {
	mCtx             *mc.Context `validate:"required"`
	prefix           *string
	location         *string
	vmSizes          []string
	systemPoolVMSize *string
	workersDesired   *int
	// "1.26.3"
	kubernetesVersion *string
	onlySystemPool    *bool
	enableAppRouting  *bool
	spot              *spotTypes.SpotArgs
}

func (r *aksRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	if err := v.Struct(r); err != nil {
		return err
	}
	if r.systemPoolVMSize == nil || *r.systemPoolVMSize == "" {
		return fmt.Errorf("the system pool vmsize is required")
	}
	if !*r.onlySystemPool {
		if len(r.vmSizes) == 0 {
			return fmt.Errorf("at least one user pool vmsize is required")
		}
		if r.workersDesired == nil || *r.workersDesired < 1 {
			return fmt.Errorf("the workers desired must be at least 1")
		}
	}
	return nil
}

func Create(mCtxArgs *mc.ContextArgs, args *AKSArgs) (err error) {
	logging.Debug("Creating AKS")
	// Create mapt Context. The location always comes from the location flag, as
	// the spot search and the default flag value already resolve it, so no
	// default location has to be set in the environment.
	mCtx, err := mc.Init(mCtxArgs, azure.ProviderOptionalLocation())
	if err != nil {
		return err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	vmSizes := []string{}
	for _, vmSize := range args.VMSizes {
		if vmSize = strings.TrimSpace(vmSize); vmSize != "" {
			vmSizes = append(vmSizes, vmSize)
		}
	}
	r := &aksRequest{
		mCtx:              mCtx,
		prefix:            &prefix,
		location:          &args.Location,
		vmSizes:           vmSizes,
		systemPoolVMSize:  &args.SystemPoolVMSize,
		workersDesired:    &args.WorkersDesired,
		kubernetesVersion: &args.KubernetesVersion,
		onlySystemPool:    &args.OnlySystemPool,
		enableAppRouting:  &args.EnableAppRouting,
		spot:              args.Spot,
	}
	if err := r.validate(); err != nil {
		return err
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAKS),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, err := manager.UpStack(mCtx, cs)
	if err != nil {
		return err
	}
	return r.manageResults(sr)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Destroy AKS")
	mCtx, err := mc.Init(mCtxArgs, azure.ProviderOptionalLocation())
	if err != nil {
		return err
	}
	if err := azure.DestroyStack(mCtx, azure.DestroyStackRequest{
		Stackname: stackAKS,
	}); err != nil {
		return err
	}
	return azure.CleanupState(mCtx)
}

// Main function to deploy all requried resources to azure
func (r *aksRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Resolve the location and the VM size of the user pool
	selection, err := r.userPoolSelection()
	if err != nil {
		return err
	}
	// Get location for creating Resouce Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(selection.location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureAKSID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(r.mCtx.RunID()),
			Tags:              r.mCtx.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	// We will control networking in the future but we need to extend the network module to accept
	// count on SN and types as all NodePools should be on dif SN from the same VN
	// nr := network.NetworkRequest{
	// 	Prefix:        r.Prefix,
	// 	ComponentID:   azureAKSID,
	// 	ResourceGroup: rg,
	// }
	// _, err = nr.Create(ctx)
	// if err != nil {
	// 	return err
	// }

	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureAKSID, "privatekey"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}

	// create a user assigned identity to use for the cluster
	identity, err := managedidentity.NewUserAssignedIdentity(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureAKSID, "identity"),
		&managedidentity.UserAssignedIdentityArgs{
			Location:          pulumi.String(selection.location),
			ResourceGroupName: rg.Name,
			Tags:              r.mCtx.ResourceTags(),
		})

	if err != nil {
		return err
	}
	// create the cluster
	managedClusterArgs := &containerservice.ManagedClusterArgs{
		ResourceGroupName: rg.Name,
		Location:          pulumi.String(selection.location),
		Identity: &containerservice.ManagedClusterIdentityArgs{
			Type: containerservice.ResourceIdentityTypeUserAssigned,
			UserAssignedIdentities: pulumi.StringArray{
				identity.ID(),
			},
		},
		KubernetesVersion: pulumi.String(*r.kubernetesVersion),
		DnsPrefix:         pulumi.String("mapt"),
		EnableRBAC:        pulumi.Bool(true),
		AgentPoolProfiles: agentPoolProfiles(r.agentPools(selection)),
		LinuxProfile: &containerservice.ContainerServiceLinuxProfileArgs{
			AdminUsername: pulumi.String("aksuser"),
			Ssh: &containerservice.ContainerServiceSshConfigurationArgs{
				PublicKeys: containerservice.ContainerServiceSshPublicKeyArray{
					&containerservice.ContainerServiceSshPublicKeyArgs{
						KeyData: privateKey.PublicKeyOpenssh,
					},
				},
			},
		},
		Tags: r.mCtx.ResourceTags(),
	}
	// Enable app routing if required
	if *r.enableAppRouting {
		managedClusterArgs.IngressProfile = containerservice.ManagedClusterIngressProfileArgs{
			WebAppRouting: containerservice.ManagedClusterIngressProfileWebAppRoutingArgs{
				Enabled: pulumi.Bool(true),
			},
		}
	}
	cluster, err := containerservice.NewManagedCluster(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureAKSID, "cluster"),
		managedClusterArgs)
	if err != nil {
		return err
	}

	// retrieve the admin credentials which contain the kubeconfig
	adminCredentials := containerservice.ListManagedClusterAdminCredentialsOutput(
		ctx,
		containerservice.ListManagedClusterAdminCredentialsOutputArgs{
			ResourceGroupName: rg.Name,
			ResourceName:      cluster.Name,
		}, nil)

	// grant the 'contributor' role to the identity on the resource group
	_, err = authorization.NewRoleAssignment(ctx, "roleAssignment", &authorization.RoleAssignmentArgs{
		PrincipalId:      identity.PrincipalId,
		PrincipalType:    pulumi.String("ServicePrincipal"),
		RoleDefinitionId: pulumi.String("/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"),
		Scope:            rg.ID(),
	})
	if err != nil {
		return err
	}
	kubeconfig := adminCredentials.ApplyT(
		func(adminCredentials containerservice.ListManagedClusterAdminCredentialsResult) (pulumi.String, error) {
			value, _ := base64.StdEncoding.DecodeString(adminCredentials.Kubeconfigs[0].Value)
			return pulumi.String(value), nil
		}).(pulumi.StringOutput)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig), pulumi.ToSecret(kubeconfig))
	return nil

}

// agentPool holds the node pool configuration which does not depend on the
// pulumi context, so the pool definition can be tested.
type agentPool struct {
	name         string
	mode         string
	count        int
	vmSize       string
	osDiskSizeGB int
	spot         bool
	spotMaxPrice *float64
}

// poolSelection holds the result of resolving the user pool configuration.
// price is only set when the user pool runs on spot instances.
type poolSelection struct {
	location string
	vmSize   string
	price    *float64
}

// userPoolSelection resolves the location and the VM size of the user pool.
// When spot is enabled the best option across the configured VM sizes is used,
// and since the spot search also selects the location, the whole cluster is
// created there. Otherwise the location passed as parameter is kept and the
// user pool runs on demand with the first configured VM size.
func (r *aksRequest) userPoolSelection() (*poolSelection, error) {
	if *r.onlySystemPool {
		return &poolSelection{location: *r.location}, nil
	}
	if r.spot == nil || !r.spot.Spot {
		return &poolSelection{
			location: *r.location,
			vmSize:   r.vmSizes[0],
		}, nil
	}
	bsc, err :=
		data.SpotInfo(
			r.mCtx,
			&data.SpotInfoArgs{
				ComputeSizes:          r.vmSizes,
				OSType:                "linux",
				SpotPriceIncreaseRate: &r.spot.IncreaseRate,
				SpotTolerance:         &r.spot.Tolerance,
				ExcludedLocations:     r.spot.ExcludedHostingPlaces,
			})
	if err != nil {
		return nil, err
	}
	logging.Debugf("Best spot price option found: %v", bsc)
	vmSize := r.vmSizes[0]
	if len(bsc.ComputeType) > 0 {
		vmSize = bsc.ComputeType[0]
	}
	price := bsc.Price
	return &poolSelection{
		location: bsc.HostingPlace,
		vmSize:   vmSize,
		price:    &price,
	}, nil
}

// agentPools returns the node pools of the cluster. The system pool always runs
// on demand and the user pool runs on spot when a spot price was selected.
func (r *aksRequest) agentPools(selection *poolSelection) []agentPool {
	pools := []agentPool{
		{
			name:         "systempool",
			mode:         string(containerservice.AgentPoolModeSystem),
			count:        1,
			vmSize:       *r.systemPoolVMSize,
			osDiskSizeGB: 30,
		},
	}
	if *r.onlySystemPool {
		return pools
	}
	return append(pools,
		agentPool{
			name:         "userpool",
			mode:         string(containerservice.AgentPoolModeUser),
			count:        *r.workersDesired,
			vmSize:       selection.vmSize,
			osDiskSizeGB: 100,
			spot:         selection.price != nil,
			spotMaxPrice: selection.price,
		})
}

// agentPoolProfiles translates the node pools into the AKS agent pool profiles.
func agentPoolProfiles(pools []agentPool) containerservice.ManagedClusterAgentPoolProfileArray {
	profiles := containerservice.ManagedClusterAgentPoolProfileArray{}
	for _, p := range pools {
		profile := &containerservice.ManagedClusterAgentPoolProfileArgs{
			Name:         pulumi.String(p.name),
			Mode:         pulumi.String(p.mode),
			Count:        pulumi.Int(p.count),
			VmSize:       pulumi.String(p.vmSize),
			OsType:       pulumi.String("Linux"),
			OsDiskSizeGB: pulumi.Int(p.osDiskSizeGB),
			Type:         pulumi.String("VirtualMachineScaleSets"),
		}
		if p.spot {
			profile.ScaleSetPriority = containerservice.ScaleSetPrioritySpot
			profile.SpotMaxPrice = pulumi.Float64(*p.spotMaxPrice)
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

// Write exported values in context to files o a selected target folder
func (r *aksRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig): "kubeconfig",
	})
}
