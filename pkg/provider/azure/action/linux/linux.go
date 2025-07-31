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
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/module/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/module/virtual-machine"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
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

type LinuxArgs struct {
	Prefix              string
	Location            string
	Arch                string
	ComputeRequest      *cr.ComputeRequestArgs
	OSType              data.OSType
	Version             string
	Username            string
	Spot                bool
	SpotTolerance       spotTypes.Tolerance
	SpotExcludedRegions []string
	GetUserdata         func() (string, error)
	ReadinessCommand    string
}

type linuxRequest struct {
	mCtx                *mc.Context `validate:"required"`
	prefix              *string
	location            *string
	vmSizes             []string
	arch                *string
	osType              *data.OSType
	version             *string
	username            *string
	spot                *bool
	spotTolerance       *spotTypes.Tolerance
	spotExcludedRegions []string
	getUserdata         func() (string, error)
	readinessCommand    *string
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
		mCtx:                mCtx,
		prefix:              &prefix,
		location:            &args.Location,
		arch:                &args.Arch,
		osType:              &args.OSType,
		version:             &args.Version,
		username:            &args.Username,
		spot:                &args.Spot,
		spotTolerance:       &args.SpotTolerance,
		spotExcludedRegions: args.SpotExcludedRegions,
		getUserdata:         args.GetUserdata,
		readinessCommand:    &args.ReadinessCommand,
	}
	if len(args.ComputeRequest.ComputeSizes) > 0 {
		r.vmSizes = args.ComputeRequest.ComputeSizes
	} else {
		vmSizes, err :=
			data.NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return err
		}
		r.vmSizes = vmSizes
	}
	logging.Debug("Creating Linux Server")
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAzureLinux),
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
	// Get values for spot machine
	location, vmType, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}

	// Get location for creating the Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*location)
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
	nr := network.NetworkRequest{
		Prefix:        *r.prefix,
		ComponentID:   azureLinuxID,
		ResourceGroup: rg,
	}
	n, err := nr.Create(ctx, r.mCtx)
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
	// Image refence info
	ir, err := data.GetImageRef(*r.osType, *r.arch, *r.version)
	if err != nil {
		return err
	}
	var userDataB64 string
	if r.getUserdata != nil {
		var err error
		userDataB64, err = r.getUserdata()
		if err != nil {
			return fmt.Errorf("error creating RHEL Server on Azure: %v", err)
		}
	}
	vmr := virtualmachine.VirtualMachineRequest{
		Prefix:          *r.prefix,
		ComponentID:     azureLinuxID,
		ResourceGroup:   rg,
		NetworkInteface: n.NetworkInterface,
		VMSize:          vmType,
		Publisher:       ir.Publisher,
		Offer:           ir.Offer,
		Sku:             ir.Sku,
		ImageID:         ir.ID,
		AdminUsername:   *r.username,
		PrivateKey:      privateKey,
		SpotPrice:       spotPrice,
		Userdata:        userDataB64,
	}
	vm, err := vmr.Create(ctx, r.mCtx)
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

func (r *linuxRequest) valuesCheckingSpot() (*string, string, *float64, error) {
	if *r.spot {
		ir, err := data.GetImageRef(*r.osType, *r.arch, *r.version)
		if err != nil {
			return nil, "", nil, err
		}
		bsc, err :=
			data.SpotInfo(
				&data.SpotInfoArgs{
					ComputeSizes: util.If(len(r.vmSizes) > 0, r.vmSizes, []string{defaultVMSize}),
					OSType:       "linux",
					// EvictionRateTolerance: r.SpotTolerance,
					ImageRef:          *ir,
					ExcludedLocations: r.spotExcludedRegions,
				})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, "", nil, err
		}
		return &bsc.Location, bsc.ComputeSize, &bsc.Price, nil
	}
	return r.location, "", nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *linuxRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *r.prefix, outputHost):           "host",
	})
}
