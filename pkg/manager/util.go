package manager

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// this function gets our stack ready for update/destroy by prepping the workspace, init/selecting the stack
// and doing a refresh to make sure state and cloud resources are in sync
func getStack(ctx context.Context, target Stack) (auto.Stack, error) {
	// create or select a stack with an inline Pulumi program
	s, err := auto.UpsertStackInlineSource(ctx,
		target.StackName,
		target.ProjectName,
		target.DeployFunc,
		getOpts(target)...)
	if err != nil {
		logging.Errorf("Failed to create or select stack: %v", err)
		return auto.Stack{}, err
	}

	if err = postStack(ctx, target, &s); err != nil {
		logging.Error(err)
		return auto.Stack{}, err
	}

	return s, nil
}

func getOpts(target Stack) []auto.LocalWorkspaceOption {
	// Build the work dir path: ./<stack-name>
	workDir := filepath.Join(".", target.StackName)

	// Ensure the directory exists
	if err := os.MkdirAll(workDir, 0755); err != nil {
		logging.Fatalf("Failed to create work directory %q: %v", workDir, err)
	}

	return []auto.LocalWorkspaceOption{
		auto.Project(workspace.Project{
			Name:    tokens.PackageName(target.ProjectName),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{
				URL: target.BackedURL,
			},
		}),
		auto.WorkDir(workDir),
	}
}

func postStack(ctx context.Context, target Stack, stack *auto.Stack) (err error) {
	// Set credentails
	if err = credentials.SetProviderCredentials(ctx, stack, target.ProviderCredentials); err != nil {
		return
	}
	_, err = stack.Refresh(ctx)
	return
}
