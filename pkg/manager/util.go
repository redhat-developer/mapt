package manager

import (
	"context"
	"os"
	"path/filepath"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

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
	if err = w.InstallPlugin(ctx, target.CloudProviderPlugin.Name, target.CloudProviderPlugin.Version); err != nil {
		return
	}
	// Set credentials
	if err = target.CloudProviderPlugin.SetCredentialFunc(ctx, *stack, target.CloudProviderPlugin.FixedCredentials); err != nil {
		return
	}
	_, err = stack.Refresh(ctx)
	return
}
