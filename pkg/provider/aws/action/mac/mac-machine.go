package mac

import (
	_ "embed"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	infra "github.com/adrianriobo/qenvs/pkg/provider"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/bastion"
	na "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network/airgap"
	ns "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network/standard"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/provider/util/command"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/provider/util/security"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/file"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed bootstrap.sh
var BootstrapScript []byte

type userDataValues struct {
	Username string
	Password string
}

// this creates the stack for the mac machine
func (r *MacRequest) createMacMachine() error {
	// If request does not set onlyHost we will create the mac machine
	if len(r.AvailabilityZone) == 0 {
		dedicatedHostAZ, err := getDedicatedHostZoneName(r.HostID)
		if err != nil {
			return err
		}
		r.AvailabilityZone = *dedicatedHostAZ
	}
	region := r.AvailabilityZone[:len(r.AvailabilityZone)-1]
	cs := manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName(stackMacMachine),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: region}),
		DeployFunc: r.deployerMachine,
	}
	csResult, err := manager.UpStack(cs)
	if err != nil {
		return err
	}
	err = r.manageResultsMachine(
		csResult, qenvsContext.GetResultsOutput())
	if err != nil {
		return err
	}
	return nil
}

// this creates the stack for the mac machine
func (r *MacRequest) createAirgapMacMachine() error {
	r.airgapPhaseConnectivity = on
	err := r.createMacMachine()
	if err != nil {
		return nil
	}
	r.airgapPhaseConnectivity = off
	return r.createMacMachine()
}

// Main function to deploy all requried resources to azure
func (r *MacRequest) deployerMachine(ctx *pulumi.Context) error {
	// Export the region for delete
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputRegion), pulumi.String(r.Region))
	// Lookup AMI
	ami, err := ami.GetAMIByName(ctx,
		fmt.Sprintf(amiRegex, r.Version),
		amiOwner,
		map[string]string{
			"architecture": awsArchIDbyArch[r.Architecture]})
	if err != nil {
		return err
	}
	vpc, targetSubnet, targetRouteTableAssociation, bastion, err := r.network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, awsMacMachineID, "pk")}
	keyResources, err := kpr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := r.securityGroups(ctx, vpc)
	if err != nil {
		return err
	}
	// Create instance
	i, err := r.instance(ctx, targetSubnet, ami, keyResources, securityGroups)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername),
		pulumi.String(defaultUsername))
	if r.Airgap {
		ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost),
			i.PrivateIp)
	} else {
		ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost),
			i.PublicIp)
	}
	// Bootstrap script
	bSDependecies := []pulumi.Resource{i}
	if bastion != nil {
		bSDependecies = append(bSDependecies,
			[]pulumi.Resource{bastion.Instance, targetRouteTableAssociation}...)
	}
	bc, userPassword, err := r.bootstrapscript(
		ctx, i, keyResources.PrivateKey, bastion, bSDependecies)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword), userPassword.Result)
	return r.readiness(ctx, i, keyResources.PrivateKey, bastion, []pulumi.Resource{bc})
}

// Write exported values in context to files o a selected target folder
func (r *MacRequest) manageResultsMachine(stackResult auto.UpResult,
	destinationFolder string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):   "userpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):           "host",
	}
	if r.Airgap {
		results[fmt.Sprintf("%s-%s", r.Prefix, outputBastionUserPrivateKey)] = "bastion_id_rsa"
		results[fmt.Sprintf("%s-%s", r.Prefix, outputBastionUsername)] = "bastion_username"
		results[fmt.Sprintf("%s-%s", r.Prefix, outputBastionHost)] = "bastion_host"
	}
	return output.Write(stackResult, destinationFolder, results)
}

// this function will return the AZ for the dedicated host
// dedicated host are tied to an specific Az, as so we need to create
// all resources within the mac machine on that specific region
func getDedicatedHostZoneName(dhID string) (*string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	dh, err := svc.DescribeHosts(&awsEC2.DescribeHostsInput{
		HostIds: []*string{&dhID},
	})
	if err != nil {
		return nil, err
	}
	return dh.Hosts[0].AvailabilityZone, nil
}

