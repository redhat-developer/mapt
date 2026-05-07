package ibmz

import (
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
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

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
}

type zRequest struct {
	mCtx     *mc.Context
	prefix   *string
	zone     *string
	subnetID *string
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
		mCtx:     mCtx,
		prefix:   &prefix,
		zone:     zone,
		subnetID: subnetID,
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
	return ibmcloudp.Destroy(mCtx, stackIBMS390)
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
	// https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui&q=s390x&tags=vpc
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "is"),
		&ibmcloud.IsInstanceArgs{
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
		})
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
	sg, err := ibmcloud.NewIsSecurityGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "sg"),
		&ibmcloud.IsSecurityGroupArgs{
			Name: pulumi.String(name),
			Vpc:  pulumi.String(subnetInfo.Vpc),
		})
	if err != nil {
		return err
	}
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "ssh"),
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     sg.ID(),
			Direction: pulumi.String("inbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
			Tcp: &ibmcloud.IsSecurityGroupRuleTcpArgs{
				PortMin: pulumi.Int(22),
				PortMax: pulumi.Int(22),
			},
		})
	if err != nil {
		return err
	}
	_, err = ibmcloud.NewIsSecurityGroupRule(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "outb"),
		&ibmcloud.IsSecurityGroupRuleArgs{
			Group:     sg.ID(),
			Direction: pulumi.String("outbound"),
			Remote:    pulumi.String("0.0.0.0/0"),
		})
	if err != nil {
		return err
	}
	fip, err := ibmcloud.NewIsFloatingIp(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "fip"),
		&ibmcloud.IsFloatingIpArgs{
			Name: pulumi.String(name),
			Zone: pulumi.String(subnetInfo.Zone),
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
	// https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui&q=s390x&tags=vpc
	i, err := ibmcloud.NewIsInstance(ctx,
		resourcesUtil.GetResourceName(*r.prefix, stackIBMS390, "is"),
		&ibmcloud.IsInstanceArgs{
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
		})
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
