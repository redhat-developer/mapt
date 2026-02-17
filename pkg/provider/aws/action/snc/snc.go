package openshiftsnc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/iam"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ssm"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	apiSNC "github.com/redhat-developer/mapt/pkg/target/service/snc"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type openshiftSNCRequest struct {
	mCtx                    *mc.Context
	prefix                  *string
	version                 *string
	disableClusterReadiness bool
	arch                    *string
	spot                    bool
	timeout                 *string
	pullSecretFile          *string
	allocationData          *allocation.AllocationResult
}

func (r *openshiftSNCRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *apiSNC.SNCArgs) (_ *apiSNC.SNCResults, err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return nil, err
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := openshiftSNCRequest{
		mCtx:                    mCtx,
		prefix:                  &prefix,
		version:                 &args.Version,
		disableClusterReadiness: args.DisableClusterReadiness,
		arch:                    &args.Arch,
		pullSecretFile:          &args.PullSecretFile,
		timeout:                 &args.Timeout}
	if args.Spot != nil {
		r.spot = args.Spot.Spot
	}
	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			Prefix:                &args.Prefix,
			ComputeRequest:        args.ComputeRequest,
			AMIProductDescription: &amiProduct,
			Spot:                  args.Spot,
		})
	if err != nil {
		return nil, err
	}
	// check if AMI exists
	amiName := amiName(&args.Version, &args.Arch)
	if err = checkAMIExists(mCtx.Context(), &amiName, r.allocationData.Region, &args.Arch); err != nil {
		return nil, err
	}
	return r.createCluster()
}

// Will destroy resources related to machine
func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	logging.Debug("Run openshift destroy")
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	// Destroy fedora related resources
	if err = aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			Stackname: apiSNC.StackName,
		}); err != nil {
		return err
	}
	// Destroy spot orchestrated stack
	if spot.Exist(mCtx) {
		if err := spot.Destroy(mCtx); err != nil {
			return err
		}
	}

	// Cleanup S3 state after all stacks have been destroyed
	return aws.CleanupState(mCtx)
}

func (r *openshiftSNCRequest) createCluster() (*apiSNC.SNCResults, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(apiSNC.StackName),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}
	sr, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return nil, fmt.Errorf("stack creation failed: %w", err)
	}

	return apiSNC.Results(sr, r.prefix,
		r.mCtx.GetResultsOutputPath(),
		r.allocationData.SpotPrice,
		r.disableClusterReadiness)
}

func (r *openshiftSNCRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", amiName(r.version, r.arch)),
		[]string{"self", amiOwner},
		map[string]string{
			"architecture": *r.arch})
	if err != nil {
		return err
	}
	nw, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:             *r.prefix,
			ID:                 apiSNC.OCPSNCID,
			Region:             *r.allocationData.Region,
			AZ:                 *r.allocationData.AZ,
			CreateLoadBalancer: r.allocationData.SpotPrice != nil,
			Airgap:             false,
		})
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, apiSNC.OCPSNCID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	if r.mCtx.Debug() {
		keyResources.PrivateKey.PrivateKeyPem.ApplyT(
			func(privateKey string) (*string, error) {
				logging.Debugf("%s", privateKey)
				return nil, nil
			})
	}
	// Security groups
	securityGroups, err := securityGroups(ctx, r.mCtx, r.prefix, nw.Vpc)
	if err != nil {
		return err
	}
	// Instance profile required by logic within userdata
	iProfile, err := iam.InstanceProfile(ctx, r.prefix, &apiSNC.OCPSNCID, requiredPolicies)
	if err != nil {
		return err
	}
	// Userdata
	udB64, kaPass, devPass, udDependecies, err := r.userData(ctx, &keyResources.PrivateKey.PublicKeyOpenssh, &nw.Eip.PublicIp)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputKubeAdminPass),
		kaPass)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputDeveloperPass),
		devPass)
	// Create instance
	cr := compute.ComputeRequest{
		MCtx:             r.mCtx,
		Prefix:           *r.prefix,
		ID:               apiSNC.OCPSNCID,
		VPC:              nw.Vpc,
		Subnet:           nw.Subnet,
		AMI:              ami,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		LB:               nw.LoadBalancer,
		Eip:              nw.Eip,
		LBTargetGroups:   []int{securityGroup.SSH_PORT, apiSNC.PortHTTPS, apiSNC.PortAPI},
		InstanceProfile:  iProfile,
		UserDataAsBase64: udB64,
		DependsOn:        udDependecies,
	}
	if r.spot && r.allocationData.SpotPrice != nil {
		cr.SpotPrice = *r.allocationData.SpotPrice
		cr.Spot = true
	}

	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputUsername),
		pulumi.String(amiUserDefault))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputHost),
		c.GetHostIP(true))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			apiSNC.OCPSNCID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"openshift-snc",
				r.mCtx.ProjectName(),
				r.mCtx.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	// Use kubeconfig as the readiness for the cluster
	kubeconfig, err := kubeconfig(ctx, r.prefix, c, keyResources.PrivateKey, *r.version, r.disableClusterReadiness)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, apiSNC.OutputKubeconfig),
		pulumi.ToSecret(kubeconfig))
	return nil
}

