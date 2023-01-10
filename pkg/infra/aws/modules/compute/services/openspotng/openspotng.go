package openspotng

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/services/openspotng/keys"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	utilRemote "github.com/adrianriobo/qenvs/pkg/infra/util/remote"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const consoleHTTPSPort int = 6443

func (r *OpenspotNGRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *OpenspotNGRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *OpenspotNGRequest) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *OpenspotNGRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *OpenspotNGRequest) CustomIngressRules() []securityGroup.IngressRules {
	return []securityGroup.IngressRules{
		securityGroup.HTTPS_TCP,
		{
			Description: fmt.Sprintf("console https port for %s", r.Specs.ID),
			FromPort:    consoleHTTPSPort,
			ToPort:      consoleHTTPSPort,
			Protocol:    "tcp",
			CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
		},
	}
}

func (r *OpenspotNGRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *OpenspotNGRequest) PostProcess(ctx *pulumi.Context,
	compute *compute.Compute) ([]pulumi.Resource, error) {
	dependencies := []pulumi.Resource{}
	swapKeysDependencies, err := r.swapSSHKeys(ctx, compute)
	if err != nil {
		return nil, err
	}
	dependencies = append(dependencies, swapKeysDependencies...)
	clusterSetupDepdencies, err := r.clusterSetup(ctx, compute, dependencies)
	if err != nil {
		return nil, err
	}
	dependencies = append(dependencies, clusterSetupDepdencies...)
	return dependencies, nil
}

func (r *OpenspotNGRequest) ReadinessCommand() string {
	// If key is changed during postscript the compute.PrivateKeyContent = pulumi.String(keyContent) can be set to null
	// to use default key from keypair created
	return r.Request.ReadinessCommand()
}

func (r *OpenspotNGRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}

// Switch the fixed initial key with self created one
func (r *OpenspotNGRequest) swapSSHKeys(ctx *pulumi.Context,
	computeResource *compute.Compute) ([]pulumi.Resource, error) {
	// var err error
	// var pubKeyRemoteCopyResource *remote.CopyFile
	// var pubKeyRemoteOverrideResource *remote.Command
	dependencies := []pulumi.Resource{}
	// pubKeyRemoteCopyResource,
	// pubKeyRemoteOverrideResource}
	// Get initial key
	instance, err := r.getInitialRemoteConnection(computeResource)
	if err != nil {
		return nil, err
	}
	_ = computeResource.AWSKeyPair.PublicKey.ApplyT(
		func(pubKey string) (string, error) {
			pubKeyFilename, err := util.WriteTempFile(pubKey)
			if err != nil {
				return "", err
			}
			pubKeyRemoteCopyResource, err := compute.CopyOnRemoteInstance(ctx, pubKeyFilename, "id_rsa.pub",
				fmt.Sprintf("%s-%s", r.Specs.ID, "pubKeyUpload"),
				*instance, []pulumi.Resource{})
			if err != nil {
				return "", err
			}
			overrideKeyCommand := "cat /home/core/id_rsa.pub > /home/core/.ssh/authorized_keys"
			_, err = compute.ExecOnRemoteInstance(ctx, pulumi.String(overrideKeyCommand),
				fmt.Sprintf("%s-%s", r.Specs.ID, "pubKeyOverride"), *instance, []pulumi.Resource{pubKeyRemoteCopyResource})
			if err != nil {
				return "", err
			}
			return "", nil
		})
	return dependencies, nil
}

// Initially the AMI comes with a fixed key we need to use to connect and swith to
// a new generated one
func (r *OpenspotNGRequest) getInitialRemoteConnection(
	computeResource *compute.Compute) (*utilRemote.RemoteInstance, error) {
	// Get initial key
	keyContent, err := keys.GetKey(r.Specs.AMI.AMISourceID)
	if err != nil {
		return nil, err
	}
	return &utilRemote.RemoteInstance{
		Instance:   computeResource.Instance,
		InstanceIP: &computeResource.InstanceIP,
		Username:   computeResource.Username,
		PrivateKey: pulumi.String(keyContent)}, nil
}

func (r *OpenspotNGRequest) clusterSetup(ctx *pulumi.Context,
	compute *compute.Compute, dependsOn []pulumi.Resource) ([]pulumi.Resource, error) {
	// var err error
	// var clusterSetupRemoteCopyResource *remote.CopyFile
	// var scriptXRightsCommandResource *remote.Command
	// var execClusterSetupCommandResource *remote.Command
	dependencies := []pulumi.Resource{}
	// clusterSetupRemoteCopyResource,
	// scriptXRightsCommandResource,
	// execClusterSetupCommandResource}
	// Create passwords for users (for time being re use)
	password, err := utilInfra.CreatePassword(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputPassword(), password.Result)
	// Load pull secret content
	pullsecret, err := ioutil.ReadFile(r.OCPPullSecretFilePath)
	if err != nil {
		return nil, err
	}
	clusterSetupTemplate, err := getClusterSetupTemplate()
	if err != nil {
		return nil, err
	}
	_ = pulumi.All(password.Result,
		compute.Instance.PublicIp, compute.Instance.PrivateIp).ApplyT(
		func(args []interface{}) (string, error) {
			// test := string(pullsecret)
			// logging.Debugf("pullsecret is %s", test)
			// unquotedPullSecret, err := strconv.Unquote(string(pullsecret))
			// if err != nil {
			// 	return "", err
			// }
			pullSecretEncoded := base64.StdEncoding.EncodeToString([]byte(pullsecret))
			clustersetupscript, err := util.Template(
				clusterSetupValues{
					InternalIP:        args[2].(string),
					ExternalIP:        args[1].(string),
					PullScret:         pullSecretEncoded,
					DeveloperPassword: args[0].(string),
					KubeadminPassword: args[0].(string),
					RedHatPassword:    args[0].(string),
				},
				"clustersetup", string(clusterSetupTemplate))
			if err != nil {
				return "", err
			}
			clusterSetupfileName, err := util.WriteTempFile(clustersetupscript)
			if err != nil {
				return "", err
			}
			clusterSetupRemoteCopyResource, err := compute.RemoteCopy(ctx,
				clusterSetupfileName,
				"/var/home/core/cluster_setup.sh",
				fmt.Sprintf("%s-%s", r.Specs.ID, "clustersetup"),
				dependsOn)
			if err != nil {
				return "", err
			}
			scriptXRightsCommand := "chmod +x /var/home/core/cluster_setup.sh"
			scriptXRightsCommandResource, err := compute.RemoteExec(ctx, pulumi.String(scriptXRightsCommand),
				fmt.Sprintf("%s-%s", r.Specs.ID, "scriptXRightsCommand"),
				[]pulumi.Resource{clusterSetupRemoteCopyResource})
			if err != nil {
				return "", err
			}
			execClusterSetupCommand := "sudo /var/home/core/cluster_setup.sh"
			_, err = compute.RemoteExec(ctx, pulumi.String(execClusterSetupCommand),
				fmt.Sprintf("%s-%s", r.Specs.ID, "execClusterSetupCommand"),
				[]pulumi.Resource{scriptXRightsCommandResource})
			if err != nil {
				return "", err
			}
			return "", err
		})
	return dependencies, nil
}