// security group for mac machine with ingress rules for ssh and vnc
func (r *MacRequest) securityGroups(ctx *pulumi.Context,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// ingress for vnc access from 0.0.0.0
	vncIngressRule := securityGroup.IngressRules{
		Description: fmt.Sprintf("VNC port for %s", awsMacMachineID),
		FromPort:    vncDefaultPort,
		ToPort:      vncDefaultPort,
		Protocol:    "tcp",
		CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
	}
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsMacMachineID),
		IngressRules: []securityGroup.IngressRules{
			sshIngressRule, vncIngressRule},
	}.Create(ctx)
	if err != nil {
		return nil, err
	}
	// Convert to an array of IDs
	sgs := util.ArrayConvert([]*ec2.SecurityGroup{sg.SG},
		func(sg *ec2.SecurityGroup) pulumi.StringInput {
			return sg.ID()
		})
	return pulumi.StringArray(sgs[:]), nil
}

// Create the mac instance
func (r *MacRequest) instance(ctx *pulumi.Context,
	subnet *ec2.Subnet,
	ami *ec2.LookupAmiResult,
	keyResources *keypair.KeyPairResources,
	securityGroups pulumi.StringArray,
) (*ec2.Instance, error) {
	instanceArgs := ec2.InstanceArgs{
		HostId:                   pulumi.String(r.HostID),
		SubnetId:                 subnet.ID(),
		Ami:                      pulumi.String(ami.Id),
		InstanceType:             pulumi.String(macTypesByArch[r.Architecture]),
		KeyName:                  keyResources.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(true),
		VpcSecurityGroupIds:      securityGroups,
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(diskSize),
		},
		Tags: qenvsContext.GetTagsAsPulumiStringMap(),
	}
	if r.Airgap {
		instanceArgs.AssociatePublicIpAddress = pulumi.Bool(false)
	}
	return ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "instance"),
		&instanceArgs)
}

func (r *MacRequest) network(ctx *pulumi.Context) (
	vpc *ec2.Vpc,
	targetSubnet *ec2.Subnet,
	targetRouteTableAssociation *ec2.RouteTableAssociation,
	b *bastion.BastionResources,
	err error) {
	if !r.Airgap {
		vpc, targetSubnet, err = r.manageNetworking(ctx)
		return
	} else {
		var publicSubnet *ec2.Subnet
		if vpc, publicSubnet, targetSubnet, targetRouteTableAssociation, err =
			r.manageAirgapNetworking(ctx); err != nil {
			return nil, nil, nil, nil, err
		}
		br := bastion.BastionRequest{
			Prefix: r.Prefix,
			VPC:    vpc,
			Subnet: publicSubnet,
			// private key for bastion will be exported with this key
			OutputKeyPrivateKey: fmt.Sprintf("%s-%s", r.Prefix, outputBastionUserPrivateKey),
			OutputKeyUsername:   fmt.Sprintf("%s-%s", r.Prefix, outputBastionUsername),
			OutputKeyHost:       fmt.Sprintf("%s-%s", r.Prefix, outputBastionHost),
		}
		b, err = br.Create(ctx)
		return
	}
}

// Create a standard network (only one public subnet)
func (r *MacRequest) manageNetworking(ctx *pulumi.Context) (*ec2.Vpc, *ec2.Subnet, error) {
	net, err := ns.NetworkRequest{
		CIDR:               cidrVN,
		Name:               resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "net"),
		Region:             r.Region,
		AvailabilityZones:  []string{r.AvailabilityZone},
		PublicSubnetsCIDRs: []string{cidrPublicSN},
		SingleNatGateway:   true,
	}.CreateNetwork(ctx)
	if err != nil {
		return nil, nil, err
	}
	return net.VPCResources.VPC,
		net.PublicSNResources[0].Subnet,
		nil
}

