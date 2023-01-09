package snc

import (
	"encoding/base64"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/infra/util/command"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *SNCRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *SNCRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *SNCRequest) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	userdataTemplate, err := getUserdataTemplate()
	if err != nil {
		return nil, err
	}
	userdata, err := util.Template(
		userDataValues{
			r.SubscriptionUsername,
			r.SubscriptionPassword},
		"userdata", string(userdataTemplate))
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}

func (r *SNCRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *SNCRequest) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *SNCRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *SNCRequest) GetPostScript(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *SNCRequest) ReadinessCommand() string {
	// return command.CommandCloudInitWait
	return command.CommandPing
}

func (r *SNCRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.RHELRequest.Request.Create(ctx, r)
}
