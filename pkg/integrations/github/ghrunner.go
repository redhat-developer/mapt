package github

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	pulgithub "github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

var runnerVersion = "2.317.0"

// 1 is version, 2 is platform: (win, linux, osx), 3 is arch: (arm64, x64, arm)
const runnerBaseURL = "https://github.com/actions/runner/releases/download/v%[1]s/actions-runner-%[2]s-%[3]s-%[1]s"

const runnerImageRepo = "https://github.com/aipcc-cicd/action-runner-image-pz.git"
const runnerImageRepoVersion = "v2.0.0"

//go:embed snippet-darwin.sh
var snippetDarwin []byte

//go:embed snippet-linux.sh
var snippetLinux []byte

//go:embed snippet-windows.ps1
var snippetWindows []byte

//go:embed snippet-linux-ppc64le.sh
var snippetLinuxPpc64le []byte

//go:embed snippet-linux-s390x.sh
var snippetLinuxS390x []byte

var snippets map[Platform][]byte = map[Platform][]byte{
	Darwin:  snippetDarwin,
	Linux:   snippetLinux,
	Windows: snippetWindows,
}

var archSnippets map[Arch][]byte = map[Arch][]byte{
	Ppc64le: snippetLinuxPpc64le,
	S390x:   snippetLinuxS390x,
}

var runnerArgs *GithubRunnerArgs

func Init(args *GithubRunnerArgs) {
	runnerArgs = args
}

func (args *GithubRunnerArgs) GetUserDataValues() *integrations.UserDataValues {
	if args == nil {
		return nil
	}
	repoURL := args.RepoURL
	if args.Org != "" {
		repoURL = "https://github.com/" + args.Org
	}
	return &integrations.UserDataValues{
		Name:                   args.Name,
		Token:                  args.Token,
		Labels:                 getLabels(),
		RepoURL:                repoURL,
		CliURL:                 downloadURL(),
		RunnerImageRepo:        runnerImageRepo,
		RunnerImageRepoVersion: runnerImageRepoVersion,
	}
}

func (args *GithubRunnerArgs) GetSetupScriptTemplate() string {
	if *runnerArgs.Platform == Linux && runnerArgs.Arch != nil {
		if archSnippet, ok := archSnippets[*runnerArgs.Arch]; ok {
			return string(archSnippet[:])
		}
	}
	return string(snippets[*runnerArgs.Platform][:])
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

// SetupRunner fetches a runner registration token from the GitHub App and
// sets it on args. It is a no-op when args is nil, Token is already set,
// or AppID is empty (PAT / plain-token paths are handled in params).
func SetupRunner(ctx *pulumi.Context, args *GithubRunnerArgs) error {
	if args == nil || args.AppID == "" || args.Token != "" {
		return nil
	}

	pemBytes, err := os.ReadFile(args.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("reading GitHub App private key: %w", err)
	}

	var owner string
	if args.Org != "" {
		owner = args.Org
	} else {
		owner, _, err = splitOwnerRepo(args.RepoURL)
		if err != nil {
			return err
		}
	}

	provider, err := pulgithub.NewProvider(ctx, "github-app-provider", &pulgithub.ProviderArgs{
		Owner: pulumi.String(owner),
		AppAuth: pulgithub.ProviderAppAuthPtr(&pulgithub.ProviderAppAuthArgs{
			Id:             pulumi.String(args.AppID),
			InstallationId: pulumi.String(args.InstallationID),
			PemFile:        pulumi.String(string(pemBytes)),
		}),
	})
	if err != nil {
		return fmt.Errorf("creating GitHub App provider: %w", err)
	}

	if args.Org != "" {
		result, err := pulgithub.GetActionsOrganizationRegistrationToken(ctx, pulumi.Provider(provider))
		if err != nil {
			return fmt.Errorf("fetching org runner registration token: %w", err)
		}
		args.Token = result.Token
	} else {
		_, repo, _ := splitOwnerRepo(args.RepoURL)
		result, err := pulgithub.GetActionsRegistrationToken(ctx,
			&pulgithub.GetActionsRegistrationTokenArgs{Repository: repo},
			pulumi.Provider(provider))
		if err != nil {
			return fmt.Errorf("fetching runner registration token: %w", err)
		}
		args.Token = result.Token
	}

	logging.Info("runner registration token generated from GitHub App successfully")
	return nil
}
