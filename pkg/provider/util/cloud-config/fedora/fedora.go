package fedora

import (
	_ "embed"
	"encoding/base64"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type userDataValues struct {
	Username             string
	ActionsRunnerSnippet string
	CirrusSnippet        string
}

//go:embed cloud-config
var CloudConfig []byte

func Userdata(amiUser string) (pulumi.StringPtrInput, error) {
	cirrusSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(cirrus.GetRunnerArgs(), amiUser)
	if err != nil {
		return nil, err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), amiUser)
	if err != nil {
		return nil, err
	}

	templateConfig := string(CloudConfig[:])
	userdata, err := file.Template(
		userDataValues{
			amiUser,
			*ghActionsRunnerSnippet,
			*cirrusSnippet},
		templateConfig)
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}
