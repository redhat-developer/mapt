package ibmpower

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	"github.com/redhat-developer/mapt/pkg/integrations/otelcol"
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

type userDataValues struct {
	Gateway                string
	OtelColScript          string
	GitLabRunnerScript     string
	GHActionsRunnerScript  string
	COSAccessKeyID         string
	COSSecretAccessKey     string
	COSEndpoint            string
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

)

type PWArgs struct {
	Prefix            string
	PIPrivateSubnetID string
	WorkspaceID       string
	// VPCPublicSubnetID is optional. When set, a small VPC bastion instance
	// with a floating IP is created in this subnet to provide SSH access to
	// the PowerVS instance over the Transit Gateway private network.
	VPCPublicSubnetID string
	// Instance sizing
	Memory      float64
	Processors  float64
	ProcType    string
	SysType     string
	StorageType string
	DiskSize    int
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
	memory            float64
	processors        float64
	procType          string
	sysType           string
	storageType       string
	diskSize          int
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

	sysTypes, err := icdata.GetAvailableSystemTypes(mCtx, &icdata.SystemTypeRequirements{
		CloudInstanceId: args.WorkspaceID,
		Zone:            os.Getenv("IC_ZONE"),
		ProcType:        args.ProcType,
		PreferredType:   args.SysType,
	})
	if err != nil {
		return fmt.Errorf("system type discovery failed: %w", err)
	}

	var lastErr error
	for i, sysType := range sysTypes.Types {
		if i > 0 {
			logging.Warnf("retrying with system type %s (%d/%d) after capacity failure",
				sysType, i+1, len(sysTypes.Types))
		}

		r := &pwRequest{
			mCtx:              mCtx,
			prefix:            &prefix,
			piPrivateSubnetID: args.PIPrivateSubnetID,
			workspaceID:       args.WorkspaceID,
			vpcPublicSubnetID: args.VPCPublicSubnetID,
			memory:            args.Memory,
			processors:        args.Processors,
			procType:          args.ProcType,
			sysType:           sysType,
			storageType:       args.StorageType,
			diskSize:          args.DiskSize,
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
		if err == nil {
			if i > 0 {
				logging.Infof("provisioning succeeded with system type %s (attempt %d)", sysType, i+1)
			}
			return manageResults(mCtx, sr, prefix, r.vpcPublicSubnetID != "")
		}

		lastErr = err
		if !isCapacityError(err) {
			return fmt.Errorf("stack creation failed: %w", err)
		}

		logging.Warnf("capacity error with system type %s: %v", sysType, err)

		if i < len(sysTypes.Types)-1 {
			logging.Infof("destroying partial stack before retry...")
			if dErr := destroyForRetry(mCtx); dErr != nil {
				logging.Warnf("failed to destroy partial stack: %v", dErr)
			}
		}
	}

	return fmt.Errorf("all system types exhausted; last error: %w", lastErr)
}

func isCapacityError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	for _, pattern := range []string{
		"insufficient resources",
		"no available host",
		"capacity is not available",
		"not enough resources",
		"resource capacity",
		"no hosts available",
		"maximum capacity",
	} {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

func destroyForRetry(mCtx *mc.Context) error {
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackIBMPowerVS),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: ibmcloudp.DefaultCredentials,
	}
	return manager.DestroyStack(mCtx, cs)
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

	otelSet := 0
	for _, f := range []string{r.otelAppCode, r.otelAuthToken, r.otelIndex} {
		if f != "" {
			otelSet++
		}
	}
	if otelSet > 0 && otelSet < 3 {
		return fmt.Errorf("partial otel configuration: --otel-app-code, --otel-auth-token, and --otel-index must all be set together")
	}
	hasOtel := otelSet == 3

	ghRunnerScript := ""
	if ghRunnerArgs := github.GetRunnerArgs(); ghRunnerArgs != nil {
		s, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(ghRunnerArgs, defaultUser)
		if err != nil {
			return err
		}
		ghRunnerScript = *s
	}

	var piUserDataInput pulumi.StringPtrInput
	glRunnerArgs := gitlab.GetRunnerArgs()
	if glRunnerArgs != nil {
		authToken, err := gitlab.CreateRunner(ctx, glRunnerArgs)
		if err != nil {
			return err
		}
		gateway := subnetInfo.Gateway
		localArgs := *glRunnerArgs
		localGHScript := ghRunnerScript
		piUserDataInput = authToken.ApplyT(func(token string) (*string, error) {
			localArgs.AuthToken = token
			glSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(&localArgs, defaultUser)
			if err != nil {
				return nil, err
			}
			var otelArgs *otelcol.OtelcolArgs
			if hasOtel {
				otelArgs = r.otelArgs(true)
			}
			ud, err := piUserData(gateway, otelArgs, *glSnippet, localGHScript)
			if err != nil {
				return nil, err
			}
			return &ud, nil
		}).(pulumi.StringPtrOutput)
	} else {
		var otelArgs *otelcol.OtelcolArgs
		if hasOtel {
			otelArgs = r.otelArgs(false)
		}
		ud, err := piUserData(subnetInfo.Gateway, otelArgs, "", ghRunnerScript)
		if err != nil {
			return fmt.Errorf("failed to render user data: %w", err)
		}
		piUserDataInput = pulumi.StringPtr(ud)
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
			PiMemory:          pulumi.Float64(r.memory),
			PiProcessors:      pulumi.Float64(r.processors),
			PiProcType:        pulumi.String(r.procType),
			PiSysType:         pulumi.String(r.sysType),
			PiImageId:         pulumi.String(*imageId),
			PiHealthStatus:    pulumi.String("WARNING"),
			PiCloudInstanceId: pulumi.String(r.workspaceID),
			PiStorageType:     pulumi.String(r.storageType),
			PiKeyPairName:     pki.PiKeyName,
			PiUserData:        piUserDataInput,
			PiNetworks: ibmcloud.PiInstancePiNetworkArray{
				&ibmcloud.PiInstancePiNetworkArgs{
					NetworkId: pulumi.String(r.piPrivateSubnetID),
				},
			},
		})
	if err != nil {
		return err
	}

	// Both i.ID() and piv.ID() return "cloudInstanceId/resourceId" — extract just the resource ID
	splitID := func(id string) (string, error) {
		if parts := strings.SplitN(id, "/", 2); len(parts) == 2 {
			return parts[1], nil
		}
		return id, nil
	}
	piInstanceId := i.ID().ApplyT(splitID).(pulumi.StringOutput)
	piv, err := ibmcloud.NewPiVolume(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "piv"),
		&ibmcloud.PiVolumeArgs{
			PiCloudInstanceId:  pulumi.String(r.workspaceID),
			PiVolumeName:       pulumi.String(r.mCtx.ProjectName()),
			PiVolumeSize:       pulumi.Float64(float64(r.diskSize)),
			PiVolumeType:       pulumi.String(r.storageType),
			PiVolumeShareable:  pulumi.Bool(false),
			PiAffinityPolicy:   pulumi.String("affinity"),
			PiAffinityInstance: piInstanceId.ToStringPtrOutput(),
		})
	if err != nil {
		return err
	}
	pivVolumeId := piv.ID().ApplyT(splitID).(pulumi.StringOutput)
	_, err = ibmcloud.NewPiVolumeAttach(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMPowerVS, "piva"),
		&ibmcloud.PiVolumeAttachArgs{
			PiCloudInstanceId: pulumi.String(r.workspaceID),
			PiInstanceId:      piInstanceId,
			PiVolumeId:        pivVolumeId,
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

func (r *pwRequest) otelArgs(monitorGitLabRunner bool) *otelcol.OtelcolArgs {
	return &otelcol.OtelcolArgs{
		AppCode:             r.otelAppCode,
		AuthToken:           r.otelAuthToken,
		Index:               r.otelIndex,
		Endpoint:            r.otelEndpoint,
		Arch:                otelcol.Ppc64le,
		SyslogPath:          "/var/log/messages",
		SecurePath:          "/var/log/secure",
		ExtraAttrs:          r.otelExtraAttrs,
		MonitorGitLabRunner: monitorGitLabRunner,
	}
}

// piUserData renders the cloud-config template and returns it base64-encoded
// for use as PiUserData on a PowerVS instance.
func piUserData(gateway string, otelArgs *otelcol.OtelcolArgs, glRunnerScript, ghRunnerScript string) (string, error) {
	otelScript := ""
	if otelArgs != nil {
		s, err := otelcol.GetSnippetAsCloudInitWritableFile(otelArgs)
		if err != nil {
			return "", err
		}
		otelScript = *s
	}
	script, err := file.Template(
		userDataValues{
			Gateway:               gateway,
			OtelColScript:         otelScript,
			GitLabRunnerScript:    glRunnerScript,
			GHActionsRunnerScript: ghRunnerScript,
			COSAccessKeyID:        os.Getenv("IBMCLOUD_COS_ACCESS_KEY_ID"),
			COSSecretAccessKey:    os.Getenv("IBMCLOUD_COS_SECRET_ACCESS_KEY"),
			COSEndpoint:           os.Getenv("IBMCLOUD_COS_ENDPOINT"),
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
