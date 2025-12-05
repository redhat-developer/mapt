package ibmpower

import (
	"fmt"

	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmcloudp "github.com/redhat-developer/mapt/pkg/provider/ibmcloud"
	icdata "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/data"

	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	stackIBMPowerVS      = "icpw"
	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"

	NCidr = "10.0.2.0/24"
	NGW   = "10.0.2.1"
)

type PWArgs struct {
	Prefix string
}

type pwRequest struct {
	mCtx   *mc.Context
	prefix *string
	zone   *string
}

func New(ctx *mc.ContextArgs, args *PWArgs) error {
	ibmcloudProvider := ibmcloudp.Provider()
	mCtx, err := mc.Init(ctx, ibmcloudProvider)
	if err != nil {
		return err
	}

	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	zone, err := ibmcloudProvider.Zone()
	if err != nil {
		return err
	}
	r := &pwRequest{
		mCtx:   mCtx,
		prefix: &prefix,
		zone:   zone}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackIBMPowerVS),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: ibmcloudp.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return fmt.Errorf("stack creation failed: %w", err)
	}
	return manageResults(mCtx, sr, prefix)
}

func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, ibmcloudp.Provider())
	if err != nil {
		return err
	}
	return ibmcloudp.Destroy(mCtx, stackIBMPowerVS)
}

func (r *pwRequest) deploy(ctx *pulumi.Context) (err error) {
	rg, err := ibmcloud.NewResourceGroup(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "rg"),
		&ibmcloud.ResourceGroupArgs{
			Name: pulumi.String(r.mCtx.ProjectName()),
		})
	if err != nil {
		return err
	}
	w, err := ibmcloud.NewPiWorkspace(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "piw"),
		&ibmcloud.PiWorkspaceArgs{
			PiName:            pulumi.String(r.mCtx.ProjectName()),
			PiDatacenter:      pulumi.String(*r.zone),
			PiResourceGroupId: rg.ID(),
		})
	if err != nil {
		return err
	}
	n, err := ibmcloud.NewPiNetwork(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "pin"),
		&ibmcloud.PiNetworkArgs{
			PiNetworkName:     pulumi.String(r.mCtx.ProjectName()),
			PiCloudInstanceId: w.ID(),
			PiNetworkType:     pulumi.String("vlan"),
			PiCidr:            pulumi.String(NCidr),
			PiGateway:         pulumi.String(NGW),
			// PiDnsServers: pulumi.StringArray{
			// 	pulumi.String("9.9.9.9"),
			// 	pulumi.String("1.1.1.1"),
			// },
		})
	if err != nil {
		return err
	}
	pk, pki, err := piKey(ctx, r.mCtx, *r.prefix, stackIBMPowerVS, w)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		pk.PrivateKeyPem)

	w.ID().ApplyT(func(workspaceId string) error {
		imageId, err := icdata.GetImage(r.mCtx,
			&icdata.PiImageArgs{
				CloudInstanceId: workspaceId,
				Name:            "RHEL9-SP6",
			})
		if err != nil {
			return err
		}

		_, err = ibmcloud.NewPiInstance(ctx,
			resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "pii"),
			&ibmcloud.PiInstanceArgs{
				PiInstanceName:    pulumi.String(r.mCtx.ProjectName()),
				PiMemory:          pulumi.Float64(256),
				PiProcessors:      pulumi.Float64(8),
				PiProcType:        pulumi.String("shared"),
				PiSysType:         pulumi.String("s922"),
				PiImageId:         pulumi.String(*imageId),
				PiHealthStatus:    pulumi.String("WARNING"),
				PiCloudInstanceId: pulumi.String(workspaceId),
				PiStorageType:     pulumi.String("tier1"),
				PiKeyPairName:     pki.PiKeyName,
				PiNetworks: ibmcloud.PiInstancePiNetworkArray{
					&ibmcloud.PiInstancePiNetworkArgs{
						NetworkId: n.NetworkId,
					},
				},
			})
		if err != nil {
			return err
		}
		return nil
	})

	// return o.ApplyT(func(v GetIsVpcsVpcDn) []GetIsVpcsVpcDnResolver { return v.Resolvers }).(GetIsVpcsVpcDnResolverArrayOutput)

	// if err != nil {
	// 	return err
	// }

	// _, err = ibmcloud.NewTgGateway(
	// 	ctx,
	// 	resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "tg"),
	// 	&ibmcloud.TgGatewayArgs{
	// 		Location: pulumi.String(*r.location),
	// 	})

	// return err
	return
}

