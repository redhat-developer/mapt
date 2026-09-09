package kind

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/mapt-oss/pulumi-ibmcloud/sdk/go/ibmcloud"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmcloudp "github.com/redhat-developer/mapt/pkg/provider/ibmcloud"
	icdata "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/data"
	"github.com/redhat-developer/mapt/pkg/provider/ibmcloud/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	utilKind "github.com/redhat-developer/mapt/pkg/target/service/kind"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

//go:embed cloud-config
var cloudConfigTemplate []byte

type kindRequest struct {
	mCtx              *mc.Context
	prefix            *string
	version           *string
	zone              *string
	profile           string
	diskSize          int
	spot              bool
	extraPortMappings []utilKind.PortMapping
}

func Create(mCtxArgs *mc.ContextArgs, args *utilKind.KindArgs) (*utilKind.KindResults, error) {
	ibmcloudProvider := ibmcloudp.Provider()
	mCtx, err := mc.Init(mCtxArgs, ibmcloudProvider)
	if err != nil {
		return nil, err
	}
	zone, err := ibmcloudProvider.Zone()
	if err != nil {
		return nil, err
	}
	if _, ok := utilKind.KindK8sVersions[args.Version]; !ok {
		return nil, fmt.Errorf("unsupported Kubernetes version %q", args.Version)
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	diskSize := defaultDiskSize
	profile := defaultProfile
	if args.ComputeRequest != nil {
		if args.ComputeRequest.DiskSize != nil {
			diskSize = *args.ComputeRequest.DiskSize
		}
		if len(args.ComputeRequest.ComputeSizes) > 0 {
			profile = args.ComputeRequest.ComputeSizes[0]
		} else if args.ComputeRequest.CPUs > 0 || args.ComputeRequest.MemoryGib > 0 {
			profiles, err := icdata.NewComputeSelector().Select(args.ComputeRequest)
			if err != nil {
				return nil, fmt.Errorf("selecting compute profile: %w", err)
			}
			profile = profiles[0]
		}
	}
	spot := args.Spot != nil && args.Spot.Spot
	if spot {
		ok, err := icdata.ProfileSupportsSpot(profile)
		if err != nil {
			return nil, fmt.Errorf("checking spot support for profile %s: %w", profile, err)
		}
		if !ok {
			return nil, fmt.Errorf("profile %s does not support spot instances", profile)
		}
	}
	r := &kindRequest{
		mCtx:              mCtx,
		prefix:            &prefix,
		version:           &args.Version,
		zone:              zone,
		profile:           profile,
		diskSize:          diskSize,
		spot:              spot,
		extraPortMappings: args.ExtraPortMappings,
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackIBMCloudKind),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: ibmcloudp.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, err := manager.UpStack(mCtx, cs)
	if err != nil {
		return nil, fmt.Errorf("stack creation failed: %w", err)
	}
	return utilKind.Results(mCtx, sr, &prefix, r.spot)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	mCtx, err := mc.Init(mCtxArgs, ibmcloudp.Provider())
	if err != nil {
		return err
	}
	if err := ibmcloudp.DestroyStack(mCtx, stackIBMCloudKind); err != nil {
		return err
	}
	return ibmcloudp.CleanupState(mCtx)
}

func (r *kindRequest) deploy(ctx *pulumi.Context) error {
	userTags := ibmcloudp.TagsAsStringArray(r.mCtx.GetTags())
	zone := *r.zone

	rg, err := ibmcloud.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudKindID, "rg"),
		&ibmcloud.ResourceGroupArgs{
			Name: pulumi.String(r.mCtx.ProjectName()),
			Tags: userTags,
		})
	if err != nil {
		return err
	}

	n, err := network.New(ctx, &network.NetworkArgs{
		Prefix:      *r.prefix,
		ComponentID: ibmCloudKindID,
		Name:        fmt.Sprintf("%s-%s", *r.prefix, r.mCtx.ProjectName()),
		Zone:        &zone,
		RG:          rg,
		Tags:        userTags,
	})
	if err != nil {
		return err
	}

	extraHostPorts := make([]int, 0, len(r.extraPortMappings))
	for _, pm := range r.extraPortMappings {
		extraHostPorts = append(extraHostPorts, pm.HostPort)
	}
	if err := addKindSecurityGroupRules(ctx, r.prefix, n.SecurityGroup, extraHostPorts); err != nil {
		return err
	}

	pk, pik, err := isKey(ctx, r.mCtx, *r.prefix, ibmCloudKindID, rg, userTags)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKPrivateKey), pk.PrivateKeyPem)

	imageID, err := icdata.GetVPCImage(&icdata.VPCImageArgs{
		Name: imageName,
		Arch: icdata.VPC_ARCH_X86_64,
	})
	if err != nil {
		return err
	}

	ud, err := userData(r.version, r.extraPortMappings, n.Floatingip.Address)
	if err != nil {
		return err
	}

	instanceArgs := &ibmcloud.IsInstanceArgs{
		Name:    pulumi.String(r.mCtx.ProjectName()),
		Image:   pulumi.String(*imageID),
		Profile: pulumi.String(r.profile),
		Vpc:     n.VPC.ID(),
		Zone:    pulumi.String(zone),
		BootVolume: &ibmcloud.IsInstanceBootVolumeArgs{
			Size: pulumi.Int(r.diskSize),
		},
		ResourceGroup: rg.ID(),
		Keys:          pulumi.StringArray{pik.ID()},
		Tags:          userTags,
		PrimaryNetworkInterface: &ibmcloud.IsInstancePrimaryNetworkInterfaceArgs{
			Subnet: n.Subnet.ID(),
			SecurityGroups: pulumi.StringArray{
				n.SecurityGroup.ID(),
			},
		},
		UserData: ud,
	}
	if r.spot {
		instanceArgs.Availability = ibmcloud.IsInstanceAvailabilityArgs{
			Class: pulumi.String("spot"),
		}
	}
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudKindID, "is"),
		instanceArgs)
	if err != nil {
		return err
	}

	fipAssoc, err := ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudKindID, "fipassoc"),
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

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKUsername), pulumi.String(defaultUser))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKHost), n.Floatingip.Address)

	kc, err := kubeconfig(ctx, r.prefix, n.Floatingip.Address, pk, []pulumi.Resource{fipAssoc})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKKubeconfig), pulumi.ToSecret(kc))

	return nil
}

