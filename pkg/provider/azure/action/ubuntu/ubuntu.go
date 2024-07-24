package ubuntu

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
	"github.com/redhat-developer/mapt/pkg/provider/azure/module/network"
	spotprice "github.com/redhat-developer/mapt/pkg/provider/azure/module/spot-price"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/module/virtual-machine"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type UbuntuRequest struct {
	Prefix        string
	Location      string
	VMSize        string
	Version       string
	Username      string
	Spot          bool
	SpotTolerance spotprice.EvictionRate
}

func Create(r *UbuntuRequest) (err error) {
	logging.Debug("Creating Ubuntu Server")
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackAzureUbuntu),
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
		maptContext.StackNameByProject(stackAzureUbuntu))
}

// Main function to deploy all requried resources to azure
func (r *UbuntuRequest) deployer(ctx *pulumi.Context) error {
	// Get values for spot machine
	location, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureUbuntuID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(*location),
			ResourceGroupName: pulumi.String(maptContext.RunID()),
			Tags:              maptContext.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix:        r.Prefix,
		ComponentID:   azureUbuntuID,
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
		resourcesUtil.GetResourceName(r.Prefix, azureUbuntuID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)
	vmr := virtualmachine.VirtualMachineRequest{
		Prefix:          r.Prefix,
		ComponentID:     azureUbuntuID,
		ResourceGroup:   rg,
		NetworkInteface: n.NetworkInterface,
		VMSize:          r.VMSize,
		Publisher:       "Canonical",
		Offer:           fmt.Sprintf("ubuntu-%s-lts-daily", r.Version),
		Sku:             "server",
		AdminUsername:   r.Username,
		PrivateKey:      privateKey,
		SpotPrice:       spotPrice,
	}
	vm, err := vmr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername), pulumi.String(r.Username))
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureUbuntuID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     privateKey.PrivateKeyOpenssh,
				User:           pulumi.String(r.Username),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(command.CommandPing),
			Update: pulumi.String(command.CommandPing),
		},
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: "10m",
				Update: "10m"}),
		pulumi.DependsOn([]pulumi.Resource{vm}))
	return err
}

func (r *UbuntuRequest) valuesCheckingSpot() (*string, *float64, error) {
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
func (r *UbuntuRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):           "host",
	})
}
