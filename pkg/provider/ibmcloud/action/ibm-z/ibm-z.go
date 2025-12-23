package ibmz

import (
	"fmt"
	"os"

	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmcloudp "github.com/redhat-developer/mapt/pkg/provider/ibmcloud"
	icdata "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/data"
	"github.com/redhat-developer/mapt/pkg/provider/ibmcloud/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	stackIBMS390         = "ics390"
	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"

	NCidr = "10.0.2.0/24"

	defaultUser = "ubuntu"
)

type ZArgs struct {
	Prefix string
}

type zRequest struct {
	mCtx   *mc.Context
	prefix *string
	zone   *string
}

func New(ctx *mc.ContextArgs, args *ZArgs) error {
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
	r := &zRequest{
		mCtx:   mCtx,
		prefix: &prefix,
		zone:   zone}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackIBMS390),
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
	return ibmcloudp.Destroy(mCtx, stackIBMS390)
}

func (r *zRequest) deploy(ctx *pulumi.Context) error {
	zone := os.Getenv("IC_ZONE")
	rg, err := ibmcloud.NewResourceGroup(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "rg"),
		&ibmcloud.ResourceGroupArgs{
			Name: pulumi.String(r.mCtx.ProjectName()),
		})
	if err != nil {
		return err
	}
	n, err := network.New(ctx,
		&network.NetworkArgs{
			Prefix:      *r.prefix,
			Zone:        &zone,
			RG:          rg,
			ComponentID: stackIBMS390,
			Name:        fmt.Sprintf("%s-%s", *r.prefix, r.mCtx.ProjectName()),
		})
	if err != nil {
		return err
	}
	pk, pik, err := isKey(ctx, r.mCtx, *r.prefix, stackIBMS390, rg)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		pk.PrivateKeyPem)
	imageId, err := icdata.GetVPCImage(&icdata.VPCImageArgs{
		Name: "ibm-ubuntu-22-04",
		Arch: icdata.VPC_ARCH_IBMZ,
	})
	if err != nil {
		return err
	}
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "is"),
		&ibmcloud.IsInstanceArgs{
			Name:  pulumi.String(r.mCtx.ProjectName()),
			Image: pulumi.String(*imageId),
			// https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui&q=arm64&tags=vpc#gpu
			Profile:       pulumi.String("bz2-16x64"), //
			Vpc:           n.VPC.ID(),
			Zone:          pulumi.String(zone),
			ResourceGroup: rg.ID(),
			Keys:          pulumi.StringArray{pik.ID()},
			PrimaryNetworkInterface: &ibmcloud.IsInstancePrimaryNetworkInterfaceArgs{
				Subnet: n.Subnet.ID(),
				SecurityGroups: pulumi.StringArray{
					n.SecurityGroup.ID(),
				},
			},
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(defaultUser))
	_, err = ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "fip"),
		&ibmcloud.IsInstanceNetworkInterfaceFloatingIpArgs{
			FloatingIp: n.Floatingip.ID(),
			Instance:   i.ID(),
			NetworkInterface: i.PrimaryNetworkInterface.ApplyT(
				func(pni ibmcloud.IsInstancePrimaryNetworkInterface) string {
					return *pni.Id
				},
			).(pulumi.StringOutput),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), n.Floatingip.Address)
	return nil
}

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", prefix, outputHost):           "host",
	})
}

func isKey(ctx *pulumi.Context, mCtx *mc.Context, prefix, cId string, rg *ibmcloud.ResourceGroup) (*tls.PrivateKey, *ibmcloud.IsSshKey, error) {
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
	pik, err := ibmcloud.NewIsSshKey(ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pik"),
		&ibmcloud.IsSshKeyArgs{
			Name:          pulumi.String(mCtx.ProjectName()),
			ResourceGroup: rg.ID(),
			PublicKey:     pk.PublicKeyOpenssh,
		})
	return pk, pik, err
}
