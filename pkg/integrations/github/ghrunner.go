package github

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/util"
)

var runnerVersion = "2.317.0"

// 1 is version, 2 is platform: (win, linux, osx), 3 is arch: (arm64, x64, arm)
const runnerBaseURL = "https://github.com/actions/runner/releases/download/v%[1]s/actions-runner-%[2]s-%[3]s-%[1]s"

//go:embed snippet-darwin.sh
var snippetDarwin []byte

//go:embed snippet-linux.sh
var snippetLinux []byte

//go:embed snippet-windows.ps1
var snippetWindows []byte

var snippets map[Platform][]byte = map[Platform][]byte{
	Darwin:  snippetDarwin,
	Linux:   snippetLinux,
	Windows: snippetWindows,
}

var runnerArgs *GithubRunnerArgs

func Init(args *GithubRunnerArgs) {
	runnerArgs = args
}

func (args *GithubRunnerArgs) GetUserDataValues() *integrations.UserDataValues {
	return &integrations.UserDataValues{
		Name:    args.Name,
		Token:   args.Token,
		Labels:  getLabels(),
		RepoURL: args.RepoURL,
		CliURL:  downloadURL(),
	}
}

func (args *GithubRunnerArgs) GetSetupScriptTemplate() string {
	templateConfig := string(snippets[*runnerArgs.Platform][:])
	return templateConfig
}

func GetRunnerArgs() *GithubRunnerArgs {
	return runnerArgs
}

// platform: darwin, linux, windows
// arch: amd64, arm64, arm
func downloadURL() string {
	url := fmt.Sprintf(runnerBaseURL, runnerVersion, *runnerArgs.Platform, *runnerArgs.Arch)
	switch *runnerArgs.Platform {
	case Windows:
		url = fmt.Sprintf("%s.zip", url)
	case Linux, Darwin:
		url = fmt.Sprintf("%s.tar.gz", url)
	}
	return url
}

func GetToken() string {
	var token = func() string {
		return runnerArgs.Token
	}
	return util.IfNillable(runnerArgs != nil, token, "")
}

func getLabels() string {
	var labels = func() string {
		if len(runnerArgs.Labels) > 0 {
			return strings.Join(runnerArgs.Labels, ",")
		}
		return ""
	}
	return util.IfNillable(runnerArgs != nil, labels, "")
}
