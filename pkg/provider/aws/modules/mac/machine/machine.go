package machine

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macSetup "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine/setup"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	qEC2 "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
)

type locked struct {
	pulumi.ResourceState
	Lock bool
}

// This function will use the information from the
// dedicated host holding the mac machine will check if stack exists
// if exists will get the lock value from it
func ReplaceMachine(h *mac.HostInformation) error {
	aN := fmt.Sprintf(amiRegex, *h.OSVersion)
	bdt := blockDeviceType
	ami, err := data.GetAMI(
		data.ImageRequest{
			Name:            &aN,
			Arch:            h.Arch,
			Region:          h.Region,
			BlockDeviceType: &bdt})
	if err != nil {
		return err
	}
	logging.Debugf("Replacing root volume for AMI %s", *ami.Image.ImageId)
	if _, err = qEC2.ReplaceRootVolume(
		qEC2.ReplaceRootVolumeRequest{
			Region:     *h.Region,
			InstanceID: *h.Host.Instances[0].InstanceId,
			// Needto lookup for AMI + check if copy is required
			AMIID: *ami.Image.ImageId,
			Wait:  true,
		}); err != nil {
		return err
	}
	// Set a default request
	r := &Request{
		Prefix:             *h.Prefix,
		Architecture:       *h.Arch,
		Version:            *h.OSVersion,
		lock:               false,
		remoteTimeout:      releaseTimeout,
		isRequestOperation: false,
	}
	return r.manageMacMachine(h)
}

// Run the bootstrap script creating new access credentials for the user
// When the machine is requested it will have different user access
// Wil include the code to join the cicd if any integration is set
// Also machine will be in use from now on, so it will be locked
func (r *Request) ManageRequest(h *mac.HostInformation) error {
	r.lock = true
	r.isRequestOperation = true
	r.remoteTimeout = requestTimeout
	return r.manageMacMachine(h)
}

// This create the machine and set as locked....meaning that it will return a way
// for it to be used (i.e mac action)
func (r *Request) CreateAndLockMacMachine(h *mac.HostInformation) error {
	r.lock = true
	r.isRequestOperation = true
	return r.manageMacMachine(h)
}

// This create the machine and set it as ready to be used (i.e when mac-pool action)
// in this case machines are added to the pool as ready to be used by request
func (r *Request) CreateAvailableMacMachine(h *mac.HostInformation) error {
	r.lock = false
	r.isRequestOperation = false
	return r.manageMacMachine(h)
}

// this creates the stack for the mac machine
func (r *Request) manageMacMachine(h *mac.HostInformation) error {
	return r.manageMacMachineTargets(h, nil)
}

// this creates the stack for the mac machine
func (r *Request) manageMacMachineTargets(h *mac.HostInformation, targetURNs []string) error {
	r.AvailabilityZone = h.Host.AvailabilityZone
	r.dedicatedHost = h
	r.Region = h.Region
	cpk, cp, err := currentCredentials(*h.ProjectName, *h.BackedURL, *h.Prefix)
	if err != nil {
		logging.Debugf("%v", err)
	}
	// Both are managed at same time
	if cpk != nil && cp != nil {
		r.currentPrivateKey = *cpk
		r.currentPassword = *cp
	}
	if err != nil {
		logging.Debugf("%v", err)
	}
	cs := manager.Stack{
		StackName: fmt.Sprintf("%s-%s",
			mac.StackMacMachine, *h.ProjectName),
		ProjectName: *h.ProjectName,
		// Backed url always should be set from request as it is picked from the
		// backed url for the dedicated host (pick from the value add as a tag on the dedicated host resoruce)
		BackedURL: *h.BackedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *h.Region}),
		DeployFunc: r.deployerMachine,
	}
	var sr auto.UpResult
	if len(targetURNs) > 0 {
		sr, _ = manager.UpStackTargets(cs, targetURNs)
	} else {
		sr, _ = manager.UpStack(cs)
	}
	return r.manageResultsMachine(sr)
}

// this creates the stack for the mac machine
func (r *Request) CreateAirgapMacMachine(h *mac.HostInformation) error {
	r.airgapPhaseConnectivity = network.ON
	err := r.CreateAndLockMacMachine(h)
	if err != nil {
		return nil
	}
	r.airgapPhaseConnectivity = network.OFF
	return r.CreateAndLockMacMachine(h)
}

