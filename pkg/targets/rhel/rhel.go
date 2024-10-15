package rhel

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
)

type userDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
	Username             string
	InstallActionsRunner bool
	ActionsRunnerSnippet string
}

//go:embed cloud-config-base
var CloudConfigBase []byte

//go:embed cloud-config-snc
var CloudConfigSNC []byte

func GetUserdata(sncProfile bool, subsUsername, subsPassword string,
	username string, ghActionRunner bool) (string, error) {
	templateConfig := string(CloudConfigBase[:])
	if sncProfile {
		templateConfig = string(CloudConfigSNC[:])
	}
	userdata, err := file.Template(
		userDataValues{
			subsUsername,
			subsPassword,
			username,
			ghActionRunner,
			ghactions.GetActionRunnerSnippetLinux()},
		templateConfig)
	// return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
	return base64.StdEncoding.EncodeToString([]byte(userdata)), err
}
