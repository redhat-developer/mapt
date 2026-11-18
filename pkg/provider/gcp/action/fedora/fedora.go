package fedora

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

type FedoraArgs struct {
	Prefix         string
	Version        string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
}

type fedoraRequest struct {
	mCtx           *mc.Context
	prefix         string
	version        string
	spot           bool
	allocationData *allocation.AllocationResult
	diskSize       *int
}

func Create(mCtxArgs *mc.ContextArgs, args *FedoraArgs) error {
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

	r := fedoraRequest{
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
	logging.Debug("Run GCP Fedora destroy")
	mCtx, err := mc.Init(mCtxArgs, gcp.Provider())
	if err != nil {
		return err
	}
	return gcp.Destroy(mCtx, stackName)
}

func (r *fedoraRequest) createMachine() error {
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

func (r *fedoraRequest) deploy(ctx *pulumi.Context) error {
	imageURI, err := data.GetFedoraImageURI(r.mCtx.Context(), r.version)
	if err != nil {
		return err
	}

	nw, err := gcpNetwork.Create(ctx, r.mCtx,
		&gcpNetwork.NetworkArgs{
			Prefix: r.prefix,
			ID:     gcpFedoraID,
			Region: r.allocationData.Region,
			Zone:   r.allocationData.Zone,
		})
	if err != nil {
		return err
	}

	privateKey, err := tls.NewPrivateKey(ctx,
		fmt.Sprintf("%s-%s-pk", r.prefix, gcpFedoraID),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(gcpCompute.GetOutputPrivateKey(r.prefix),
		privateKey.PrivateKeyPem)

	effectiveDiskSize := diskSize
	if r.diskSize != nil {
		effectiveDiskSize = *r.diskSize
	}

	computeReq := gcpCompute.ComputeRequest{
		Prefix:      r.prefix,
		ID:          gcpFedoraID,
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
		Tags:        []string{fmt.Sprintf("%s-fw", gcpFedoraID)},
	}
	c, err := computeReq.NewCompute(ctx)
	if err != nil {
		return err
	}

	ctx.Export(gcpCompute.GetOutputUsername(r.prefix),
		pulumi.String(userDefault))
	ctx.Export(gcpCompute.GetOutputHost(r.prefix),
		c.GetHostIP())

	_, err = c.Readiness(ctx, r.prefix, gcpFedoraID, userDefault, privateKey)
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
