package fedora

import (
	_ "embed"
	"encoding/base64"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type userDataValues struct {
	Username             string
	ActionsRunnerSnippet string
	GitLabSnippet        string
}

//go:embed cloud-config
var CloudConfig []byte

func Userdata(amiUser string) (pulumi.StringPtrInput, error) {
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), amiUser)
	if err != nil {
		return nil, err
	}
	gitlabSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(gitlab.GetRunnerArgs(), amiUser)
	if err != nil {
		return nil, err
	}

	templateConfig := string(CloudConfig[:])
	userdata, err := file.Template(
		userDataValues{
			amiUser,
			*ghActionsRunnerSnippet,
			*gitlabSnippet},
		templateConfig)
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}

// UserdataWithGitLabToken generates userdata with the GitLab auth token set.
// This is used when the token is obtained asynchronously via Pulumi ApplyT.
func UserdataWithGitLabToken(amiUser string, gitlabAuthToken string) (string, error) {
	// Set auth token in global state temporarily
	gitlab.SetAuthToken(gitlabAuthToken)
	defer gitlab.SetAuthToken("") // Clear after use

	// Generate userdata as normal
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), amiUser)
	if err != nil {
		return "", err
	}
	gitlabSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(gitlab.GetRunnerArgs(), amiUser)
	if err != nil {
		return "", err
	}

	templateConfig := string(CloudConfig[:])
	userdata, err := file.Template(
		userDataValues{
			amiUser,
			*ghActionsRunnerSnippet,
			*gitlabSnippet},
		templateConfig)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
}

// GenerateUserdata generates userdata for Fedora instances.
// If GitLab runner args are present, it creates the runner via Pulumi first and uses the auth token.
// Otherwise, it generates normal userdata without GitLab integration.
func GenerateUserdata(ctx *pulumi.Context, amiUser string, runID string) (pulumi.StringPtrInput, error) {
	if ghRunnerArgs := github.GetRunnerArgs(); ghRunnerArgs != nil {
		if err := github.SetupRunner(ctx, ghRunnerArgs); err != nil {
			return nil, err
		}
		github.ExportRunnerOutputs(ctx)
	}

	glRunnerArgs := gitlab.GetRunnerArgs()

	if glRunnerArgs != nil {
		glRunnerArgs.Name = runID

		authToken, err := gitlab.CreateRunner(ctx, glRunnerArgs)
		if err != nil {
			return nil, err
		}

		// Generate userdata with auth token using ApplyT
		return authToken.ApplyT(func(token string) (string, error) {
			return UserdataWithGitLabToken(amiUser, token)
		}).(pulumi.StringOutput), nil
	}

	// No GitLab runner, use normal userdata
	return Userdata(amiUser)
}
