package ghactions

import (
	"fmt"

	"github.com/pkg/errors"
)

type RunnerArgs struct {
	Token   string
	RepoURL string
	Name    string
}

const runnerVersion = "2.317.0"

var (
	args *RunnerArgs

	runnerBaseURLWin   = "https://github.com/actions/runner/releases/download/v%s/actions-runner-win-x64-%s.zip"
	runnerBaseURLLinux = "https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-x64-%s.tar.gz"
)

func getRunnerDownloadURLWin() string {
	return fmt.Sprintf(runnerBaseURLWin, runnerVersion, runnerVersion)
}

func getRunnerDownloadURLLinux() string {
	return fmt.Sprintf(runnerBaseURLLinux, runnerVersion, runnerVersion)
}

func InitGHRunnerArgs(token, name, repoURL string) error {
	if token == "" || name == "" || repoURL == "" {
		return errors.New("All args are required and must have non-empty values")
	}
	args = &RunnerArgs{
		Token:   token,
		RepoURL: repoURL,
		Name:    name,
	}
	return nil
}

func GetToken() string {
	if (args == &RunnerArgs{}) {
		return ""
	}
	return args.Token
}

// $ghToken needs to be set externally before use; it is defined in the platform specific setup scripts
// for aws this is defined in the script and for azure it is passed as an arg to the setup script
const WindowsActionsRunnerInstallSnippet string = `New-Item -Path C:\actions-runner -Type Directory ; cd C:\actions-runner
Invoke-WebRequest -Uri %s -OutFile actions-runner-win.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win.zip", "$PWD")
./config.cmd --token $ghToken --url %s --name %s --unattended --runasservice --replace`

// whitespace at the start is required since this is expanded in a cloud-init yaml file
// to start as service need to relable the runsvc.sh file on rhel: https://github.com/actions/runner/issues/3222
const LinuxActionsRunnerInstallSnippet string = `  mkdir ~/actions-runner && cd ~/actions-runner` + "\n" +
	`      curl -o actions-runner-linux.tar.gz -L %s` + "\n" +
	`      tar xzf ./actions-runner-linux.tar.gz` + "\n" +
	`      sudo ./bin/installdependencies.sh` + "\n" +
	`      ./config.sh --token %s --url %s --name %s --unattended --replace` + "\n" +
	`      sudo ./svc.sh install` + "\n" +
	`      chcon system_u:object_r:usr_t:s0 $(pwd)/runsvc.sh` + "\n" +
	`      sudo ./svc.sh start`

func GetActionRunnerSnippetWin() string {
	if (args == &RunnerArgs{}) {
		return ""
	}
	return fmt.Sprintf(WindowsActionsRunnerInstallSnippet,
		getRunnerDownloadURLWin(), args.RepoURL, args.Name)
}

func GetActionRunnerSnippetLinux() string {
	if (args == &RunnerArgs{}) {
		return ""
	}
	return fmt.Sprintf(LinuxActionsRunnerInstallSnippet,
		getRunnerDownloadURLLinux(), args.Token, args.RepoURL, args.Name)
}
