package aks

import (
	"encoding/base64"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v3"

	// containerservice "github.com/pulumi/pulumi-azure-native-sdk/containerservice/v2/v20240801"
	containerservice "github.com/pulumi/pulumi-azure-native-sdk/containerservice/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/managedidentity/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
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
	VMSize   string
	// "1.26.3"
	KubernetesVersion   string
	OnlySystemPool      bool
	EnableAppRouting    bool
	Spot                bool
	SpotTolerance       spotTypes.Tolerance
	SpotExcludedRegions []string
}

type aksRequest struct {
	mCtx     *mc.Context `validate:"required"`
	prefix   *string
	location *string
	vmSize   *string
	// "1.26.3"
	kubernetesVersion   *string
	onlySystemPool      *bool
	enableAppRouting    *bool
	spot                *bool
	spotTolerance       *spotTypes.Tolerance
	spotExcludedRegions []string
}

func (r *aksRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

func Create(mCtxArgs *mc.ContextArgs, args *AKSArgs) (err error) {
	// Create mapt Context
	logging.Debug("Creating AKS")
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &aksRequest{
		mCtx:                mCtx,
		prefix:              &prefix,
		location:            &args.Location,
		vmSize:              &args.VMSize,
		kubernetesVersion:   &args.KubernetesVersion,
		onlySystemPool:      &args.OnlySystemPool,
		enableAppRouting:    &args.EnableAppRouting,
		spot:                &args.Spot,
		spotTolerance:       &args.SpotTolerance,
		spotExcludedRegions: args.SpotExcludedRegions,
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAKS),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(mCtx, cs)
	return r.manageResults(sr)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	// Create mapt Context
	logging.Debug("Destroy AKS")
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	return azure.Destroy(mCtx, stackAKS)
}

// Main function to deploy all requried resources to azure
func (r *aksRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get values for spot machine
	location, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	// Get location for creating Resouce Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*location)
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
			Location:          rg.Location,
			ResourceGroupName: rg.Name,
			Tags:              r.mCtx.ResourceTags(),
		})

	if err != nil {
		return err
	}
	// create the cluster
	agentPoolProfiles := containerservice.ManagedClusterAgentPoolProfileArray{
		&containerservice.ManagedClusterAgentPoolProfileArgs{
			Name:         pulumi.String("systempool"),
			Mode:         containerservice.AgentPoolModeSystem,
			Count:        pulumi.Int(1),
			VmSize:       pulumi.String(systemPoolVMSize),
			OsType:       pulumi.String("Linux"),
			OsDiskSizeGB: pulumi.Int(30),
			Type:         pulumi.String("VirtualMachineScaleSets"),
		},
	}
	if !*r.onlySystemPool {
		agentPoolProfiles = append(agentPoolProfiles,
			&containerservice.ManagedClusterAgentPoolProfileArgs{
				Name:         pulumi.String("userpool"),
				Mode:         containerservice.AgentPoolModeUser,
				Count:        pulumi.Int(1),
				VmSize:       pulumi.String(*r.vmSize),
				OsType:       pulumi.String("Linux"),
				OsDiskSizeGB: pulumi.Int(100),
				Type:         pulumi.String("VirtualMachineScaleSets"),
				// VnetSubnetID:     n.PublicSubnet.ID(),
				ScaleSetPriority: containerservice.ScaleSetPrioritySpot,
				SpotMaxPrice:     pulumi.Float64(*spotPrice)},
		)
	}
	managedClusterArgs := &containerservice.ManagedClusterArgs{
		ResourceGroupName: rg.Name,
		Location:          rg.Location,
		Identity: &containerservice.ManagedClusterIdentityArgs{
			Type: containerservice.ResourceIdentityTypeUserAssigned,
			UserAssignedIdentities: pulumi.StringArray{
				identity.ID(),
			},
		},
		KubernetesVersion: pulumi.String(*r.kubernetesVersion),
		DnsPrefix:         pulumi.String("mapt"),
		EnableRBAC:        pulumi.Bool(true),
		AgentPoolProfiles: agentPoolProfiles,
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
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig), kubeconfig)
	return nil

}

func (r *aksRequest) valuesCheckingSpot() (*string, *float64, error) {
	if *r.spot {
		bsc, err :=
			data.SpotInfo(&data.SpotInfoArgs{
				ComputeSizes: []string{*r.vmSize},
				OSType:       "linux",
				// TODO review this
				// EvictionRateTolerance: r.SpotTolerance,
				ExcludedLocations: r.spotExcludedRegions,
			})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, nil, err
		}
		return &bsc.HostingPlace, &bsc.Price, nil
	}
	return r.location, nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *aksRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig): "kubeconfig",
	})
}
