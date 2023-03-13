package rhel

import (
	"encoding/base64"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/util"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *RHELRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *RHELRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	amiNameRegex := fmt.Sprintf(r.Specs.AMI.RegexPattern, r.VersionMajor)
	return ami.GetAMIByName(ctx, amiNameRegex, "", r.Specs.AMI.Filters)
}

func (r *RHELRequest) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	userdata, err := util.Template(
		UserDataValues{
			r.SubscriptionUsername,
			r.SubscriptionPassword},
		"userdata", cloudConfig)
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}

func (r *RHELRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *RHELRequest) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *RHELRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *RHELRequest) GetPostScript(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *RHELRequest) ReadinessCommand() string {
	return r.Request.ReadinessCommand()
}

func (r *RHELRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}

var cloudConfig string = `
#cloud-config  
rh_subscription:
  username: {{.SubscriptionUsername}}
  password: {{.SubscriptionPassword}}
  auto-attach: true
packages:
  - podman
`

type UserDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
}
