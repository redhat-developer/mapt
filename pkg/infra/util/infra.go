package util

import (
	"context"
	"os"
	"path/filepath"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
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

func UpStack(targetStack Stack) (auto.UpResult, error) {
	logging.Debugf("Creating stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := getStack(ctx, targetStack)
	// TODO add when loglevel debug control in place
	w := logging.GetWritter()
	defer w.Close()
	stdoutStreamer := optup.ProgressStreams(w)
	return objectStack.Up(ctx, stdoutStreamer)
}

func DestroyStack(targetStack Stack) (err error) {
	logging.Debugf("Destroying stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := getStack(ctx, targetStack)
	w := logging.GetWritter()
	defer w.Close()
	stdoutStreamer := optdestroy.ProgressStreams(w)
	if _, err = objectStack.Destroy(ctx, stdoutStreamer); err != nil {
		return
	}
	err = objectStack.Workspace().RemoveStack(ctx, targetStack.StackName)
	return
}

func CheckStack(target Stack) (*auto.Stack, error) {
	logging.Debugf("Checking stack %s", target.StackName)
	stack, err := auto.SelectStackInlineSource(context.Background(), target.StackName,
		target.ProjectName, target.DeployFunc, getOpts(target)...)
	if err != nil {
		return nil, err
	}
	return &stack, err
}

func GetOutputs(stack auto.Stack) (auto.OutputMap, error) {
	return stack.Outputs(context.Background())
}

// this function gets our stack ready for update/destroy by prepping the workspace, init/selecting the stack
// and doing a refresh to make sure state and cloud resources are in sync
func getStack(ctx context.Context, target Stack) auto.Stack {
	// create or select a stack with an inline Pulumi program
	s, err := auto.UpsertStackInlineSource(ctx, target.StackName,
		target.ProjectName, target.DeployFunc, getOpts(target)...)
	if err != nil {
		logging.Errorf("Failed to create or select stack: %v", err)
		os.Exit(1)
	}
	if err = postStack(ctx, target, &s); err != nil {
		logging.Error(err)
		os.Exit(1)
	}
	return s
}

func getOpts(target Stack) []auto.LocalWorkspaceOption {
	return []auto.LocalWorkspaceOption{
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
}

func postStack(ctx context.Context, target Stack, stack *auto.Stack) (err error) {
	w := stack.Workspace()
	// for inline source programs, we must manage plugins ourselves
	if err = w.InstallPlugin(ctx, target.Plugin.Name, target.Plugin.Version); err != nil {
		return
	}
	// Set credentials
	if err = target.Plugin.SetCredentialFunc(ctx, *stack, target.Plugin.FixedCredentials); err != nil {
		return
	}
	_, err = stack.Refresh(ctx)
	return
}
