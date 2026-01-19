package setup

import (
	_ "embed"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

//go:embed release.sh
var ReleaseScript []byte

//go:embed request.sh
var RequestScript []byte

type releaseDataValues struct {
	Username      string
	Password      string
	AuthorizedKey string
}

type requestDataValues struct {
	Username             string
	OldPassword          string
	NewPassword          string
	AuthorizedKey        string
	ActionsRunnerSnippet string
	CirrusSnippet        string
	GitLabSnippet        string
}

func Release(username, pass, authorizedKey string) (string, error) {
	return file.Template(
		releaseDataValues{
			username,
			pass,
			authorizedKey},
		string(ReleaseScript[:]))
}

func Request(username, oldPassword, newPassword, authorizedKey string) (string, error) {
	cirrusSnippet, err := integrations.GetIntegrationSnippet(cirrus.GetRunnerArgs(), username)
	if err != nil {
		return "", err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippet(github.GetRunnerArgs(), username)
	if err != nil {
		return "", err
	}
	gitlabSnippet, err := integrations.GetIntegrationSnippet(gitlab.GetRunnerArgs(), username)
	if err != nil {
		return "", err
	}
	return file.Template(
		requestDataValues{
			username,
			oldPassword,
			newPassword,
			authorizedKey,
			*ghActionsRunnerSnippet,
			*cirrusSnippet,
			*gitlabSnippet},
		string(RequestScript[:]))
}

// GenerateBootstrapScript generates bootstrap script for Mac instances.
// If GitLab runner args are present, it creates the runner via Pulumi first and uses the auth token.
// Otherwise, it generates normal bootstrap script without GitLab integration.
func GenerateBootstrapScript(
	ctx *pulumi.Context,
	password *random.RandomPassword,
	publicKeyOpenssh pulumi.StringOutput,
	runID string,
	isRequestOperation bool,
	currentPassword string,
	username string,
) (pulumi.StringOutput, error) {
	glRunnerArgs := gitlab.GetRunnerArgs()

	if glRunnerArgs != nil {
		// Create GitLab runner via Pulumi
		glRunnerArgs.Name = runID
		authToken, err := gitlab.CreateRunner(ctx, glRunnerArgs)
		if err != nil {
			return pulumi.StringOutput{}, err
		}

		// Include GitLab token in ApplyT
		return pulumi.All(password.Result, publicKeyOpenssh, authToken).ApplyT(
			func(args []interface{}) (string, error) {
				passwordVal := args[0].(string)
				authorizedKey := args[1].(string)
				gitlabToken := args[2].(string)

				// Set GitLab token in global state
				gitlab.SetAuthToken(gitlabToken)
				defer gitlab.SetAuthToken("")

				if isRequestOperation {
					return Request(username, currentPassword, passwordVal, authorizedKey)
				}
				return Release(username, passwordVal, authorizedKey)
			}).(pulumi.StringOutput), nil
	}

	// No GitLab runner, use normal flow
	return pulumi.All(password.Result, publicKeyOpenssh).ApplyT(
		func(args []interface{}) (string, error) {
			passwordVal := args[0].(string)
			authorizedKey := args[1].(string)

			if isRequestOperation {
				return Request(username, currentPassword, passwordVal, authorizedKey)
			}
			return Release(username, passwordVal, authorizedKey)
		}).(pulumi.StringOutput), nil
}
