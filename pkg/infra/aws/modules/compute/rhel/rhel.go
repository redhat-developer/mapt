package rhel

import (
	"fmt"

	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *RHELRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	amiNameRegex := fmt.Sprintf(r.Specs.AMI.RegexPattern, r.VersionMajor)
	return ami.GetAMIByName(ctx, amiNameRegex, "", r.Specs.AMI.Filters)
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

func (r *RHELRequest) GetPostScript() (string, error) {
	return "", nil
}
