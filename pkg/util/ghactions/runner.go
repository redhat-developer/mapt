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

var args *RunnerArgs

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
Invoke-WebRequest -Uri https://github.com/actions/runner/releases/download/v2.317.0/actions-runner-win-x64-2.317.0.zip -OutFile actions-runner-win-x64-2.317.0.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
if((Get-FileHash -Path actions-runner-win-x64-2.317.0.zip -Algorithm SHA256).Hash.ToUpper() -ne 'a74dcd1612476eaf4b11c15b3db5a43a4f459c1d3c1807f8148aeb9530d69826'.ToUpper()){ throw 'Computed checksum did not match' }
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win-x64-2.317.0.zip", "$PWD")
./config.cmd --token $ghToken --url %s --name %s --unattended --runasservice --replace`

// whitespace at the start is required since this is expanded in a cloud-init yaml file
// to start as service need to relable the runsvc.sh file on rhel: https://github.com/actions/runner/issues/3222
const LinuxActionsRunnerInstallSnippet string = `  mkdir ~/actions-runner && cd ~/actions-runner` + "\n" +
	`      curl -o actions-runner-linux-x64-2.317.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.317.0/actions-runner-linux-x64-2.317.0.tar.gz` + "\n" +
	`      echo "9e883d210df8c6028aff475475a457d380353f9d01877d51cc01a17b2a91161d  actions-runner-linux-x64-2.317.0.tar.gz" | sha256sum -c` + "\n" +
	`      tar xzf ./actions-runner-linux-x64-2.317.0.tar.gz` + "\n" +
	`      sudo ./bin/installdependencies.sh` + "\n" +
	`      ./config.sh --token %s --url %s --name %s --unattended --replace` + "\n" +
	`      sudo ./svc.sh install` + "\n" +
	`      chcon system_u:object_r:usr_t:s0 $(pwd)/runsvc.sh` + "\n" +
	`      sudo ./svc.sh start`

func GetActionRunnerSnippetWin() string {
	if (args == &RunnerArgs{}) {
		return ""
	}
	return fmt.Sprintf(WindowsActionsRunnerInstallSnippet, args.RepoURL, args.Name)
}

func GetActionRunnerSnippetLinux() string {
	if (args == &RunnerArgs{}) {
		return ""
	}
	return fmt.Sprintf(LinuxActionsRunnerInstallSnippet, args.Token, args.RepoURL, args.Name)
}
