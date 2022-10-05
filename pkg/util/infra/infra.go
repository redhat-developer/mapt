package infra

import (
	"context"
	"os"
	"path/filepath"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SetCredentials func(ctx context.Context, stack auto.Stack, fixedCredentials map[string]string) error

type PluginInfo struct {
	Name              string
	Version           string
	SetCredentialFunc SetCredentials
	FixedCredentials  map[string]string
}

type Stack struct {
	ProjectName string
	StackName   string
	BackedURL   string
	DeployFunc  pulumi.RunFunc
	Plugin      PluginInfo
}

// this function gets our stack ready for update/destroy by prepping the workspace, init/selecting the stack
// and doing a refresh to make sure state and cloud resources are in sync
func GetStack(ctx context.Context, target Stack) auto.Stack {
	opts := []auto.LocalWorkspaceOption{
		auto.Project(workspace.Project{
			Name:    tokens.PackageName(target.ProjectName),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{
				URL: target.BackedURL,
			},
		}),
		auto.WorkDir(filepath.Join(".")),
		// auto.SecretsProvider("awskms://alias/pulumi-secret-encryption"),
	}

	// create or select a stack with an inline Pulumi program
	s, err := auto.UpsertStackInlineSource(ctx, target.StackName, target.ProjectName, target.DeployFunc, opts...)
	if err != nil {
		logging.Errorf("Failed to create or select stack: %v", err)
		os.Exit(1)
	}

	w := s.Workspace()

	// for inline source programs, we must manage plugins ourselves
	if err = w.InstallPlugin(ctx, target.Plugin.Name, target.Plugin.Version); err != nil {
		logging.Errorf("Failed to install program plugins: %v", err)
		os.Exit(1)
	}
	// Set credentials
	if err = target.Plugin.SetCredentialFunc(ctx, s, target.Plugin.FixedCredentials); err != nil {
		logging.Errorf("Failed setting credentials: %v", err)
		os.Exit(1)
	}
	_, err = s.Refresh(ctx)
	if err != nil {
		logging.Errorf("Failed to refresh stack: %v\n", err)
		os.Exit(1)
	}
	return s
}

func UpStack(targetStack Stack) (auto.UpResult, error) {
	logging.Debugf("Creating stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := GetStack(ctx, targetStack)
	// TODO add when loglevel debug control in place
	// stdoutStreamer := optup.ProgressStreams(os.Stdout)
	// return objectStack.Up(ctx, stdoutStreamer)
	return objectStack.Up(ctx)
}

func DestroyStack(targetStack Stack) (auto.DestroyResult, error) {
	logging.Debugf("Destroying stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := GetStack(ctx, targetStack)
	// TODO add when loglevel debug control in place
	// stdoutStreamer := optup.ProgressStreams(os.Stdout)
	// return objectStack.Up(ctx, stdoutStreamer)
	return objectStack.Destroy(ctx)
}
