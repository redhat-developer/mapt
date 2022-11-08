package windows

import (
	"encoding/base64"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/util"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *WindowsRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *WindowsRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *WindowsRequest) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	password, err := utilInfra.CreatePassword(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	ctx.Export(r.OutputPassword(), password.Result)
	udBase64 := pulumi.All(password.Result, r.PublicKeyOpenssh).ApplyT(
		func(args []interface{}) string {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			userdata, _ := util.Template(
				userDataValues{
					r.Specs.AMI.DefaultUser,
					password,
					authorizedKey},
				fmt.Sprintf("%s-%s", "userdata", r.GetName()),
				userdata)
			return base64.StdEncoding.EncodeToString([]byte(userdata))
		}).(pulumi.StringOutput)
	return udBase64, nil
}

func (r *WindowsRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *WindowsRequest) CustomIngressRules() []securityGroup.IngressRules {
	return []securityGroup.IngressRules{
		securityGroup.RDP_TCP}
}

func (r *WindowsRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *WindowsRequest) GetPostScript(ctx *pulumi.Context) (string, error) {
	return "", nil
}

func (r *WindowsRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.Request.Create(ctx, r)
}
