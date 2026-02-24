package integrations

import (
	cloudinit "github.com/redhat-developer/mapt/pkg/util/cloud-init"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

type UserDataValues struct {
	CliURL   string
	User     string
	Name     string
	Token    string
	Labels   string
	Port     string
	RepoURL  string
	Executor string
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

// If we add the snippet as part of a cloud init file the strategy
// would be create the file with write_files:
// i.e.
// write_files:
//
//	# Cirrus service setup
//	- content: |
//	    {{ .CirrusSnippet }} <----- 6 spaces
//
// to do so we need to indent 6 spaces each line of the snippet
func GetIntegrationSnippetAsCloudInitWritableFile(intCfg IntegrationConfig, username string) (*string, error) {
	snippet, err := GetIntegrationSnippet(intCfg, username)
	if err != nil || len(*snippet) == 0 {
		return snippet, err
	}
	return cloudinit.IndentWriteFile(snippet)
}
