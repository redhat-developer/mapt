package rhel

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
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
	userdata, err := file.Template(
		userDataValues{
			r.SubsUsername,
			r.SubsPassword,
			r.Username,
			*ghActionsRunnerSnippet,
			*cirrusSnippet},
		templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(userdata))
	return &ccB64, nil
}
