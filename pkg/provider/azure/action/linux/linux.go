package linux

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	userDataApi "github.com/redhat-developer/mapt/pkg/provider/api/config/userdata"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/modules/virtual-machine"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
	rhelApi "github.com/redhat-developer/mapt/pkg/targets/host/rhel"
)

const (
	stackAzureLinux = "stackAzureLinux"

	azureLinuxID = "als"

	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"
	defaultVMSize        = "Standard_D8as_v5"
)

type LinuxArgs struct {
	Prefix                string
	Location              string
	Arch                  string
	ComputeRequest        *cr.ComputeRequestArgs
	OSType                data.OSType
	Version               string
	Username              string
	Spot                  *spotTypes.SpotArgs
	CloudConfigAsUserData userDataApi.CloudConfig
	ReadinessCommand      string
}

type linuxRequest struct {
	mCtx                  *mc.Context `validate:"required"`
	prefix                *string
	arch                  *string
	osType                *data.OSType
	version               *string
	allocationData        *allocation.AllocationResult
	username              *string
	cloudConfigAsUserData userDataApi.CloudConfig
	readinessCommand      *string
}

func (r *linuxRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

func Create(mCtxArgs *mc.ContextArgs, args *LinuxArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &linuxRequest{
		mCtx:                  mCtx,
		prefix:                &prefix,
		arch:                  &args.Arch,
		osType:                &args.OSType,
		version:               &args.Version,
		username:              &args.Username,
		cloudConfigAsUserData: args.CloudConfigAsUserData,
		readinessCommand:      &args.ReadinessCommand,
	}
	ir, err := data.GetImageRef(*r.osType, *r.arch, *r.version)
	if err != nil {
		return err
	}
	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			ComputeRequest: args.ComputeRequest,
			OSType:         "linux",
			ImageRef:       ir,
			Location:       &args.Location,
			Spot:           args.Spot})
	if err != nil {
		return err
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAzureLinux),
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
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	// destroy
	return azure.Destroy(mCtx, stackAzureLinux)
}

// Main function to deploy all requried resources to azure
func (r *linuxRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get location for creating the Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*r.allocationData.Location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureLinuxID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(r.mCtx.RunID()),
			Tags:              r.mCtx.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	sg, err := securityGroups(ctx, r.mCtx, r.prefix, r.allocationData.Location, rg)
	if err != nil {
		return err
	}
	n, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:        *r.prefix,
			ComponentID:   azureLinuxID,
			ResourceGroup: rg,
			Location:      r.allocationData.Location,
			SecurityGroup: sg,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), n.PublicIP.IpAddress)
	// Virutal machine
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureLinuxID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)

	// Generate cloud config userdata
	var userDataB64Input pulumi.StringInput
	if r.cloudConfigAsUserData != nil {
		// Check if this is RHEL cloud config
		if rhelConfig, isRHEL := r.cloudConfigAsUserData.(*rhelApi.CloudConfigArgs); isRHEL {
			// Use RHEL helper's GenerateCloudConfig which handles GitLab integration
			userDataB64Input, err = rhelConfig.GenerateCloudConfig(ctx, r.mCtx.RunID())
			if err != nil {
				return err
			}
		} else {
			// Other cloud config types, use normal userdata
			userDataB64, err := r.cloudConfigAsUserData.CloudConfig()
			if err != nil {
				return fmt.Errorf("error creating Linux Server on Azure: %v", err)
			}
			userDataB64Input = pulumi.String(*userDataB64)
		}
	} else {
		userDataB64Input = pulumi.String("")
	}

	vm, err := virtualmachine.Create(ctx, r.mCtx,
		&virtualmachine.VirtualMachineArgs{
			Prefix:          *r.prefix,
			ComponentID:     azureLinuxID,
			ResourceGroup:   rg,
			NetworkInteface: n.NetworkInterface,
			// Check this
			VMSize:           r.allocationData.ComputeSizes[0],
			Publisher:        r.allocationData.ImageRef.Publisher,
			Offer:            r.allocationData.ImageRef.Offer,
			Sku:              r.allocationData.ImageRef.Sku,
			ImageID:          r.allocationData.ImageRef.ID,
			AdminUsername:    *r.username,
			PrivateKey:       privateKey,
			SpotPrice:        r.allocationData.Price,
			UserDataAsBase64: userDataB64Input,
			Location:         *r.allocationData.Location,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(*r.username))
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureLinuxID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     privateKey.PrivateKeyOpenssh,
				User:           pulumi.String(*r.username),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(util.If(
				len(*r.readinessCommand) == 0,
				command.CommandPing,
				*r.readinessCommand)),
			Update: pulumi.String(util.If(
				len(*r.readinessCommand) == 0,
				command.CommandPing,
				*r.readinessCommand)),
		},
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: "10m",
				Update: "10m"}),
		pulumi.DependsOn([]pulumi.Resource{vm}))
	return err
}

// security group for mac machine with ingress rules for ssh and vnc
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context,
	prefix, location *string,
	rg *resources.ResourceGroup) (securityGroup.SecurityGroup, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	return securityGroup.Create(
		ctx,
		mCtx,
		&securityGroup.SecurityGroupArgs{
			Name:     resourcesUtil.GetResourceName(*prefix, azureLinuxID, "sg"),
			RG:       rg,
			Location: location,
			IngressRules: []securityGroup.IngressRules{
				sshIngressRule},
		})
}

// Write exported values in context to files o a selected target folder
func (r *linuxRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *r.prefix, outputHost):           "host",
	})
}