// addKindSecurityGroupRules adds inbound rules for Kubernetes API, HTTP, HTTPS,
// and any extra host ports to the security group created by network.New.
func addKindSecurityGroupRules(ctx *pulumi.Context, prefix *string, sg *ibmcloud.IsSecurityGroup, extraHostPorts []int) error {
	seen := make(map[int]struct{})
	for _, port := range append([]int{utilKind.PortAPI, utilKind.PortHTTP, utilKind.PortHTTPS}, extraHostPorts...) {
		if _, dup := seen[port]; dup {
			continue
		}
		seen[port] = struct{}{}
		_, err := ibmcloud.NewIsSecurityGroupRule(ctx,
			resourcesUtil.GetResourceName(*prefix, ibmCloudKindID, fmt.Sprintf("sgr%d", port)),
			&ibmcloud.IsSecurityGroupRuleArgs{
				Group:     sg.ID(),
				Direction: pulumi.String("inbound"),
				Remote:    pulumi.String("0.0.0.0/0"),
				Protocol: pulumi.String("tcp"),
				PortMin:  pulumi.Int(port),
				PortMax:  pulumi.Int(port),
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func userData(k8sVersion *string, parsedPortMappings []utilKind.PortMapping, ip pulumi.StringOutput) (pulumi.StringPtrInput, error) {
	wrapped := ip.ApplyT(
		func(publicIP string) (*string, error) {
			cc := &utilKind.CloudConfigArgs{
				Arch:              utilKind.X86_64,
				KindVersion:       utilKind.KindK8sVersions[*k8sVersion].KindVersion,
				KindImage:         utilKind.KindK8sVersions[*k8sVersion].KindImage,
				Username:          defaultUser,
				PublicIP:          publicIP,
				ExtraPortMappings: parsedPortMappings,
			}
			rawCC, err := file.Template(cc, string(cloudConfigTemplate))
			if err != nil {
				return nil, err
			}
			result := mimeWrapCloudConfig(rawCC)
			return &result, nil
		}).(pulumi.StringPtrOutput)
	return wrapped, nil
}

// mimeWrapCloudConfig wraps a cloud-config YAML string in a MIME multipart
// envelope. IBM Cloud VPC passes user_data as-is to cloud-init and does not
// recognise a bare base64 string; wrapping it this way causes cloud-init to
// decode and process the payload correctly.
func mimeWrapCloudConfig(rawCC string) string {
	const boundary = "MAPT-CLOUD-CONFIG"
	encoded := base64.StdEncoding.EncodeToString([]byte(rawCC))
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
	}, "\n")
}

func kubeconfig(ctx *pulumi.Context, prefix *string, ip pulumi.StringOutput,
	mk *tls.PrivateKey, dependencies []pulumi.Resource,
) (pulumi.StringOutput, error) {
	kindReadyCmd, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudKindID, "ready"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				User:           pulumi.String(defaultUser),
				PrivateKey:     mk.PrivateKeyOpenssh,
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(command.CommandCloudInitWait),
			Update: pulumi.String(command.CommandCloudInitWait),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "15m", Update: "15m"}),
		pulumi.DependsOn(dependencies))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	getKCCmd := fmt.Sprintf("cat /home/%s/kubeconfig", defaultUser)
	getKC, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudKindID, "kc"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				User:           pulumi.String(defaultUser),
				PrivateKey:     mk.PrivateKeyOpenssh,
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(getKCCmd),
			Update: pulumi.String(getKCCmd),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn([]pulumi.Resource{kindReadyCmd}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	kc := pulumi.All(getKC.Stdout, ip).ApplyT(
		func(args []interface{}) string {
			re := regexp.MustCompile(`https://[^:]+:\d+`)
			return re.ReplaceAllString(
				args[0].(string),
				fmt.Sprintf("https://%s:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kc, nil
}

func isKey(ctx *pulumi.Context, mCtx *mc.Context, prefix, cId string,
	rg *ibmcloud.ResourceGroup, tags pulumi.StringArray,
) (*tls.PrivateKey, *ibmcloud.IsSshKey, error) {
	pk, err := tls.NewPrivateKey(ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pk"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return nil, nil, err
	}
	if mCtx.Debug() {
		pk.PrivateKeyPem.ApplyT(func(privateKey string) error {
			logging.Debugf("%s", privateKey)
			return nil
		})
	}
	sshKeyArgs := &ibmcloud.IsSshKeyArgs{
		Name:      pulumi.String(mCtx.ProjectName()),
		PublicKey: pk.PublicKeyOpenssh,
		Tags:      tags,
	}
	if rg != nil {
		sshKeyArgs.ResourceGroup = rg.ID()
	}
	pik, err := ibmcloud.NewIsSshKey(ctx,
		resourcesUtil.GetResourceName(prefix, cId, "pik"),
		sshKeyArgs)
	return pk, pik, err
}