// Create an airgap scenario (on and off phases will be executed to remove the nat gateway on the off phase)
func (r *MacRequest) manageAirgapNetworking(ctx *pulumi.Context) (
	vpc *ec2.Vpc,
	publicSubnet *ec2.Subnet,
	targetSubnet *ec2.Subnet,
	targetRouteTableAssociation *ec2.RouteTableAssociation,
	err error) {
	net, err := na.AirgapNetworkRequest{
		CIDR:             cidrVN,
		Name:             resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "net"),
		Region:           r.Region,
		AvailabilityZone: r.AvailabilityZone,
		PublicSubnetCIDR: cidrPublicSN,
		TargetSubnetCIDR: cidrIntraSN,
		SetAsAirgap:      r.airgapPhaseConnectivity == off}.CreateNetwork(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return net.VPCResources.VPC,
		net.PublicSubnet.Subnet,
		net.TargetSubnet.Subnet,
		net.TargetSubnet.RouteTableAssociation,
		nil
}

func (r *MacRequest) bootstrapscript(ctx *pulumi.Context,
	m *ec2.Instance,
	mk *tls.PrivateKey,
	b *bastion.BastionResources,
	dependecies []pulumi.Resource) (
	*remote.Command,
	*random.RandomPassword,
	error) {
	// Bootstrap script
	remoteCommand, userPassword, err := r.getBootstrapScript(ctx)
	if err != nil {
		return nil, nil, err
	}
	rc, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "bootstrap-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(m, mk, b),
			Create:     remoteCommand,
			Update:     remoteCommand,
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: remoteTimeout,
				Update: remoteTimeout}),
		pulumi.DependsOn(dependecies))
	return rc, userPassword, err
}

// fuction will return the bootstrap script which will be execute on the mac machine
// during the start of the machine
func (r *MacRequest) getBootstrapScript(ctx *pulumi.Context) (
	pulumi.StringPtrInput,
	*random.RandomPassword,
	error) {
	password, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "passwd"))
	if err != nil {
		return nil, nil, err
	}

	postscript := password.Result.ApplyT(func(password string) (string, error) {
		return file.Template(
			userDataValues{
				defaultUsername,
				password},
			fmt.Sprintf("%s-%s", r.Prefix, outputUsername),
			string(BootstrapScript[:]))

	}).(pulumi.StringOutput)
	return postscript, password, nil
}

func (r *MacRequest) readiness(ctx *pulumi.Context,
	m *ec2.Instance,
	mk *tls.PrivateKey,
	b *bastion.BastionResources,
	dependecies []pulumi.Resource) error {
	_, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "readiness-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(m, mk, b),
			Create:     pulumi.String(command.CommandPing),
			Update:     pulumi.String(command.CommandPing),
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: remoteTimeout,
				Update: remoteTimeout}),
		pulumi.DependsOn(dependecies))
	return err
}

// helper function to set the connection args
// based on bastion or direct connection to target host
func remoteCommandArgs(
	m *ec2.Instance,
	mk *tls.PrivateKey,
	b *bastion.BastionResources) remote.ConnectionArgs {
	ca := remote.ConnectionArgs{
		Host:           m.PublicIp,
		PrivateKey:     mk.PrivateKeyOpenssh,
		User:           pulumi.String(defaultUsername),
		Port:           pulumi.Float64(defaultSSHPort),
		DialErrorLimit: pulumi.Int(-1)}
	if b != nil {
		// If airgap set the private IP for host
		// And bastion details
		ca.Host = m.PrivateIp
		ca.Proxy = remote.ProxyConnectionArgs{
			Host:           b.Instance.PublicIp,
			PrivateKey:     b.PrivateKey.PrivateKeyOpenssh,
			User:           pulumi.String(b.Usarname),
			Port:           pulumi.Float64(b.Port),
			DialErrorLimit: pulumi.Int(-1)}

	}
	return ca
}
