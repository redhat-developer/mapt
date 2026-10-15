package rhel

import (
	"fmt"

	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
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
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type RHELArgs struct {
	Prefix         string
	Version        string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
}

type rhelRequest struct {
	mCtx           *mc.Context
	prefix         string
	version        string
	spot           bool
	allocationData *allocation.AllocationResult
	diskSize       *int
}

func Create(mCtxArgs *mc.ContextArgs, args *RHELArgs) error {
	mCtx, err := mc.Init(mCtxArgs, gcp.Provider())
	if err != nil {
		return err
	}

	prefix := "main"
	if len(args.Prefix) > 0 {
		prefix = args.Prefix
	}

	r := rhelRequest{
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
	logging.Debug("Run GCP RHEL destroy")
	mCtx, err := mc.Init(mCtxArgs, gcp.Provider())
	if err != nil {
		return err
	}
	return gcp.Destroy(mCtx, stackName)
}

func (r *rhelRequest) createMachine() error {
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

func (r *rhelRequest) deploy(ctx *pulumi.Context) error {
	imageURI, err := data.GetRHELImageURI(r.mCtx.Context(), r.version)
	if err != nil {
		return err
	}

	// Networking
	nw, err := gcpNetwork.Create(ctx, r.mCtx,
		&gcpNetwork.NetworkArgs{
			Prefix: r.prefix,
			ID:     gcpRHELID,
			Region: r.allocationData.Region,
			Zone:   r.allocationData.Zone,
		})
	if err != nil {
		return err
	}

	// SSH key pair
	privateKey, err := tls.NewPrivateKey(ctx,
		fmt.Sprintf("%s-%s-pk", r.prefix, gcpRHELID),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(gcpCompute.GetOutputPrivateKey(r.prefix),
		privateKey.PrivateKeyPem)

	// Compute
	effectiveDiskSize := diskSize
	if r.diskSize != nil {
		effectiveDiskSize = *r.diskSize
	}

	cr := gcpCompute.ComputeRequest{
		Prefix:      r.prefix,
		ID:          gcpRHELID,
		Zone:        r.allocationData.Zone,
		Network:     nw.Network,
		Subnet:      nw.Subnet,
		PublicIP:    nw.PublicIP,
		MachineType: r.allocationData.MachineTypes[0],
		ImageURI:    imageURI,
		KeyPair:     privateKey,
		DiskSize:    &effectiveDiskSize,
		Spot:        r.spot,
		Username:    userDefault,
		Tags:        []string{fmt.Sprintf("%s-ssh", gcpRHELID)},
	}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}

	ctx.Export(gcpCompute.GetOutputUsername(r.prefix),
		pulumi.String(userDefault))
	ctx.Export(gcpCompute.GetOutputHost(r.prefix),
		c.GetHostIP())

	// Readiness check
	_, err = c.Readiness(ctx, r.prefix, gcpRHELID, userDefault, privateKey)
	return err
}

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	results := map[string]string{
		gcpCompute.GetOutputUsername(prefix):   "username",
		gcpCompute.GetOutputPrivateKey(prefix): "id_rsa",
		gcpCompute.GetOutputHost(prefix):       "host",
	}
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}
