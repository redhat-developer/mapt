package windowsserver

import (
	_ "embed"
	"encoding/base64"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type userDataValues struct {
	Username             string
	Password             string
	AuthorizedKey        string
	ActionsRunnerSnippet string
	RunnerToken          string
	CirrusSnippet        string
	CirrusToken          string
}

//go:embed bootstrap.ps1
var BootstrapScript []byte

// function to template userdata script to be executed on boot
func Userdata(ctx *pulumi.Context, amiUser *string, password *random.RandomPassword,
	keypair *keypair.KeyPairResources) (pulumi.StringPtrInput, error) {
	udBase64 := pulumi.All(password.Result, keypair.PrivateKey.PublicKeyOpenssh).ApplyT(
		func(args []interface{}) (string, error) {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			cirrusSnippet, err := integrations.GetIntegrationSnippet(cirrus.GetRunnerArgs(), *amiUser)
			if err != nil {
				return "", err
			}
			ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippet(github.GetRunnerArgs(), *amiUser)
			if err != nil {
				return "", err
			}
			udv := userDataValues{
				*amiUser,
				password,
				authorizedKey,
				*ghActionsRunnerSnippet,
				github.GetToken(),
				*cirrusSnippet,
				cirrus.GetToken(),
			}
			userdata, err := file.Template(udv, string(BootstrapScript[:]))
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
		}).(pulumi.StringOutput)
	return udBase64, nil
}
