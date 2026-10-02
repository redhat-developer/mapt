package snc

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
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
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	apiSNC "github.com/redhat-developer/mapt/pkg/target/service/snc"
	"github.com/redhat-developer/mapt/pkg/target/service/snc/profile"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

var kubeconfigURLPattern = regexp.MustCompile(`https://api\.crc\.testing:\d+`)

//go:embed cloud-config
var cloudConfigTemplate []byte

type dataValues struct {
	Username         string
	PubKey           string
	PublicIP         string
	PullSecretB64    string
	KubeAdminPassword string
	DeveloperPassword string
}

type sncRequest struct {
	mCtx                    *mc.Context
	prefix                  *string
	version                 *string
	disableClusterReadiness bool
	spot                    bool
	pullSecretFile          *string
	offeringCRN             *string
	profile                 string
	diskSize                int
	zone                    *string
	profiles                []string
	operatorChannels        map[string]string
	catalogSources          map[string]string
}

func Create(mCtxArgs *mc.ContextArgs, args *apiSNC.SNCArgs) (*apiSNC.SNCResults, error) {
	ibmcloudProvider := ibmcloudp.Provider()
	mCtx, err := mc.Init(mCtxArgs, ibmcloudProvider)
	if err != nil {
		return nil, err
	}
	if err := profile.Validate(args.Profiles); err != nil {
		return nil, err
	}
	if err := profile.ValidateOperatorOverrides(args.OperatorChannels, args.CatalogSources); err != nil {
		return nil, err
	}
	offeringName := imageName(args.Version, args.Arch)
	logging.Debugf("Looking up catalog offering %s", offeringName)
	offeringCRN, err := icdata.GetCatalogOfferingVersionCRN(offeringName, args.Version)
	if err != nil {
		return nil, fmt.Errorf("looking up catalog offering for version %s: %w", args.Version, err)
	}
	logging.Debugf("Found offering CRN: %s", offeringCRN)
	zone, err := ibmcloudProvider.Zone()
	if err != nil {
		return nil, err
	}

	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	diskSize := defaultDiskSize
	computeProfile := defaultProfile
	if args.ComputeRequest != nil {
		if args.ComputeRequest.DiskSize != nil {
			diskSize = *args.ComputeRequest.DiskSize
		}
		if len(args.ComputeRequest.ComputeSizes) > 0 {
			computeProfile = args.ComputeRequest.ComputeSizes[0]
		} else if args.ComputeRequest.CPUs > 0 || args.ComputeRequest.MemoryGib > 0 {
			profiles, err := icdata.NewComputeSelector().Select(args.ComputeRequest)
			if err != nil {
				return nil, fmt.Errorf("selecting compute profile: %w", err)
			}
			computeProfile = profiles[0]
		}
	}
	spot := args.Spot != nil && args.Spot.Spot
	if spot {
		ok, err := icdata.ProfileSupportsSpot(computeProfile)
		if err != nil {
			return nil, fmt.Errorf("checking spot support for profile %s: %w", computeProfile, err)
		}
		if !ok {
			return nil, fmt.Errorf("profile %s does not support spot instances", computeProfile)
		}
	}

	r := &sncRequest{
		mCtx:                    mCtx,
		prefix:                  &prefix,
		version:                 &args.Version,
		disableClusterReadiness: args.DisableClusterReadiness,
		pullSecretFile:          &args.PullSecretFile,
		offeringCRN:             &offeringCRN,
		profile:                 computeProfile,
		diskSize:                diskSize,
		spot:                    spot,
		zone:                    zone,
		profiles:                args.Profiles,
		operatorChannels:        args.OperatorChannels,
		catalogSources:          args.CatalogSources,
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(apiSNC.StackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: ibmcloudp.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, err := manager.UpStack(mCtx, cs)
	if err != nil {
		return nil, fmt.Errorf("stack creation failed: %w", err)
	}
	return apiSNC.Results(sr, &prefix,
		mCtx.GetResultsOutputPath(),
		nil,
		r.disableClusterReadiness)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Run openshift-snc destroy")
	mCtx, err := mc.Init(mCtxArgs, ibmcloudp.Provider())
	if err != nil {
		return err
	}
	if err := ibmcloudp.DestroyStack(mCtx, apiSNC.StackName); err != nil {
		return err
	}
	return ibmcloudp.CleanupState(mCtx)
}

func (r *sncRequest) deploy(ctx *pulumi.Context) error {
	userTags := ibmcloudp.TagsAsStringArray(r.mCtx.GetTags())
	zone := *r.zone

	rg, err := ibmcloud.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudSNCID, "rg"),
		&ibmcloud.ResourceGroupArgs{
			Name: pulumi.String(r.mCtx.ProjectName()),
			Tags: userTags,
		})
	if err != nil {
		return err
	}

	n, err := network.New(ctx, &network.NetworkArgs{
		Prefix:      *r.prefix,
		ComponentID: ibmCloudSNCID,
		Name:        fmt.Sprintf("%s-%s", *r.prefix, r.mCtx.ProjectName()),
		Zone:        &zone,
		RG:          rg,
		Tags:        userTags,
	})
	if err != nil {
		return err
	}

	if err := addSNCSecurityGroupRules(ctx, r.prefix, n.SecurityGroup); err != nil {
		return err
	}

	pk, pik, err := isKey(ctx, r.mCtx, *r.prefix, ibmCloudSNCID, rg, userTags)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputUserPrivateKey),
		pk.PrivateKeyPem)
	if r.mCtx.Debug() {
		pk.PrivateKeyPem.ApplyT(func(privateKey string) error {
			logging.Debugf("%s", privateKey)
			return nil
		})
	}

	kaPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudSNCID, "kubeadminpassword"))
	if err != nil {
		return err
	}
	devPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudSNCID, "devpassword"))
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputKubeAdminPass),
		kaPassword.Result)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputDeveloperPass),
		devPassword.Result)

	ud, err := r.userData(pk.PublicKeyOpenssh, n.Floatingip.Address,
		kaPassword.Result, devPassword.Result)
	if err != nil {
		return err
	}

	instanceArgs := &ibmcloud.IsInstanceArgs{
		Name: pulumi.String(r.mCtx.ProjectName()),
		CatalogOffering: &ibmcloud.IsInstanceCatalogOfferingArgs{
			VersionCrn: pulumi.StringPtr(*r.offeringCRN),
		},
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
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudSNCID, "is"),
		instanceArgs)
	if err != nil {
		return err
	}

	fipAssoc, err := ibmcloud.NewIsInstanceNetworkInterfaceFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, ibmCloudSNCID, "fipassoc"),
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

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputUsername),
		pulumi.String(defaultUser))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputHost),
		n.Floatingip.Address)

	kc, err := kubeconfig(ctx, r.prefix, n.Floatingip.Address, pk,
		*r.version, r.disableClusterReadiness, []pulumi.Resource{fipAssoc})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputKubeconfig),
		pulumi.ToSecret(kc))

	if len(r.profiles) > 0 {
		k8sProvider, err := profile.NewK8sProvider(ctx, "k8s-provider", kc)
		if err != nil {
			return err
		}
		if err := profile.Deploy(ctx, r.profiles, &profile.DeployArgs{
			K8sProvider:      k8sProvider,
			Kubeconfig:       kc,
			Prefix:           *r.prefix,
			DeletedWith:      i,
			OperatorChannels: r.operatorChannels,
			CatalogSources:   r.catalogSources,
		}); err != nil {
			return err
		}
	}

	return nil
}

