package ibmz

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmcloudp "github.com/redhat-developer/mapt/pkg/provider/ibmcloud"
	icdata "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/data"
	"github.com/redhat-developer/mapt/pkg/provider/ibmcloud/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

//go:embed cloud-config
var CloudConfig []byte

// otelColVersion is overridden at build time via -ldflags.
var otelColVersion = "0.151.0"

type userDataValues struct {
	AppCode        string
	OtelAuthToken  string
	OtelEndpoint   string
	OtelColVersion string
	OtelIndex      string
	OtelArch       string
	OtelExtraAttrs map[string]string
}

const (
	stackIBMS390         = "ics390"
	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"

	defaultUser = "ubuntu"
)

type ZArgs struct {
	Prefix string
	// SubnetID, when set, deploys the instance into an existing VPC subnet
	// instead of provisioning a new VPC and subnet. IC_ZONE is not required
	// when this field is provided.
	SubnetID string
	// OtelAppCode, OtelAuthToken, and OtelEndpoint are optional. When AppCode
	// and AuthToken are both set, the otelcol-contrib filelog collector is
	// installed and started, shipping logs to OtelEndpoint.
	OtelAppCode    string
	OtelAuthToken  string
	OtelEndpoint   string
	OtelIndex      string
	OtelExtraAttrs map[string]string
}

type zRequest struct {
	mCtx          *mc.Context
	prefix        *string
	zone          *string
	subnetID      *string
	otelAppCode    string
	otelAuthToken  string
	otelEndpoint   string
	otelIndex      string
	otelExtraAttrs map[string]string
}

// New provisions an IBM Z (s390x) VPC instance. When SubnetID is set the
// instance is placed in the existing subnet; otherwise a new VPC, subnet,
// and gateway are created using the IC_ZONE environment variable.
func New(ctx *mc.ContextArgs, args *ZArgs) error {
	ibmcloudProvider := ibmcloudp.Provider()
	mCtx, err := mc.Init(ctx, ibmcloudProvider)
	if err != nil {
		return err
	}

	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")

	var zone *string
	var subnetID *string
	if args.SubnetID != "" {
		s := strings.TrimSpace(args.SubnetID)
		if s == "" {
			return fmt.Errorf("--subnet-id must not be blank")
		}
		subnetID = &s
	} else {
		z, err := ibmcloudProvider.Zone()
		if err != nil {
			return err
		}
		zone = z
	}

	r := &zRequest{
		mCtx:          mCtx,
		prefix:        &prefix,
		zone:          zone,
		subnetID:      subnetID,
		otelAppCode:    args.OtelAppCode,
		otelAuthToken:  args.OtelAuthToken,
		otelEndpoint:   args.OtelEndpoint,
		otelIndex:      args.OtelIndex,
		otelExtraAttrs: args.OtelExtraAttrs,
	}
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

// Destroy tears down the IBM Z VPC stack identified by mCtxArgs.
func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, ibmcloudp.Provider())
	if err != nil {
		return err
	}
	if err := ibmcloudp.DestroyStack(mCtx, stackIBMS390); err != nil {
		return err
	}
	return ibmcloudp.CleanupState(mCtx)
}

func (r *zRequest) deploy(ctx *pulumi.Context) error {
	if r.subnetID != nil {
		return r.deployWithExistingSubnet(ctx)
	}
	zone := *r.zone
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
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), pk.PrivateKeyPem)
	imageId, err := icdata.GetVPCImage(&icdata.VPCImageArgs{
		Name: "ibm-ubuntu-22-04",
		Arch: icdata.VPC_ARCH_IBMZ,
	})
	if err != nil {
		return err
	}
	instanceArgs := &ibmcloud.IsInstanceArgs{
		Name:          pulumi.String(r.mCtx.ProjectName()),
		Image:         pulumi.String(*imageId),
		Profile:       pulumi.String("bz2-16x64"),
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
	}
	if r.otelAppCode != "" && r.otelAuthToken != "" {
		ud, err := izUserData(r.otelAppCode, r.otelAuthToken, r.otelEndpoint, r.otelIndex, r.otelExtraAttrs)
		if err != nil {
			return fmt.Errorf("failed to render user data: %w", err)
		}
		instanceArgs.UserData = pulumi.StringPtr(ud)
	}
	// https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui&q=s390x&tags=vpc
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "is"),
		instanceArgs)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(defaultUser))
	_, err = ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "fipassoc"),
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

