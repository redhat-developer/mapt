package openspotng

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

func (r *OpenspotNGRequest) GetPostScript(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	password, err := utilInfra.CreatePassword(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputPassword(), password.Result)
	postscript := password.Result.ApplyT(func(password string) (string, error) {
		return util.Template(
			scriptDataValues{},
			"postscript", script)

	}).(pulumi.StringOutput)
	return postscript, nil
}

func (r *OpenspotNGRequest) ReadinessCommand() string {
	return r.Request.ReadinessCommand()
}

func (r *OpenspotNGRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}
