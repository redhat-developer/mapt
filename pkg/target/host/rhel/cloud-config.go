package rhel

import (
	_ "embed"
	"encoding/base64"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type CloudConfigArgs struct {
	SNCProfile                 bool
	SubsUsername, SubsPassword string
	Username                   string
}

type userDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
	Username             string
	ActionsRunnerSnippet string
	CirrusSnippet        string
	GitLabSnippet        string
}

//go:embed cloud-config-base
var CloudConfigBase []byte

//go:embed cloud-config-snc
var CloudConfigSNC []byte

func (r *CloudConfigArgs) CloudConfig() (*string, error) {
	templateConfig := string(CloudConfigBase[:])
	if r.SNCProfile {
		templateConfig = string(CloudConfigSNC[:])
	}
	cirrusSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(cirrus.GetRunnerArgs(), r.Username)
	if err != nil {
		return nil, err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), r.Username)
	if err != nil {
		return nil, err
	}
	gitlabSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(gitlab.GetRunnerArgs(), r.Username)
	if err != nil {
		return nil, err
	}
	userdata, err := file.Template(
		userDataValues{
			r.SubsUsername,
			r.SubsPassword,
			r.Username,
			*ghActionsRunnerSnippet,
			*cirrusSnippet,
			*gitlabSnippet},
		templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(userdata))
	return &ccB64, nil
}

// CloudConfigWithGitLabToken generates cloud config with GitLab auth token set via ApplyT
func (r *CloudConfigArgs) CloudConfigWithGitLabToken(gitlabAuthToken string) (string, error) {
	// Set auth token in global state temporarily
	gitlab.SetAuthToken(gitlabAuthToken)
	defer gitlab.SetAuthToken("") // Clear after use

	templateConfig := string(CloudConfigBase[:])
	if r.SNCProfile {
		templateConfig = string(CloudConfigSNC[:])
	}
	cirrusSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(cirrus.GetRunnerArgs(), r.Username)
	if err != nil {
		return "", err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), r.Username)
	if err != nil {
		return "", err
	}
	gitlabSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(gitlab.GetRunnerArgs(), r.Username)
	if err != nil {
		return "", err
	}
	userdata, err := file.Template(
		userDataValues{
			r.SubsUsername,
			r.SubsPassword,
			r.Username,
			*ghActionsRunnerSnippet,
			*cirrusSnippet,
			*gitlabSnippet},
		templateConfig)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
}

// GenerateCloudConfig generates cloud config for RHEL instances.
// If GitLab runner args are present, it creates the runner via Pulumi first and uses the auth token.
// Otherwise, it generates normal cloud config without GitLab integration.
func (r *CloudConfigArgs) GenerateCloudConfig(ctx *pulumi.Context, runID string) (pulumi.StringInput, error) {
	glRunnerArgs := gitlab.GetRunnerArgs()

	if glRunnerArgs != nil {
		glRunnerArgs.Name = runID

		authToken, err := gitlab.CreateRunner(ctx, glRunnerArgs)
		if err != nil {
			return nil, err
		}

		// Generate cloud config with auth token using ApplyT
		return authToken.ApplyT(func(token string) (string, error) {
			return r.CloudConfigWithGitLabToken(token)
		}).(pulumi.StringOutput), nil
	}

	// No GitLab runner, use normal cloud config
	udB64, err := r.CloudConfig()
	if err != nil {
		return nil, err
	}
	return pulumi.String(*udB64), nil
}
