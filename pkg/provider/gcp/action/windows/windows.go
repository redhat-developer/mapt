package windows

import (
	"fmt"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/gcp"
	gcpConstants "github.com/redhat-developer/mapt/pkg/provider/gcp/constants"
	"github.com/redhat-developer/mapt/pkg/provider/gcp/data"
	"github.com/redhat-developer/mapt/pkg/provider/gcp/modules/allocation"
	gcpCompute "github.com/redhat-developer/mapt/pkg/provider/gcp/modules/compute"
	gcpNetwork "github.com/redhat-developer/mapt/pkg/provider/gcp/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	outputUserPassword = "userpassword"
)

type WindowsArgs struct {
	Prefix         string
	Version        string
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
}

type windowsRequest struct {
	mCtx           *mc.Context
	prefix         string
	version        string
	spot           bool
	allocationData *allocation.AllocationResult
	diskSize       *int
}

func Create(mCtxArgs *mc.ContextArgs, args *WindowsArgs) error {
	if args.ComputeRequest == nil {
		return fmt.Errorf("compute request is required")
	}

	mCtx, err := mc.Init(mCtxArgs, gcp.Provider())
	if err != nil {
		return err
	}

	prefix := "main"
	if len(args.Prefix) > 0 {
		prefix = args.Prefix
	}

	r := windowsRequest{
		mCtx:     mCtx,
		prefix:   prefix,
		version:  args.Version,
		diskSize: args.ComputeRequest.DiskSize,
	}
	if args.Spot != nil {
		r.spot = args.Spot.Spot
	}

	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			ComputeRequest: args.ComputeRequest,
			Spot:           args.Spot,
		})
	if err != nil {
		return err
	}

	return r.createMachine()
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Run GCP Windows destroy")
	mCtx, err := mc.Init(mCtxArgs, gcp.Provider())
	if err != nil {
		return err
	}
	return gcp.Destroy(mCtx, stackName)
}

func (r *windowsRequest) createMachine() error {
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(stackName),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: gcp.GetClouProviderCredentials(
			map[string]string{
				gcpConstants.CONFIG_GCP_REGION: r.allocationData.Region,
			}),
		DeployFunc: r.deploy,
	}

	sr, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return err
	}
	return manageResults(r.mCtx, sr, r.prefix)
}

func (r *windowsRequest) deploy(ctx *pulumi.Context) error {
	imageURI, err := data.GetWindowsImageURI(r.mCtx.Context(), r.version)
	if err != nil {
		return err
	}

	nw, err := gcpNetwork.Create(ctx, r.mCtx,
		&gcpNetwork.NetworkArgs{
			Prefix:       r.prefix,
			ID:           gcpWindowsID,
			Region:       r.allocationData.Region,
			Zone:         r.allocationData.Zone,
			IngressPorts: []string{"22", "3389"},
		})
	if err != nil {
		return err
	}

	password, err := security.CreatePassword(ctx,
		fmt.Sprintf("%s-%s-password", r.prefix, gcpWindowsID))
	if err != nil {
		return err
	}

	startupScript := windowsStartupScript(userDefault, password)

	effectiveDiskSize := diskSize
	if r.diskSize != nil {
		effectiveDiskSize = *r.diskSize
	}

	computeReq := gcpCompute.ComputeRequest{
		Prefix:        r.prefix,
		ID:            gcpWindowsID,
		Zone:          r.allocationData.Zone,
		Network:       nw.Network,
		Subnet:        nw.Subnet,
		PublicIP:      nw.PublicIP,
		MachineType:   r.allocationData.MachineTypes[0],
		ImageURI:      imageURI,
		DiskSize:      &effectiveDiskSize,
		Spot:          r.spot,
		StartupScript: "",
		Username:      userDefault,
		Tags:          []string{fmt.Sprintf("%s-fw", gcpWindowsID)},
	}
	computeReq.WindowsStartupScript = startupScript
	c, err := computeReq.NewCompute(ctx)
	if err != nil {
		return err
	}

	ctx.Export(gcpCompute.GetOutputUsername(r.prefix),
		pulumi.String(userDefault))
	ctx.Export(gcpCompute.GetOutputHost(r.prefix),
		c.GetHostIP())
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputUserPassword),
		password.Result)

	return nil
}

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	results := map[string]string{
		gcpCompute.GetOutputUsername(prefix):                    "username",
		gcpCompute.GetOutputHost(prefix):                       "host",
		fmt.Sprintf("%s-%s", prefix, outputUserPassword):       "userpassword",
	}
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}

func windowsStartupScript(username string, password *random.RandomPassword) pulumi.StringOutput {
	return pulumi.Sprintf(`$user = "%s"
$pass = ConvertTo-SecureString -String "%s" -AsPlainText -Force
if (-not (Get-LocalUser -Name $user -ErrorAction SilentlyContinue)) {
    New-LocalUser -Name $user -Password $pass -FullName $user -Description "MAPT user"
    Add-LocalGroupMember -Group "Administrators" -Member $user
    Add-LocalGroupMember -Group "Remote Desktop Users" -Member $user
}
Set-ItemProperty -Path 'HKLM:\System\CurrentControlSet\Control\Terminal Server' -Name fDenyTSConnections -Value 0
Enable-NetFirewallRule -DisplayGroup "Remote Desktop"
`, username, password.Result)
}