// security group for Openshift
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix *string,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*prefix, apiSNC.OCPSNCID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", apiSNC.OCPSNCID),
		IngressRules: []securityGroup.IngressRules{securityGroup.SSH_TCP,
			{Description: "Console", FromPort: apiSNC.PortHTTPS, ToPort: apiSNC.PortHTTPS, Protocol: "tcp"},
			{Description: "API", FromPort: apiSNC.PortAPI, ToPort: apiSNC.PortAPI, Protocol: "tcp"}},
	}.Create(ctx, mCtx)
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

func checkAMIExists(ctx context.Context, amiName, region, arch *string) error {
	isAMIOffered, _, err := data.IsAMIOffered(
		ctx,
		data.ImageRequest{
			Name:   amiName,
			Arch:   arch,
			Region: region,
			Owner:  &amiOwner})
	if err != nil {
		return err
	}
	if !isAMIOffered {
		return fmt.Errorf("AMI %s could not be found in region: %s", *amiName, *region)
	}
	return nil
}

func (r *openshiftSNCRequest) userData(ctx *pulumi.Context,
	newPublicKey, lbEIP *pulumi.StringOutput,
) (pulumi.StringPtrInput, pulumi.StringInput, pulumi.StringInput, []pulumi.Resource, error) {
	// Resources
	dependecies := []pulumi.Resource{}
	// Manage pull secret
	ps, err := os.ReadFile(*r.pullSecretFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	psString := string(ps)
	psName, psParam, err := ssm.AddSSM(ctx, r.prefix, &ocpPullSecretID, &psString)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, psParam)
	// KubeAdmin pass
	kaPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, apiSNC.OCPSNCID, "kubeadminpassword"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	kaPassName, kaPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &kapass, kaPassword.Result)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, kaPassParam)
	// Developer pass
	devPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, apiSNC.OCPSNCID, "devpassword"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	devPassName, devPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &devpass, devPassword.Result)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, devPassParam)
	ccB64 := pulumi.All(newPublicKey, lbEIP).ApplyT(
		func(args []interface{}) (string, error) {
			ccB64, err := apiSNC.CloudConfig(apiSNC.DataValues{
				Username:                 amiUserDefault,
				PubKey:                   args[0].(string),
				PublicIP:                 args[1].(string),
				SSMPullSecretName:        *psName,
				SSMKubeAdminPasswordName: *kaPassName,
				SSMDeveloperPasswordName: *devPassName})
			return *ccB64, err
		}).(pulumi.StringOutput)

	return ccB64, kaPassword.Result, devPassword.Result, dependecies, err
}

func kubeconfig(ctx *pulumi.Context,
	prefix *string,
	c *compute.Compute, mk *tls.PrivateKey,
	ocpVersion string,
	disableClusterReadiness bool,
) (pulumi.StringOutput, error) {
	// Once the cluster setup is comleted we
	// get the kubeconfig file from the host running the cluster
	// then we replace the internal access with the public IP
	// the resulting kubeconfig file can be used to access the cluster

	// Check SSH connectivity first
	sshReadyCmd, err := c.RunCommand(ctx,
		command.CommandPing,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ssh-readiness", *prefix), apiSNC.OCPSNCID,
		mk, amiUserDefault, nil, c.Dependencies)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Check cluster is ready
	ocpReadyCmd, err := c.RunCommand(ctx,
		util.If(disableClusterReadiness, apiSNC.CommandKubeconfigExists, apiSNC.CommandCrcReadiness),
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ocp-readiness", *prefix), apiSNC.OCPSNCID,
		mk, amiUserDefault, nil, []pulumi.Resource{sshReadyCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	// Check ocp-cluster-ca.service succeeds
	ocpCaRotatedCmd, err := c.RunCommand(ctx,
		apiSNC.CommandCaServiceRan(ocpVersion),
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ocp-ca-rotated", *prefix), apiSNC.OCPSNCID,
		mk, amiUserDefault, nil, []pulumi.Resource{ocpReadyCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Get content for /opt/kubeconfig
	getKCCmd := ("sudo cat /opt/crc/kubeconfig")
	getKC, err := c.RunCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), apiSNC.OCPSNCID, mk, amiUserDefault,
		nil, []pulumi.Resource{ocpCaRotatedCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, c.Eip.PublicIp).ApplyT(
		func(args []interface{}) string {
			return strings.ReplaceAll(args[0].(string),
				"https://api.crc.testing:6443",
				fmt.Sprintf("https://api.%s.nip.io:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}