// Main function to deploy all requried resources to azure
func (r *Request) deployerMachine(ctx *pulumi.Context) error {
	// Export information
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputRegion), pulumi.String(*r.Region))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID), pulumi.String(*r.dedicatedHost.Host.HostId))
	// Lookup AMI
	aN := fmt.Sprintf(amiRegex, r.Version)
	bdt := blockDeviceType
	arch := util.If(
		isAWSArchID(r.Architecture),
		r.Architecture,
		awsArchIDbyArch[r.Architecture])
	ami, err := data.GetAMI(
		data.ImageRequest{
			Name:            &aN,
			Arch:            &arch,
			Region:          r.Region,
			BlockDeviceType: &bdt})
	if err != nil {
		return err
	}
	nr := network.NetworkRequest{
		Prefix:                  r.Prefix,
		ID:                      awsMacMachineID,
		Region:                  *r.Region,
		AZ:                      *r.AvailabilityZone,
		Airgap:                  r.Airgap,
		AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
	}
	vpc, targetSubnet, targetRouteTableAssociation, bastion, _, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	machineKeyPair := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, awsMacMachineID, "pk-machine")}
	machineKeyPairResources, err := machineKeyPair.Create(ctx)
	if err != nil {
		return err
	}
	// TODO Check Export it is this needed?
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputMachinePrivateKey),
		machineKeyPairResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := r.securityGroups(ctx, vpc)
	if err != nil {
		return err
	}
	// Create instance
	i, err := r.instance(ctx, targetSubnet, ami, machineKeyPairResources, securityGroups)
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
	bc, userPassword, ukp, err := r.bootstrapscript(
		ctx, i, machineKeyPairResources.PrivateKey, bastion, bSDependecies)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword), userPassword.Result)
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey),
		ukp.PrivateKey.PrivateKeyPem)
	readiness, err := r.readiness(ctx, i, ukp.PrivateKey, bastion, []pulumi.Resource{bc})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputLock), pulumi.Bool(r.lock))
	return machineLock(ctx,
		resourcesUtil.GetResourceName(
			r.Prefix, awsMacMachineID, "mac-lock"), r.lock,
		pulumi.DependsOn([]pulumi.Resource{readiness}))
}

// Write exported values in context to files o a selected target folder
func (r *Request) manageResultsMachine(stackResult auto.UpResult) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):          "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):      "userpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey):    "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):              "host",
		fmt.Sprintf("%s-%s", r.Prefix, outputMachinePrivateKey): "machine_id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID):   "dedicated_host_id",
	}
	if r.Airgap {
		err := bastion.WriteOutputs(stackResult, r.Prefix, maptContext.GetResultsOutputPath())
		if err != nil {
			return err
		}
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func (r *Request) securityGroups(ctx *pulumi.Context,
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
	ingressRules := []securityGroup.IngressRules{
		sshIngressRule, vncIngressRule}
	// Integration ports
	cirrusPort, err := cirrus.CirrusPort()
	if err != nil {
		return nil, err
	}
	if cirrusPort != nil {
		ingressRules = append(ingressRules,
			securityGroup.IngressRules{
				Description: fmt.Sprintf("Cirrus port for %s", awsMacMachineID),
				FromPort:    *cirrusPort,
				ToPort:      *cirrusPort,
				Protocol:    "tcp",
				CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
			})
	}
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", awsMacMachineID),
		IngressRules: ingressRules,
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
func (r *Request) instance(ctx *pulumi.Context,
	subnet *ec2.Subnet,
	ami *data.ImageInfo,
	keyResources *keypair.KeyPairResources,
	securityGroups pulumi.StringArray,
) (*ec2.Instance, error) {
	instanceArgs := ec2.InstanceArgs{
		HostId:                   pulumi.String(*r.dedicatedHost.Host.HostId),
		SubnetId:                 subnet.ID(),
		Ami:                      pulumi.String(*ami.Image.ImageId),
		InstanceType:             pulumi.String(mac.TypesByArch[r.Architecture]),
		KeyName:                  keyResources.AWSKeyPair.KeyName,
		AssociatePublicIpAddress: pulumi.Bool(true),
		VpcSecurityGroupIds:      securityGroups,
		RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(diskSize),
		},
		Tags: maptContext.ResourceTags(),
	}
	if r.Airgap {
		instanceArgs.AssociatePublicIpAddress = pulumi.Bool(false)
	}
	return ec2.NewInstance(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "instance"),
		&instanceArgs,
		// Retain on delete to speed destroy operation for the dedicated host,
		// destroy is managed by replace root volume operation
		// pulumi.RetainOnDelete(true),
		// All changes on the instance should be done through root volume replace
		// as so we ignore Amis missmatch
		pulumi.IgnoreChanges([]string{"ami"}))
}