// deployWithExistingSubnet places the instance in a pre-existing VPC subnet.
// VPC ID and zone are resolved from the subnet lookup; a new security group
// and floating IP are created in that VPC. No VPC, subnet, or gateway
// resources are provisioned.
func (r *zRequest) deployWithExistingSubnet(ctx *pulumi.Context) error {
	subnetInfo, err := ibmcloud.LookupIsSubnet(ctx, &ibmcloud.LookupIsSubnetArgs{
		Identifier: r.subnetID,
	})
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s-%s", *r.prefix, r.mCtx.ProjectName())
	sg, err := network.NewSecurityGroupWithSSH(ctx, &network.SecurityGroupArgs{
		Prefix:      *r.prefix,
		ComponentID: stackIBMS390,
		Name:        name,
		VPC:         pulumi.String(subnetInfo.Vpc),
	})
	if err != nil {
		return err
	}
	fip, err := network.NewFloatingIP(ctx, &network.FloatingIPArgs{
		Prefix:      *r.prefix,
		ComponentID: stackIBMS390,
		Name:        name,
		Zone:        pulumi.String(subnetInfo.Zone),
	})
	if err != nil {
		return err
	}
	// rg is nil: the SSH key is placed in the account default resource group.
	pk, pik, err := isKey(ctx, r.mCtx, *r.prefix, stackIBMS390, nil)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), pk.PrivateKeyPem)
	imageId, err := icdata.GetVPCImage(&icdata.VPCImageArgs{
		Name: "ibm-ubuntu-22-04",
		Arch: icdata.VPC_ARCH_IBMZ,
	})
	if err != nil {
		return err
	}
	existingSubnetInstanceArgs := &ibmcloud.IsInstanceArgs{
		Name:    pulumi.String(r.mCtx.ProjectName()),
		Image:   pulumi.String(*imageId),
		Profile: pulumi.String("bz2-16x64"),
		Vpc:     pulumi.String(subnetInfo.Vpc),
		Zone:    pulumi.String(subnetInfo.Zone),
		Keys:    pulumi.StringArray{pik.ID()},
		PrimaryNetworkInterface: &ibmcloud.IsInstancePrimaryNetworkInterfaceArgs{
			Subnet:         pulumi.String(*r.subnetID),
			SecurityGroups: pulumi.StringArray{sg.ID()},
		},
	}
	if r.otelAppCode != "" && r.otelAuthToken != "" {
		ud, err := izUserData(r.otelAppCode, r.otelAuthToken, r.otelEndpoint, r.otelIndex, r.otelExtraAttrs)
		if err != nil {
			return fmt.Errorf("failed to render user data: %w", err)
		}
		existingSubnetInstanceArgs.UserData = pulumi.StringPtr(ud)
	}
	// https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui&q=s390x&tags=vpc
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "is"),
		existingSubnetInstanceArgs)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(defaultUser))
	_, err = ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "fipassoc"),
		&ibmcloud.IsInstanceNetworkInterfaceFloatingIpArgs{
			FloatingIp: fip.ID(),
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
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), fip.Address)
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "readiness-cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:       fip.Address,
				User:       pulumi.String(defaultUser),
				PrivateKey: pk.PrivateKeyOpenssh,
			},
			Create: pulumi.String(command.CommandPing),
			Update: pulumi.String(command.CommandPing),
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: command.RemoteTimeout,
				Update: command.RemoteTimeout}),
		pulumi.DependsOn([]pulumi.Resource{i}))
	return err
}

func izUserData(otelAppCode, otelAuthToken, otelEndpoint, otelIndex string, otelExtraAttrs map[string]string) (string, error) {
	script, err := file.Template(
		userDataValues{
			AppCode:        otelAppCode,
			OtelAuthToken:  otelAuthToken,
			OtelEndpoint:   otelEndpoint,
			OtelColVersion: otelColVersion,
			OtelIndex:      otelIndex,
			OtelArch:       "s390x",
			OtelExtraAttrs: otelExtraAttrs,
		},
		string(CloudConfig))
	if err != nil {
		return "", err
	}
	// IBM Cloud VPC passes user_data as-is to cloud-init, so we cannot
	// send a bare base64 string (cloud-init won't recognise it). Wrapping in
	// a MIME multipart envelope with Content-Transfer-Encoding: base64 tells
	// cloud-init to decode the payload before processing.
	const boundary = "MAPT-CLOUD-CONFIG"
	encoded := base64.StdEncoding.EncodeToString([]byte(script))
	return strings.Join([]string{
		"MIME-Version: 1.0",
		`Content-Type: multipart/mixed; boundary="` + boundary + `"`,
		"",
		"--" + boundary,
		`Content-Type: text/cloud-config; charset="us-ascii"`,
		"Content-Transfer-Encoding: base64",
		"",
		encoded,
		"--" + boundary + "--",
		"",
	}, "\n"), nil
}

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string) error {
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", prefix, outputHost):           "host",
	})
}

// isKey creates a 4096-bit RSA TLS key pair and registers the public key as
// an IBM Cloud VPC SSH key. Pass rg=nil to place the key in the account
// default resource group.
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
	sshKeyArgs := &ibmcloud.IsSshKeyArgs{
		Name:      pulumi.String(mCtx.ProjectName()),
		PublicKey: pk.PublicKeyOpenssh,
	}
	if rg != nil {
		sshKeyArgs.ResourceGroup = rg.ID()
	}
	pik, err := ibmcloud.NewIsSshKey(ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pik"),
		sshKeyArgs)
	return pk, pik, err
}
