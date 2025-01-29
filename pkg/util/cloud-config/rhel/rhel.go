package rhel

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type RequestArgs struct {
	SNCProfile                 bool
	SubsUsername, SubsPassword string
	Username                   string
	GHActionRunner             bool
}

type userDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
	Username             string
	InstallActionsRunner bool
	ActionsRunnerSnippet string
	CirrusSnippet        string
}

//go:embed cloud-config-base
var CloudConfigBase []byte

//go:embed cloud-config-snc
var CloudConfigSNC []byte

func (r *RequestArgs) GetAsUserdata() (string, error) {
	templateConfig := string(CloudConfigBase[:])
	if r.SNCProfile {
		templateConfig = string(CloudConfigSNC[:])
	}
	cirrusSnippet, err := cirrus.PersistentWorkerSnippetAsCloudInitWritableFile(r.Username)
	if err != nil {
		return "", err
	}
	userdata, err := file.Template(
		userDataValues{
			r.SubsUsername,
			r.SubsPassword,
			r.Username,
			r.GHActionRunner,
			github.GetActionRunnerSnippetLinux(),
			*cirrusSnippet},
		templateConfig)
	// return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
	return base64.StdEncoding.EncodeToString([]byte(userdata)), err
}
