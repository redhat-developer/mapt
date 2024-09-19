package aks

import (
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/containerservice/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/managedidentity/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	spotprice "github.com/redhat-developer/mapt/pkg/provider/azure/module/spot-price"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type AKSRequest struct {
	Prefix   string
	Location string
	VMSize   string
	// "1.26.3"
	KubernetesVersion string
	OnlySystemPool    bool
	Spot              bool
	SpotTolerance     spotprice.EvictionRate
}

func Create(r *AKSRequest) (err error) {
	logging.Debug("Creating AKS")
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackAKS),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

func Destroy() error {
	return azure.Destroy(
		maptContext.ProjectName(),
		maptContext.BackedURL(),
		maptContext.StackNameByProject(stackAKS))
}

// Main function to deploy all requried resources to azure
func (r *AKSRequest) deployer(ctx *pulumi.Context) error {
	// Get values for spot machine
	location, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureAKSID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(*location),
			ResourceGroupName: pulumi.String(maptContext.RunID()),
			Tags:              maptContext.ResourceTags(),
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
		resourcesUtil.GetResourceName(r.Prefix, azureAKSID, "privatekey"),
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
		resourcesUtil.GetResourceName(r.Prefix, azureAKSID, "identity"),
		&managedidentity.UserAssignedIdentityArgs{
			Location:          rg.Location,
			ResourceGroupName: rg.Name,
			Tags:              maptContext.ResourceTags(),
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
	if !r.OnlySystemPool {
		agentPoolProfiles = append(agentPoolProfiles,
			&containerservice.ManagedClusterAgentPoolProfileArgs{
				Name:         pulumi.String("userpool"),
				Mode:         containerservice.AgentPoolModeUser,
				Count:        pulumi.Int(1),
				VmSize:       pulumi.String(r.VMSize),
				OsType:       pulumi.String("Linux"),
				OsDiskSizeGB: pulumi.Int(100),
				Type:         pulumi.String("VirtualMachineScaleSets"),
				// VnetSubnetID:     n.PublicSubnet.ID(),
				ScaleSetPriority: containerservice.ScaleSetPrioritySpot,
				SpotMaxPrice:     pulumi.Float64(*spotPrice)},
		)
	}
	cluster, err := containerservice.NewManagedCluster(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureAKSID, "cluster"),
		&containerservice.ManagedClusterArgs{
			ResourceGroupName: rg.Name,
			Location:          rg.Location,
			Identity: &containerservice.ManagedClusterIdentityArgs{
				Type: containerservice.ResourceIdentityTypeUserAssigned,
				UserAssignedIdentities: pulumi.StringArray{
					identity.ID(),
				},
			},
			KubernetesVersion: pulumi.String(r.KubernetesVersion),
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
			Tags: maptContext.ResourceTags(),
		})

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
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig), kubeconfig)
	return nil

}

func (r *AKSRequest) valuesCheckingSpot() (*string, *float64, error) {
	if r.Spot {
		bsc, err :=
			spotprice.GetBestSpotChoice(spotprice.BestSpotChoiceRequest{
				VMTypes:              []string{r.VMSize},
				OSType:               "linux",
				EvictioRateTolerance: r.SpotTolerance,
			})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, nil, err
		}
		return &bsc.Location, &bsc.Price, nil
	}
	return &r.Location, nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *AKSRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputKubeconfig): "kubeconfig",
	})
}
