package macm1

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const vncDefaultPort int = 5900

func (r *MacM1Request) GetRequest() *compute.Request {
	return &r.Request
}

func (r *MacM1Request) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *MacM1Request) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *MacM1Request) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return ec2.NewDedicatedHost(ctx,
		r.GetName(),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(r.AvailabilityZones[0]),
			InstanceType:     pulumi.String(r.Specs.InstaceTypes[0]),
		})
}

func (r *MacM1Request) CustomIngressRules() []securityGroup.IngressRules {
	return []securityGroup.IngressRules{
		{
			Description: fmt.Sprintf("VNC port for %s", r.Specs.ID),
			FromPort:    vncDefaultPort,
			ToPort:      vncDefaultPort,
			Protocol:    "tcp",
			CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
		},
	}
}

func (r *MacM1Request) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *MacM1Request) PostProcess(ctx *pulumi.Context, compute *compute.Compute) ([]pulumi.Resource, error) {
	password, err := utilInfra.CreatePassword(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputPassword(), password.Result)
	postscript := password.Result.ApplyT(func(password string) (string, error) {
		return util.Template(
			scriptDataValues{
				r.Specs.AMI.DefaultUser,
				password},
			"postscript", script)

	}).(pulumi.StringOutput)
	waitCmddependencies := []pulumi.Resource{}
	rc, err := compute.RemoteExec(ctx,
		postscript,
		fmt.Sprintf("%s-%s", r.Specs.ID, "postscript"),
		nil)
	if err != nil {
		return nil, err
	}
	waitCmddependencies = append(waitCmddependencies, rc)
	return waitCmddependencies, nil
}

func (r *MacM1Request) ReadinessCommand() string {
	return r.Request.ReadinessCommand()
}

func (r *MacM1Request) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}
