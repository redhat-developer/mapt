package macm1

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"fmt"

	infra "github.com/adrianriobo/qenvs/pkg/provider"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/provider/util/security"
	"github.com/adrianriobo/qenvs/pkg/util/file"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const vncDefaultPort int = 5900

func (r *Request) GetRequest() *compute.Request {
	return &r.Request
}

func (r *Request) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	amiNameRegex := fmt.Sprintf(r.Specs.AMI.RegexPattern, r.VersionMajor)
	return ami.GetAMIByName(ctx, amiNameRegex, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *Request) GetDiskSize() int {
	return r.Request.GetDiskSize()
}

func (r *Request) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *Request) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return ec2.NewDedicatedHost(ctx,
		r.GetName(),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(r.AvailabilityZones[0]),
			InstanceType:     pulumi.String(r.Specs.InstaceTypes[0]),
		})
}

func (r *Request) CustomIngressRules() []securityGroup.IngressRules {
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

func (r *Request) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *Request) GetPostScript(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	password, err := security.CreatePassword(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputPassword(), password.Result)
	postscript := password.Result.ApplyT(func(password string) (string, error) {
		return file.Template(
			scriptDataValues{
				r.Specs.AMI.DefaultUser,
				password},
			"postscript", script)

	}).(pulumi.StringOutput)
	return postscript, nil
}

func (r *Request) ReadinessCommand() string {
	return r.Request.ReadinessCommand()
}

func (r *Request) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}
