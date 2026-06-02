package ibmpower

import (
	_ "embed"
	"encoding/base64"
	"fmt"

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
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

//go:embed cloud-config
var CloudConfig []byte

// otelColVersion is overridden at build time via -ldflags.
var otelColVersion = "0.151.0"

type userDataValues struct {
	Gateway        string
	AppCode        string
	OtelAuthToken  string
	OtelEndpoint   string
	OtelColVersion string
	OtelIndex      string
	OtelArch       string
	OtelExtraAttrs map[string]string
}

const (
	stackIBMPowerVS      = "icpw"
	bastionComponentID   = "icpw-bst"
	outputHost           = "alsHost"
	outputUsername       = "alsUsername"
	outputUserPrivateKey = "alsUserPrivatekey"

	outputBastionHost           = "alsBastionHost"
	outputBastionUsername       = "alsBastionUsername"
	outputBastionUserPrivateKey = "alsBastionUserPrivatekey"

	imageRHEL9   = "RHEL9-SP6"
	defaultUser  = "root"

	bastionUser    = "ubuntu"
	bastionProfile = "cx2-2x4"
	bastionImage   = "ibm-ubuntu-24-04"

	// Standard large build-host sizing on an s1022 (Power10) frame with shared processors and tier1 SSD.
	instanceMemory      = 256.0
	instanceProcs       = 8.0
	instanceProcType    = "shared"
	instanceSysType     = "s1022"
	instanceStorageType = "tier1"
)

type PWArgs struct {
	Prefix            string
	PIPrivateSubnetID string
	WorkspaceID       string
	// VPCPublicSubnetID is optional. When set, a small VPC bastion instance
	// with a floating IP is created in this subnet to provide SSH access to
	// the PowerVS instance over the Transit Gateway private network.
	VPCPublicSubnetID string
	// OtelAppCode, OtelAuthToken, and OtelEndpoint are optional. When AppCode
	// and AuthToken are both set, the otelcol-contrib filelog collector is
	// installed and started, shipping logs to OtelEndpoint.
	OtelAppCode    string
	OtelAuthToken  string
	OtelEndpoint   string
	OtelIndex      string
	OtelExtraAttrs map[string]string
}

type pwRequest struct {
	mCtx              *mc.Context
	prefix            *string
	piPrivateSubnetID string
	workspaceID       string
	vpcPublicSubnetID string
	otelAppCode       string
	otelAuthToken     string
	otelEndpoint      string
	otelIndex         string
	otelExtraAttrs    map[string]string
}

// New provisions a Power VS (ppc64) instance inside an existing workspace and
// network. Both NetworkID and WorkspaceID are required. When VPCSubnetID is
// set, a VPC bastion with a floating IP is also created for SSH access.
func New(ctx *mc.ContextArgs, args *PWArgs) error {
	if args.PIPrivateSubnetID == "" || args.WorkspaceID == "" {
		return fmt.Errorf("--pi-private-subnet-id and --workspace-id are required")
	}

	ibmcloudProvider := ibmcloudp.Provider()
	mCtx, err := mc.Init(ctx, ibmcloudProvider)
	if err != nil {
		return err
	}

	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &pwRequest{
		mCtx:              mCtx,
		prefix:            &prefix,
		piPrivateSubnetID: args.PIPrivateSubnetID,
		workspaceID:       args.WorkspaceID,
		vpcPublicSubnetID: args.VPCPublicSubnetID,
		otelAppCode:       args.OtelAppCode,
		otelAuthToken:     args.OtelAuthToken,
		otelEndpoint:      args.OtelEndpoint,
		otelIndex:         args.OtelIndex,
		otelExtraAttrs:    args.OtelExtraAttrs,
	}
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
	return manageResults(mCtx, sr, prefix, r.vpcPublicSubnetID != "")
}

// Destroy tears down the Power VS stack identified by mCtxArgs.
func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, ibmcloudp.Provider())
	if err != nil {
		return err
	}
	if err := ibmcloudp.DestroyStack(mCtx, stackIBMPowerVS); err != nil {
		return err
	}
	return ibmcloudp.CleanupState(mCtx)
}

