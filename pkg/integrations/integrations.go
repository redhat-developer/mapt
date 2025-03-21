package integrations

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type UserDataValues struct {
	CliURL  string
	User    string
	Name    string
	Token   string
	Labels  string
	Port    string
	RepoURL string
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
	lines := strings.Split(strings.TrimSpace(*snippet), "\n")
	for i, line := range lines {
		// Added 6 spaces before each line
		lines[i] = fmt.Sprintf("      %s", line)
	}
	identedSnippet := strings.Join(lines, "\n")
	return &identedSnippet, nil
}
