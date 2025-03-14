package linux

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/module/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/module/virtual-machine"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	stackAzureLinux = "stackAzureLinux"

	azureLinuxID = "als"

	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"
	defaultVMSize        = "Standard_D8as_v5"
)

type LinuxRequest struct {
	Prefix              string
	Location            string
	VMSizes             []string
	Arch                string
	InstanceRequest     instancetypes.InstanceRequest
	OSType              data.OSType
	Version             string
	Username            string
	Spot                bool
	SpotTolerance       spotAzure.EvictionRate
	SpotExcludedRegions []string
	GetUserdata         func() (string, error)
	ReadinessCommand    string
}

func Create(ctx *maptContext.ContextArgs, r *LinuxRequest) (err error) {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	if len(r.VMSizes) == 0 {
		vmSizes, err := r.InstanceRequest.GetMachineTypes()
		if err != nil {
			logging.Debugf("Unable to fetch desired instance type: %v", err)
		}
		if len(vmSizes) > 0 {
			r.VMSizes = append(r.VMSizes, vmSizes...)
		}
	}
	logging.Debug("Creating Linux Server")
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackAzureLinux),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

func Destroy(ctx *maptContext.ContextArgs) error {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}
	// destroy
	return azure.Destroy(
		maptContext.ProjectName(),
		maptContext.BackedURL(),
		maptContext.StackNameByProject(stackAzureLinux))
}

// Main function to deploy all requried resources to azure
func (r *LinuxRequest) deployer(ctx *pulumi.Context) error {
	// Get values for spot machine
	location, vmType, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}

	// Get location for creating the Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureLinuxID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(maptContext.RunID()),
			Tags:              maptContext.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix:        r.Prefix,
		ComponentID:   azureLinuxID,
		ResourceGroup: rg,
	}
	n, err := nr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost), n.PublicIP.IpAddress)
	// Virutal machine
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureLinuxID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)
	// Image refence info
	ir, err := data.GetImageRef(r.OSType, r.Arch, r.Version)
	if err != nil {
		return err
	}
	var userDataB64 string
	if r.GetUserdata != nil {
		var err error
		userDataB64, err = r.GetUserdata()
		if err != nil {
			return fmt.Errorf("error creating RHEL Server on Azure: %v", err)
		}
	}
	vmr := virtualmachine.VirtualMachineRequest{
		Prefix:          r.Prefix,
		ComponentID:     azureLinuxID,
		ResourceGroup:   rg,
		NetworkInteface: n.NetworkInterface,
		VMSize:          vmType,
		Publisher:       ir.Publisher,
		Offer:           ir.Offer,
		Sku:             ir.Sku,
		ImageID:         ir.ID,
		AdminUsername:   r.Username,
		PrivateKey:      privateKey,
		SpotPrice:       spotPrice,
		Userdata:        userDataB64,
	}
	vm, err := vmr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername), pulumi.String(r.Username))
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureLinuxID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     privateKey.PrivateKeyOpenssh,
				User:           pulumi.String(r.Username),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(util.If(
				len(r.ReadinessCommand) == 0,
				command.CommandPing,
				r.ReadinessCommand)),
			Update: pulumi.String(util.If(
				len(r.ReadinessCommand) == 0,
				command.CommandPing,
				r.ReadinessCommand)),
		},
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: "10m",
				Update: "10m"}),
		pulumi.DependsOn([]pulumi.Resource{vm}))
	return err
}

func (r *LinuxRequest) valuesCheckingSpot() (*string, string, *float64, error) {
	if r.Spot {
		ir, err := data.GetImageRef(r.OSType, r.Arch, r.Version)
		if err != nil {
			return nil, "", nil, err
		}
		bsc, err :=
			spotAzure.GetBestSpotChoice(
				spotAzure.BestSpotChoiceRequest{
					VMTypes:               util.If(len(r.VMSizes) > 0, r.VMSizes, []string{defaultVMSize}),
					OSType:                "linux",
					EvictionRateTolerance: r.SpotTolerance,
					ImageRef:              *ir,
					ExcludedRegions:       r.SpotExcludedRegions,
				})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, "", nil, err
		}
		return &bsc.Location, bsc.VMType, &bsc.Price, nil
	}
	return &r.Location, "", nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *LinuxRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):           "host",
	})
}
