package setup

import (
	_ "embed"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
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
	cirrusSnippet, err := cirrus.PersistentWorkerSnippet(username)
	if err != nil {
		return "", err
	}
	ghActionsRunnerSnippet, err := github.SelfHostedRunnerSnippet(username)
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
			*cirrusSnippet},
		string(RequestScript[:]))
}
