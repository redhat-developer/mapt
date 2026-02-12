package windowsserver

import (
	_ "embed"
	"encoding/base64"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
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
	GitLabSnippet        string
	GitLabToken          string
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
			gitlabSnippet, err := integrations.GetIntegrationSnippet(gitlab.GetRunnerArgs(), *amiUser)
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
				*gitlabSnippet,
				gitlab.GetToken(),
			}
			userdata, err := file.Template(udv, string(BootstrapScript[:]))
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
		}).(pulumi.StringOutput)
	return udBase64, nil
}

// UserdataWithGitLabToken generates userdata with GitLab auth token via ApplyT
func UserdataWithGitLabToken(ctx *pulumi.Context, amiUser *string, password *random.RandomPassword,
	keypair *keypair.KeyPairResources, gitlabAuthToken pulumi.StringOutput) (pulumi.StringPtrInput, error) {
	udBase64 := pulumi.All(password.Result, keypair.PrivateKey.PublicKeyOpenssh, gitlabAuthToken).ApplyT(
		func(args []interface{}) (string, error) {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			gitlabToken := args[2].(string)

			// Set GitLab token in global state
			gitlab.SetAuthToken(gitlabToken)
			defer gitlab.SetAuthToken("")

			cirrusSnippet, err := integrations.GetIntegrationSnippet(cirrus.GetRunnerArgs(), *amiUser)
			if err != nil {
				return "", err
			}
			ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippet(github.GetRunnerArgs(), *amiUser)
			if err != nil {
				return "", err
			}
			gitlabSnippet, err := integrations.GetIntegrationSnippet(gitlab.GetRunnerArgs(), *amiUser)
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
				*gitlabSnippet,
				gitlabToken,
			}
			userdata, err := file.Template(udv, string(BootstrapScript[:]))
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
		}).(pulumi.StringOutput)
	return udBase64, nil
}

// GenerateUserdata generates userdata for Windows instances.
// If GitLab runner args are present, it creates the runner via Pulumi first and uses the auth token.
// Otherwise, it generates normal userdata without GitLab integration.
func GenerateUserdata(ctx *pulumi.Context, amiUser *string, password *random.RandomPassword,
	keypair *keypair.KeyPairResources, runID string) (pulumi.StringPtrInput, error) {
	glRunnerArgs := gitlab.GetRunnerArgs()

	if glRunnerArgs != nil {
		glRunnerArgs.Name = runID

		authToken, err := gitlab.CreateRunner(ctx, glRunnerArgs)
		if err != nil {
			return nil, err
		}

		// Generate userdata with auth token using ApplyT
		return UserdataWithGitLabToken(ctx, amiUser, password, keypair, authToken)
	}

	// No GitLab runner, use normal userdata
	return Userdata(ctx, amiUser, password, keypair)
}
