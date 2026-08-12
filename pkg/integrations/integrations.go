package integrations

import (
	cloudinit "github.com/redhat-developer/mapt/pkg/util/cloud-init"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type UserDataValues struct {
	CliURL         string
	User           string
	Name           string
	Token          string
	Labels         string
	Port           string
	RepoURL        string
	Executor       string
	Unsecure       bool
	Concurrent     int
	LogToJournald  bool
	Ephemeral      bool
	RunnerImageRepo        string
	RunnerImageRepoVersion string
}

type IntegrationConfig interface {
	GetUserDataValues() *UserDataValues
	GetSetupScriptTemplate() string
}

func GetIntegrationSnippet(intCfg IntegrationConfig, username string) (*string, error) {
	userDataValues := intCfg.GetUserDataValues()
	if userDataValues == nil {
		noSnippet := ""
		return &noSnippet, nil
	}
	userDataValues.User = username
	snippet, err := file.Template(userDataValues, intCfg.GetSetupScriptTemplate())
	return &snippet, err
}

// GetIntegrationSnippetAsCloudInitWritableFile wraps GetIntegrationSnippet
// and indents the result for use inside a cloud-init write_files block.
func GetIntegrationSnippetAsCloudInitWritableFile(intCfg IntegrationConfig, username string) (*string, error) {
	snippet, err := GetIntegrationSnippet(intCfg, username)
	if err != nil || len(*snippet) == 0 {
		return snippet, err
	}
	return cloudinit.IndentWriteFile(snippet)
}