func (r *pwRequest) deploy(ctx *pulumi.Context) error {
	pk, pki, err := piKey(ctx, r.mCtx, *r.prefix, stackIBMPowerVS, pulumi.String(r.workspaceID))
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), pk.PrivateKeyPem)

	subnetInfo, err := ibmcloud.LookupPiNetwork(ctx, &ibmcloud.LookupPiNetworkArgs{
		PiCloudInstanceId: r.workspaceID,
		PiNetworkId:       &r.piPrivateSubnetID,
	})
	if err != nil {
		return fmt.Errorf("failed to look up private subnet: %w", err)
	}

	userData, err := piUserData(subnetInfo.Gateway, r.otelAppCode, r.otelAuthToken, r.otelEndpoint, r.otelIndex, r.otelExtraAttrs)
	if err != nil {
		return fmt.Errorf("failed to render user data: %w", err)
	}

	imageId, err := icdata.GetImage(r.mCtx,
		&icdata.PiImageArgs{
			CloudInstanceId: r.workspaceID,
			Name:            imageRHEL9,
		})
	if err != nil {
		return err
	}

	i, err := ibmcloud.NewPiInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "pii"),
		&ibmcloud.PiInstanceArgs{
			PiInstanceName:    pulumi.String(r.mCtx.ProjectName()),
			PiMemory:          pulumi.Float64(instanceMemory),
			PiProcessors:      pulumi.Float64(instanceProcs),
			PiProcType:        pulumi.String(instanceProcType),
			PiSysType:         pulumi.String(instanceSysType),
			PiImageId:         pulumi.String(*imageId),
			PiHealthStatus:    pulumi.String("WARNING"),
			PiCloudInstanceId: pulumi.String(r.workspaceID),
			PiStorageType:     pulumi.String(instanceStorageType),
			PiKeyPairName:     pki.PiKeyName,
			PiUserData:        pulumi.StringPtr(userData),
			PiNetworks: ibmcloud.PiInstancePiNetworkArray{
				&ibmcloud.PiInstancePiNetworkArgs{
					NetworkId: pulumi.String(r.piPrivateSubnetID),
				},
			},
		})
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(defaultUser))
	// Use ExternalIp when available (pub-vlan network); fall back to IpAddress for private networks.
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		i.PiNetworks.ApplyT(func(networks []ibmcloud.PiInstancePiNetwork) (string, error) {
			if len(networks) == 0 {
				return "", fmt.Errorf("instance has no network interfaces")
			}
			if networks[0].ExternalIp != nil && *networks[0].ExternalIp != "" {
				return *networks[0].ExternalIp, nil
			}
			if networks[0].IpAddress != nil && *networks[0].IpAddress != "" {
				return *networks[0].IpAddress, nil
			}
			return "", fmt.Errorf("instance network has no IP address")
		}).(pulumi.StringOutput))

	if r.vpcPublicSubnetID != "" {
		return r.deployBastion(ctx)
	}
	return nil
}

