package github

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
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

type snippetDataValues struct {
	Username  string
	Token     string
	RepoURL   string
	Name      string
	Labels    string
	RunnerURL string
}

var runnerArgs *GithubRunnerArgs

func Init(args *GithubRunnerArgs) {
	runnerArgs = args
}

func SelfHostedRunnerSnippet(username string) (*string, error) {
	if runnerArgs == nil {
		noSnippet := ""
		return &noSnippet, nil
	}
	templateConfig := string(snippets[*runnerArgs.Platform][:])
	snippet, err := file.Template(
		snippetDataValues{
			Name:      runnerArgs.Name,
			Token:     runnerArgs.Token,
			Labels:    GetLabels(),
			RepoURL:   runnerArgs.RepoURL,
			RunnerURL: downloadURL(),
			Username:  username,
		},
		templateConfig)
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
func SelfHostedRunnerSnippetAsCloudInitWritableFile(username string) (*string, error) {
	snippet, err := SelfHostedRunnerSnippet(username)
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

func GetLabels() string {
	var labels = func() string {
		if len(runnerArgs.Labels) > 0 {
			return strings.Join(runnerArgs.Labels, ",")
		}
		return ""
	}
	return util.IfNillable(runnerArgs != nil, labels, "")
}