// func powerArgs(cloudInstanceId, serverName, networkId, pkiName string) *power.PowerArgs {
// 	var memory float64 = 4
// 	processors := 0.25
// 	procType := "shared"
// 	imageId := "f7961557-7fe9-480a-b26a-0a62f8eaa4b7"
// 	return &power.PowerArgs{
// 		CloudInstanceId: cloudInstanceId,
// 		InstanceArgs: models.PVMInstanceCreate{
// 			ServerName:  &serverName,
// 			Memory:      &memory,
// 			Processors:  &processors,
// 			ProcType:    &procType,
// 			SysType:     "s922",
// 			ImageID:     &imageId,
// 			KeyPairName: pkiName,
// 			NetworkIDs:  []string{networkId},
// 		},
// 	}
// }

// // Create a Pi Key
func piKey(ctx *pulumi.Context, mCtx *mc.Context, prefix, cId string, w *ibmcloud.PiWorkspace) (*tls.PrivateKey, *ibmcloud.PiKey, error) {
	pk, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pk"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return nil, nil, err
	}
	if mCtx.Debug() {
		pk.PrivateKeyPem.ApplyT(
			func(privateKey string) error {
				logging.Debugf("%s", privateKey)
				return nil
			})
	}
	pik, err := ibmcloud.NewPiKey(ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pik"),
		&ibmcloud.PiKeyArgs{
			PiKeyName:         pulumi.String(mCtx.ProjectName()),
			PiCloudInstanceId: w.ID(),
			PiSshKey:          pk.PublicKeyOpenssh,
		})
	return pk, pik, err
}

// func publicAccess(ctx *pulumi.Context, prefix, cId string) error {
// 	vpc, err := ibmcloud.NewIsVpc(ctx,
// 		resourcesUtil.GetResourceName(prefix, cId, "isvpc"),
// 		&ibmcloud.IsVpcArgs{
// 			// Name:          pulumi.String(args.Name),
// 			// ResourceGroup: args.RG.ID(),
// 		})
// 	if err != nil {
// 		return err
// 	}

// 	// 	resource "ibm_is_vpc" "main_vpc" {
// 	//   name = "demo-vpc"
// 	// }

// }

// func convert(s *power.PowerArgs) *ibmcloud.PiInstanceArgs {
// 	return &ibmcloud.PiInstanceArgs{
// 		PiInstanceName:    pulumi.String(*s.InstanceArgs.ServerName),
// 		PiMemory:          pulumi.Float64(*s.InstanceArgs.Memory),
// 		PiProcessors:      pulumi.Float64(*s.InstanceArgs.Processors),
// 		PiProcType:        pulumi.String(*s.InstanceArgs.ProcType),
// 		PiSysType:         pulumi.String(s.InstanceArgs.SysType),
// 		PiImageId:         pulumi.String(*s.InstanceArgs.ImageID),
// 		PiCloudInstanceId: pulumi.String(s.CloudInstanceId),
// 		PiKeyPairName:     pulumi.String(s.InstanceArgs.KeyPairName),
// 		PiNetworks: ibmcloud.PiInstancePiNetworkArray{
// 			&ibmcloud.PiInstancePiNetworkArgs{
// 				NetworkId: pulumi.String(s.InstanceArgs.NetworkIDs[0]),
// 			},
// 		},
// 	}
// }

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", prefix, outputHost):           "host",
	})
}