func (r *Request) bootstrapscript(ctx *pulumi.Context,
	m *ec2.Instance,
	mk *tls.PrivateKey,
	b *bastion.Bastion,
	dependecies []pulumi.Resource) (
	*remote.Command,
	*random.RandomPassword,
	*keypair.KeyPairResources,
	error) {
	// Bootstrap script
	remoteCommand, userPassword, ukp, err := r.getBootstrapScript(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	remoteKey := util.If[pulumi.StringPtrInput](
		r.isRequestOperation,
		pulumi.String(r.currentPrivateKey),
		mk.PrivateKeyOpenssh)
	timeout := util.If(len(r.remoteTimeout) > 0, r.remoteTimeout, defaultTimeout)
	rc, err := remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "bootstrap-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(m, remoteKey, b),
			Create:     remoteCommand,
			Update:     remoteCommand,
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: timeout,
				Update: timeout}),
		pulumi.DependsOn(append(dependecies, ukp.PrivateKey, ukp.AWSKeyPair)),
		pulumi.DeleteBeforeReplace(true))
	return rc, userPassword, ukp, err
}

// fuction will return the bootstrap script which will be execute on the mac machine
// during the start of the machine
func (r *Request) getBootstrapScript(ctx *pulumi.Context) (
	pulumi.StringPtrInput,
	*random.RandomPassword,
	*keypair.KeyPairResources,
	error) {

	ukpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, awsMacMachineID, "pk-user")}
	ukp, err := ukpr.CreateAlways(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	password, err := security.CreatePasswordAlways(ctx,
		resourcesUtil.GetResourceName(
			r.Prefix, awsMacMachineID, "pass-user"))
	if err != nil {
		return nil, nil, nil, err
	}
	postscript := pulumi.All(password.Result, ukp.PrivateKey.PublicKeyOpenssh).ApplyT(
		func(args []interface{}) (string, error) {
			password := args[0].(string)
			authorizedKey := args[1].(string)

			if r.isRequestOperation {
				return macSetup.Request(
					defaultUsername,
					r.currentPassword,
					password,
					authorizedKey,
					r.SetupGHActionsRunner,
					github.GetActionRunnerSnippetMacos())
			}
			return macSetup.Release(
				defaultUsername,
				password,
				authorizedKey)
		}).(pulumi.StringOutput)
	return postscript, password, ukp, nil
}

func (r *Request) readiness(ctx *pulumi.Context,
	m *ec2.Instance,
	mk *tls.PrivateKey,
	b *bastion.Bastion,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	timeout := util.If(len(r.remoteTimeout) > 0, r.remoteTimeout, defaultTimeout)
	return remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "readiness-cmd"),
		&remote.CommandArgs{
			Connection: remoteCommandArgs(m, mk.PrivateKeyOpenssh, b),
			Create:     pulumi.String(command.CommandPing),
			Update:     pulumi.String(command.CommandPing),
		}, pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: timeout,
				Update: timeout}),
		pulumi.DependsOn(dependecies))
}

// helper function to set the connection args
// based on bastion or direct connection to target host
func remoteCommandArgs(
	m *ec2.Instance,
	pk pulumi.StringPtrInput,
	b *bastion.Bastion) remote.ConnectionArgs {
	ca := remote.ConnectionArgs{
		Host:           m.PublicIp,
		PrivateKey:     pk,
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

// This will create mark to see if machine is lock or free
func machineLock(ctx *pulumi.Context, name string, lockedValue bool, opts ...pulumi.ResourceOption) error {
	l := &locked{
		Lock: lockedValue,
	}
	if err := ctx.RegisterComponentResource(
		customResourceTypeLock,
		name,
		l,
		opts...); err != nil {
		return err
	}
	if err := ctx.RegisterResourceOutputs(l, pulumi.Map{
		"lockState": pulumi.Bool(lockedValue),
	}); err != nil {
		return err
	}
	return nil
}

// In order to change the user credentials the old value for the user private key is required
// this behavior is not offered by pulumi, but we can rely on inspecting previous state file
func currentCredentials(projecName, backedURL, prefix string) (
	pk, pass *string, err error) {
	stack, err := manager.CheckStack(manager.Stack{
		StackName: fmt.Sprintf("%s-%s",
			mac.StackMacMachine, projecName),
		ProjectName: projecName,
		BackedURL:   backedURL})
	if err != nil {
		return nil, nil, err
	}
	return getCredentialsFromStack(stack, prefix)
}

// Function to get secrets for current user o
func getCredentialsFromStack(stack *auto.Stack, prefix string) (
	pk, pass *string, err error) {
	outputs, err := manager.GetOutputs(stack)
	if err != nil {
		return nil, nil, err
	}
	if len(outputs) == 0 {
		return nil, nil, errors.New("stack outputs are empty please destroy and re-create")
	}
	if value, exists := outputs[fmt.Sprintf("%s-%s", prefix, outputUserPrivateKey)]; exists {
		pkv := value.Value.(string)
		pk = &pkv
	}
	if value, exists := outputs[fmt.Sprintf("%s-%s", prefix, outputUserPassword)]; exists {
		pv := value.Value.(string)
		pass = &pv
	}
	return pk, pass, nil
}
