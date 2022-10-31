package rhel

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *WindowsRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *WindowsRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *WindowsRequest) GetUserdata() (pulumi.StringPtrInput, error) {

	// https://charlesxu.io/wiki/infra-as-code/pulumi/
	// https://www.pulumi.com/registry/packages/random/api-docs/randompassword/?utm_source=performance-max&utm_medium=cpc&utm_campaign=&utm_term=&utm_medium=ppc&utm_source=adwords&hsa_grp=&hsa_cam=18353585506&hsa_mt=&hsa_net=adwords&hsa_ver=3&hsa_acc=1926559913&hsa_ad=&hsa_src=x&hsa_tgt=&hsa_kw=&gclid=EAIaIQobChMIwP3C2sqK-wIVPY1oCR0EOgJoEAAYASAAEgJM6vD_BwE
	// t := pulumi.All(r.KeyPair.Arn).ApplyT(
	// 	func(args []interface{}) string {
	// 		return args[0].(string)
	// 	}).(pulumi.StringOutput)

	// return t, nil

	// st := pulumi.String("lalal")

	// return st, nil
	return nil, nil
}

func (r *WindowsRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *WindowsRequest) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *WindowsRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *WindowsRequest) GetPostScript() (string, error) {
	return "", nil
}

func (r *WindowsRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}

// var cloudConfig string = `
// #cloud-config
// rh_subscription:
//   username: {{.SubscriptionUsername}}
//   password: {{.SubscriptionPassword}}
//   auto-attach: true
// packages:
//   - podman
// `

// type UserDataValues struct {
// 	SubscriptionUsername string
// 	SubscriptionPassword string
// }
