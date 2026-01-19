package gitlab

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi-gitlab/sdk/v8/go/gitlab"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// GitLab Runner version
var version = "18.8.0"

// Download URL - %s placeholders: 1=version, 2=platform, 3=arch
const runnerBaseURL = "https://gitlab-runner-downloads.s3.amazonaws.com/v%s/binaries/gitlab-runner-%s-%s"

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

var runnerArgs *GitLabRunnerArgs

func Init(args *GitLabRunnerArgs) {
	runnerArgs = args
}

func (args *GitLabRunnerArgs) GetUserDataValues() *integrations.UserDataValues {
	if args == nil {
		return nil
	}
	return &integrations.UserDataValues{
		Name:    args.Name,
		Token:   args.AuthToken, // Use auth token (set by Pulumi during deployment)
		CliURL:  downloadURL(),
		RepoURL: args.URL,
	}
}

func (args *GitLabRunnerArgs) GetSetupScriptTemplate() string {
	templateConfig := string(snippets[*runnerArgs.Platform][:])
	return templateConfig
}

func GetRunnerArgs() *GitLabRunnerArgs {
	return runnerArgs
}

// platform: darwin, linux, windows
// arch: amd64, arm64, arm
func downloadURL() string {
	platform := string(*runnerArgs.Platform)
	arch := string(*runnerArgs.Arch)

	url := fmt.Sprintf(runnerBaseURL, version, platform, arch)
	if *runnerArgs.Platform == Windows {
		url = fmt.Sprintf("%s.exe", url)
	}
	return url
}

func GetToken() string {
	var token = func() string {
		return runnerArgs.AuthToken
	}
	return util.IfNillable(runnerArgs != nil, token, "")
}

// SetAuthToken sets the authentication token in the global runner args.
// This is called during Pulumi deployment after the runner is created.
func SetAuthToken(token string) {
	if runnerArgs != nil {
		runnerArgs.AuthToken = token
	}
}

// CreateRunner creates a GitLab UserRunner via Pulumi and returns auth token.
// This function should be called during Pulumi stack deployment (inside deploy() functions).
// Supports both project runners (with ProjectID) and group runners (with GroupID).
func CreateRunner(ctx *pulumi.Context, args *GitLabRunnerArgs) (pulumi.StringOutput, error) {
	if args.ProjectID != "" && args.GroupID != "" {
		return pulumi.StringOutput{}, fmt.Errorf("cannot specify both ProjectID and GroupID - use only one")
	}
	if args.ProjectID == "" && args.GroupID == "" {
		return pulumi.StringOutput{}, fmt.Errorf("must specify either ProjectID or GroupID")
	}

	// Convert tags to pulumi.StringArray
	tagArray := pulumi.StringArray{}
	for _, tag := range args.Tags {
		tagArray = append(tagArray, pulumi.String(tag))
	}

	// If no tags are provided, the runner can run untagged jobs
	runUntagged := len(args.Tags) == 0

	// Create common runner args with shared fields
	runnerArgs := &gitlab.UserRunnerArgs{
		Description: pulumi.String(args.Name),
		TagLists:    tagArray,
		Untagged:    pulumi.Bool(runUntagged),
		Locked:      pulumi.Bool(false),
		AccessLevel: pulumi.String("not_protected"),
	}

	// Set project or group specific fields
	var runnerType string
	if args.ProjectID != "" {
		projectID, err := strconv.Atoi(args.ProjectID)
		if err != nil {
			logging.Error(fmt.Sprintf("Failed to convert ProjectID '%s' to int: %v", args.ProjectID, err))
			return pulumi.StringOutput{}, err
		}
		logging.Debug(fmt.Sprintf("Creating GitLab project runner: URL=%s, ProjectID=%d, Name=%s, Tags=%v",
			args.URL, projectID, args.Name, args.Tags))

		runnerType = "project_type"
		runnerArgs.RunnerType = pulumi.String(runnerType)
		runnerArgs.ProjectId = pulumi.Int(projectID)
	} else {
		groupID, err := strconv.Atoi(args.GroupID)
		if err != nil {
			logging.Error(fmt.Sprintf("Failed to convert GroupID '%s' to int: %v", args.GroupID, err))
			return pulumi.StringOutput{}, err
		}
		logging.Debug(fmt.Sprintf("Creating GitLab group runner: URL=%s, GroupID=%d, Name=%s, Tags=%v",
			args.URL, groupID, args.Name, args.Tags))

		runnerType = "group_type"
		runnerArgs.RunnerType = pulumi.String(runnerType)
		runnerArgs.GroupId = pulumi.Int(groupID)
	}

	// Configure GitLab provider with PAT
	provider, err := gitlab.NewProvider(ctx, "gitlab-provider", &gitlab.ProviderArgs{
		Token:   pulumi.String(args.GitLabPAT),
		BaseUrl: pulumi.String(args.URL),
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Create UserRunner resource
	// This creates the runner in GitLab and returns an authentication token
	runner, err := gitlab.NewUserRunner(ctx, "gitlab-runner", runnerArgs, pulumi.Provider(provider))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Export runner ID for reference
	ctx.Export("gitlab-runner-id", runner.ID())
	ctx.Export("gitlab-runner-type", pulumi.String(runnerType))

	// Return the authentication token as a Pulumi output
	// This will be used with ApplyT to pass to userdata generation
	return runner.Token, nil
}