// deployBastion creates a small VPC instance with a floating IP in the
// provided subnet. It acts as an SSH jump host to reach the PowerVS instance
// over the Transit Gateway private network.
func (r *pwRequest) deployBastion(ctx *pulumi.Context) error {
	subnetInfo, err := ibmcloud.LookupIsSubnet(ctx, &ibmcloud.LookupIsSubnetArgs{
		Identifier: &r.vpcPublicSubnetID,
	})
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s-%s-bastion", *r.prefix, r.mCtx.ProjectName())

	sg, err := network.NewSecurityGroupWithSSH(ctx, &network.SecurityGroupArgs{
		Prefix:      *r.prefix,
		ComponentID: bastionComponentID,
		Name:        name,
		VPC:         pulumi.String(subnetInfo.Vpc),
	})
	if err != nil {
		return err
	}

	bpk, err := tls.NewPrivateKey(ctx,
		resourcesUtil.GetResourceName(*r.prefix, bastionComponentID, "pk"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}

	bsshKey, err := ibmcloud.NewIsSshKey(ctx,
		resourcesUtil.GetResourceName(*r.prefix, bastionComponentID, "pik"),
		&ibmcloud.IsSshKeyArgs{
			Name:      pulumi.String(name),
			PublicKey: bpk.PublicKeyOpenssh,
		})
	if err != nil {
		return err
	}

	bastionImageId, err := icdata.GetVPCImage(&icdata.VPCImageArgs{
		Name: bastionImage,
		Arch: icdata.VPC_ARCH_X86_64,
	})
	if err != nil {
		return err
	}

	bastion, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, bastionComponentID, "is"),
		&ibmcloud.IsInstanceArgs{
			Name:    pulumi.String(name),
			Image:   pulumi.String(*bastionImageId),
			Profile: pulumi.String(bastionProfile),
			Vpc:     pulumi.String(subnetInfo.Vpc),
			Zone:    pulumi.String(subnetInfo.Zone),
			Keys:    pulumi.StringArray{bsshKey.ID()},
			PrimaryNetworkInterface: &ibmcloud.IsInstancePrimaryNetworkInterfaceArgs{
				Subnet:         pulumi.String(r.vpcPublicSubnetID),
				SecurityGroups: pulumi.StringArray{sg.ID()},
			},
		})
	if err != nil {
		return err
	}

	fip, err := network.NewFloatingIP(ctx, &network.FloatingIPArgs{
		Prefix:      *r.prefix,
		ComponentID: bastionComponentID,
		Name:        name,
		Zone:        pulumi.String(subnetInfo.Zone),
	})
	if err != nil {
		return err
	}

	_, err = ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, bastionComponentID, "fipassoc"),
		&ibmcloud.IsInstanceNetworkInterfaceFloatingIpArgs{
			FloatingIp: fip.ID(),
			Instance:   bastion.ID(),
			NetworkInterface: bastion.PrimaryNetworkInterface.ApplyT(
				func(pni ibmcloud.IsInstancePrimaryNetworkInterface) string {
					return *pni.Id
				},
			).(pulumi.StringOutput),
		})
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputBastionHost), fip.Address)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputBastionUsername), pulumi.String(bastionUser))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputBastionUserPrivateKey), bpk.PrivateKeyPem)
	return nil
}

// piKey creates a 4096-bit RSA TLS key pair and registers the public key as a
// Power VS SSH key in the given workspace.
func piKey(ctx *pulumi.Context, mCtx *mc.Context, prefix, cId string, cloudInstanceID pulumi.StringInput) (*tls.PrivateKey, *ibmcloud.PiKey, error) {
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
			PiCloudInstanceId: cloudInstanceID,
			PiSshKey:          pk.PublicKeyOpenssh,
		})
	return pk, pik, err
}

// piUserData renders the cloud-config template and returns it base64-encoded
// for use as PiUserData on a PowerVS instance.
func piUserData(gateway, otelAppCode, otelAuthToken, otelEndpoint, otelIndex string, otelExtraAttrs map[string]string) (string, error) {
	script, err := file.Template(
		userDataValues{
			Gateway:        gateway,
			AppCode:        otelAppCode,
			OtelAuthToken:  otelAuthToken,
			OtelEndpoint:   otelEndpoint,
			OtelColVersion: otelColVersion,
			OtelIndex:      otelIndex,
			OtelArch:       "ppc64le",
			OtelExtraAttrs: otelExtraAttrs,
		},
		string(CloudConfig))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(script)), nil
}

func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix string, withBastion bool) error {
	outputMap := map[string]string{
		fmt.Sprintf("%s-%s", prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", prefix, outputHost):           "host",
	}
	if withBastion {
		outputMap[fmt.Sprintf("%s-%s", prefix, outputBastionHost)]           = "bastion_host"
		outputMap[fmt.Sprintf("%s-%s", prefix, outputBastionUsername)]       = "bastion_username"
		outputMap[fmt.Sprintf("%s-%s", prefix, outputBastionUserPrivateKey)] = "bastion_id_rsa"
	}
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), outputMap)
}
