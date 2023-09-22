package fedora

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/security-group"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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
	return nil, nil
}

func (r *Request) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *Request) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *Request) GetPostScript(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	return nil, nil
}

func (r *Request) ReadinessCommand() string {
	return r.Request.ReadinessCommand()
}

func (r *Request) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}