func addSNCSecurityGroupRules(ctx *pulumi.Context, prefix *string, sg *ibmcloud.IsSecurityGroup) error {
	for _, port := range []int{apiSNC.PortHTTPS, apiSNC.PortAPI} {
		_, err := ibmcloud.NewIsSecurityGroupRule(ctx,
			resourcesUtil.GetResourceName(*prefix, ibmCloudSNCID, fmt.Sprintf("sgr%d", port)),
			&ibmcloud.IsSecurityGroupRuleArgs{
				Group:     sg.ID(),
				Direction: pulumi.String("inbound"),
				Remote:    pulumi.String("0.0.0.0/0"),
				Protocol:  pulumi.String("tcp"),
				PortMin:   pulumi.Int(port),
				PortMax:   pulumi.Int(port),
			})
		if err != nil {
			return err
		}
	}
	return nil
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

func (r *sncRequest) userData(
	pubKey, ip pulumi.StringOutput,
	kaPass, devPass pulumi.StringOutput,
) (pulumi.StringPtrInput, error) {
	ps, err := os.ReadFile(*r.pullSecretFile)
	if err != nil {
		return nil, err
	}
	psB64 := base64.StdEncoding.EncodeToString(ps)

	wrapped := pulumi.All(pubKey, ip, kaPass, devPass).ApplyT(
		func(args []interface{}) (*string, error) {
			data := dataValues{
				Username:          defaultUser,
				PubKey:            args[0].(string),
				PublicIP:          args[1].(string),
				PullSecretB64:     psB64,
				KubeAdminPassword: args[2].(string),
				DeveloperPassword: args[3].(string),
			}
			rawCC, err := file.Template(data, string(cloudConfigTemplate))
			if err != nil {
				return nil, err
			}
			result := mimeWrapCloudConfig(rawCC)
			return &result, nil
		}).(pulumi.StringPtrOutput)
	return wrapped, nil
}

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
	mk *tls.PrivateKey, ocpVersion string, disableClusterReadiness bool,
	dependencies []pulumi.Resource,
) (pulumi.StringOutput, error) {
	sshReadyCmd, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudSNCID, "ssh-ready"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				User:           pulumi.String(defaultUser),
				PrivateKey:     mk.PrivateKeyOpenssh,
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(command.CommandPing),
			Update: pulumi.String(command.CommandPing),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: command.RemoteTimeout, Update: command.RemoteTimeout}),
		pulumi.DependsOn(dependencies))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	readinessCmd := util.If(disableClusterReadiness,
		apiSNC.CommandKubeconfigExists, apiSNC.CommandCrcReadiness)
	ocpReadyCmd, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudSNCID, "ocp-ready"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				User:           pulumi.String(defaultUser),
				PrivateKey:     mk.PrivateKeyOpenssh,
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(readinessCmd),
			Update: pulumi.String(readinessCmd),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: command.RemoteTimeout, Update: command.RemoteTimeout}),
		pulumi.DependsOn([]pulumi.Resource{sshReadyCmd}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	caCmd := apiSNC.CommandCaServiceRan(ocpVersion)
	ocpCaCmd, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudSNCID, "ocp-ca"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				User:           pulumi.String(defaultUser),
				PrivateKey:     mk.PrivateKeyOpenssh,
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(caCmd),
			Update: pulumi.String(caCmd),
		},
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: command.RemoteTimeout, Update: command.RemoteTimeout}),
		pulumi.DependsOn([]pulumi.Resource{ocpReadyCmd}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	getKCCmd := "sudo cat /opt/crc/kubeconfig"
	getKC, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*prefix, ibmCloudSNCID, "kc"),
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
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: command.RemoteTimeout, Update: command.RemoteTimeout}),
		pulumi.DependsOn([]pulumi.Resource{ocpCaCmd}),
		pulumi.AdditionalSecretOutputs([]string{"stdout"}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	kc := pulumi.All(getKC.Stdout, ip).ApplyT(
		func(args []interface{}) string {
			return kubeconfigURLPattern.ReplaceAllString(
				args[0].(string),
				fmt.Sprintf("https://api.%s.nip.io:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kc, nil
}
