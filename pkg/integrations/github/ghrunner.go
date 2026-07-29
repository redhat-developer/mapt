package github

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
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
		Ephemeral:              args.Ephemeral,
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

const (
	outputRunnerName = "gh-runner-name"
	outputRunnerOrg  = "gh-runner-org"
	outputRunnerRepo = "gh-runner-repo"
)

// GetManagementToken returns a GitHub token suitable for runner management.
// For GitHub App auth it generates a fresh installation access token.
// For PAT auth it reads GITHUB_TOKEN from the environment.
func GetManagementToken() (string, error) {
	if runnerArgs != nil && runnerArgs.AppID != "" {
		return GenerateInstallationToken(runnerArgs.AppID, runnerArgs.InstallationID, runnerArgs.PrivateKeyPath)
	}
	if pat := os.Getenv("GITHUB_TOKEN"); pat != "" {
		return pat, nil
	}
	return "", fmt.Errorf("no GitHub credentials: set GITHUB_TOKEN or use --ghactions-app-* flags")
}

// ExportRunnerOutputs writes the runner name and target (org or repo URL) as
// Pulumi stack outputs. Call this from deploy() after SetupRunner so the
// destroy path can read them back without re-parsing CLI flags.
func ExportRunnerOutputs(ctx *pulumi.Context) {
	if runnerArgs == nil {
		return
	}
	ctx.Export(outputRunnerName, pulumi.String(runnerArgs.Name))
	if runnerArgs.Org != "" {
		ctx.Export(outputRunnerOrg, pulumi.String(runnerArgs.Org))
	} else {
		ctx.Export(outputRunnerRepo, pulumi.String(runnerArgs.RepoURL))
	}
}

// TryDeregister removes the runner from GitHub using pat and the stack outputs
// previously written by ExportRunnerOutputs. It is best-effort: errors are
// logged as warnings so the caller's destroy always proceeds.
// outputs is the raw value map from auto.OutputMap (key → output.Value).
func TryDeregister(pat string, outputs map[string]interface{}) {
	if pat == "" {
		logging.Warn("GITHUB_TOKEN not set, skipping GitHub runner deregistration")
		return
	}
	name, _ := outputs[outputRunnerName].(string)
	if name == "" {
		return
	}
	if org, _ := outputs[outputRunnerOrg].(string); org != "" {
		if err := DeleteOrgRunner(pat, org, name); err != nil {
			logging.Warnf("GitHub runner deregistration failed: %v", err)
			return
		}
		logging.Infof("GitHub runner %q deregistered from org %q", name, org)
		return
	}
	if repo, _ := outputs[outputRunnerRepo].(string); repo != "" {
		if err := DeleteRepoRunner(pat, repo, name); err != nil {
			logging.Warnf("GitHub runner deregistration failed: %v", err)
			return
		}
		logging.Infof("GitHub runner %q deregistered from repo %q", name, repo)
	}
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

	provider, err := github.NewProvider(ctx, "github-app-provider",
		&github.ProviderArgs{
			Owner: pulumi.String(owner),
			AppAuth: github.ProviderAppAuthArgs{
				Id:             pulumi.String(args.AppID),
				InstallationId: pulumi.String(args.InstallationID),
				PemFile:        pulumi.ToSecret(pulumi.String(string(pemBytes))).(pulumi.StringInput),
			},
		})

	if err != nil {
		return fmt.Errorf("creating GitHub App provider: %w", err)
	}

	if args.Org != "" {
		result, err := github.GetActionsOrganizationRegistrationToken(ctx, pulumi.Provider(provider))
		if err != nil {
			return fmt.Errorf("fetching org runner registration token: %w", err)
		}
		args.Token = result.Token
	} else {
		_, repo, err := splitOwnerRepo(args.RepoURL)
		if err != nil {
			return err
		}
		result, err := github.GetActionsRegistrationToken(ctx,
			&github.GetActionsRegistrationTokenArgs{Repository: repo},
			pulumi.Provider(provider))
		if err != nil {
			return fmt.Errorf("fetching runner registration token: %w", err)
		}
		args.Token = result.Token
	}

	logging.Info("runner registration token generated from GitHub App successfully")
	return nil
}
