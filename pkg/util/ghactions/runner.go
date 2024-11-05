package ghactions

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/mapt/pkg/util"
)

type RunnerArgs struct {
	Token   string
	RepoURL string
	Name    string
	Labels  []string
}

const (
	runnerVersion = "2.317.0"

	runnerBaseURLWin   = "https://github.com/actions/runner/releases/download/v%[1]s/actions-runner-win-x64-%[1]s.zip"
	runnerBaseURLLinux = "https://github.com/actions/runner/releases/download/v%[1]s/actions-runner-linux-x64-%[1]s.tar.gz"
	runnerBaseURLMacos = "https://github.com/actions/runner/releases/download/v%[1]s/actions-runner-osx-x64-%[1]s.tar.gz"

	// $ghToken needs to be set externally before use; it is defined in the platform specific setup scripts
	// for aws this is defined in the script and for azure it is passed as an arg to the setup script
	installActionRunnerSnippetWindows string = `New-Item -Path C:\actions-runner -Type Directory ; cd C:\actions-runner
Invoke-WebRequest -Uri %s -OutFile actions-runner-win.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win.zip", "$PWD")
./config.cmd --token $ghToken --url %s --name %s --unattended --runasservice --replace %s`

	// whitespace at the start is required since this is expanded in a cloud-init yaml file
	// to start as service need to relable the runsvc.sh file on rhel: https://github.com/actions/runner/issues/3222
	installActionRunnerSnippetLinux string = `  mkdir ~/actions-runner && cd ~/actions-runner` + "\n" +
		`      curl -o actions-runner-linux.tar.gz -L %s` + "\n" +
		`      tar xzf ./actions-runner-linux.tar.gz` + "\n" +
		`      sudo ./bin/installdependencies.sh` + "\n" +
		`      ./config.sh --token %s --url %s --name %s --unattended --replace %s` + "\n" +
		`      sudo ./svc.sh install` + "\n" +
		`      chcon system_u:object_r:usr_t:s0 $(pwd)/runsvc.sh` + "\n" +
		`      sudo ./svc.sh start`

	installActionRunnerSnippetMacos string = `mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-osx.tar.gz -L %s
tar xzf ./actions-runner-osx.tar.gz
./config.sh --token %s --url %s --name %s --unattended --replace %s
./svc.sh install
plistName=$(basename $(./svc.sh status | grep "plist$"))
mkdir -p ~/Library/LaunchDaemons
mv ~/Library/LaunchAgents/"${plistName}" ~/Library/LaunchDaemons/"${plistName}"
launchctl load ~/Library/LaunchDaemons/"${plistName}"`
)

var args *RunnerArgs

func InitGHRunnerArgs(token, name, repoURL string, labels []string) error {
	if token == "" || name == "" || repoURL == "" {
		return errors.New("All args are required and must have non-empty values")
	}
	args = &RunnerArgs{
		Token:   token,
		RepoURL: repoURL,
		Name:    name,
	}

	if len(labels) > 0 {
		args.Labels = labels
	}
	return nil
}

func GetToken() string {
	var token = func() string {
		return args.Token
	}
	return util.IfNillable(args != nil, token, "")
}

func GetLabels() string {
	var labels = func() string {
		if len(args.Labels) > 0 {
			return fmt.Sprintf("--labels %s", strings.Join(args.Labels, ","))
		}
		return ""
	}
	return util.IfNillable(args != nil, labels, "")
}

func GetActionRunnerSnippetWin() string {
	var snippetWindows = func() string {
		return fmt.Sprintf(installActionRunnerSnippetWindows,
			fmt.Sprintf(runnerBaseURLWin, runnerVersion), args.RepoURL, args.Name, GetLabels())
	}
	return util.IfNillable(args != nil, snippetWindows, "")
}

func GetActionRunnerSnippetLinux() string {
	var snippetLinux = func() string {
		return fmt.Sprintf(installActionRunnerSnippetLinux,
			fmt.Sprintf(runnerBaseURLLinux, runnerVersion), args.Token, args.RepoURL, args.Name, GetLabels())
	}
	return util.IfNillable(args != nil, snippetLinux, "")
}

func GetActionRunnerSnippetMacos() string {
	var snippetMacos = func() string {
		return fmt.Sprintf(installActionRunnerSnippetMacos,
			fmt.Sprintf(runnerBaseURLMacos, runnerVersion), args.Token, args.RepoURL, args.Name, GetLabels())
	}
	return util.IfNillable(args != nil, snippetMacos, "")
}
