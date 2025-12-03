package manager

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/debug"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type Stack struct {
	ProjectName         string
	StackName           string
	BackedURL           string
	DeployFunc          pulumi.RunFunc
	ProviderCredentials credentials.ProviderCredentials
}

type ManagerOptions struct {
	// This option informs the manager the actions will be run on background
	// through a routine so in that case we can not return exit but an error
	Baground bool
}

func UpStack(c *mc.Context, targetStack Stack, opts ...ManagerOptions) (auto.UpResult, error) {
	return UpStackTargets(c, targetStack, nil, opts...)
}

func UpStackTargets(mCtx *mc.Context, targetStack Stack, targetURNs []string, opts ...ManagerOptions) (auto.UpResult, error) {
	logging.Debugf("managing stack %s", targetStack.StackName)
	ctx := mCtx.Context()

	objectStack, err := getStack(ctx, mCtx, targetStack)
	if err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return auto.UpResult{}, err
		}
		return auto.UpResult{}, fmt.Errorf("failed to get stack: %w", err)
	}

	w := logging.GetWritter()
	defer func() {
		if err := w.Close(); err != nil {
			logging.Error(err)
		}
	}()

	mOpts := []optup.Option{
		optup.ProgressStreams(w),
	}
	if mCtx.Debug() {
		dl := mCtx.DebugLevel()
		mOpts = append(mOpts, optup.DebugLogging(debug.LoggingOptions{
			LogLevel:      &dl,
			Debug:         true,
			FlowToPlugins: true,
			LogToStdErr:   true,
		}))
	}
	if len(targetURNs) > 0 {
		mOpts = append(mOpts, optup.Target(targetURNs))
	}

	result, err := objectStack.Up(ctx, mOpts...)
	if err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return auto.UpResult{}, err
		}
		return auto.UpResult{}, fmt.Errorf("failed to update stack: %w", err)
	}

	return result, nil
}

func DestroyStack(mCtx *mc.Context, targetStack Stack, opts ...ManagerOptions) error {
	logging.Debugf("destroying stack %s", targetStack.StackName)
	ctx := mCtx.Context()

	objectStack, err := getStack(ctx, mCtx, targetStack)
	if err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return err
		}
		return fmt.Errorf("failed to get stack: %w", err)
	}

	w := logging.GetWritter()
	defer func() {
		if err := w.Close(); err != nil {
			logging.Error(err)
		}
	}()

	mOpts := []optdestroy.Option{
		optdestroy.ProgressStreams(w),
	}
	if mCtx.Debug() {
		dl := mCtx.DebugLevel()
		mOpts = append(mOpts, optdestroy.DebugLogging(
			debug.LoggingOptions{
				LogLevel:      &dl,
				FlowToPlugins: true,
				LogToStdErr:   true}))
	}

	// Destroy resources
	if _, err := objectStack.Destroy(ctx, mOpts...); err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return err
		}
		return fmt.Errorf("failed to destroy stack resources: %w", err)
	}

	// Remove the stack from workspace
	if err := objectStack.Workspace().RemoveStack(ctx, targetStack.StackName); err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return err
		}
		return fmt.Errorf("failed to remove stack from workspace: %w", err)
	}

	return nil
}

func CheckStack(mCtx *mc.Context, target Stack) (*auto.Stack, error) {
	logging.Debugf("checking stack %s", target.StackName)
	stack, err := auto.SelectStackInlineSource(mCtx.Context(), target.StackName,
		target.ProjectName, target.DeployFunc, getOpts(target)...)
	if err != nil {
		return nil, err
	}
	return &stack, err
}

func GetOutputs(mCtx *mc.Context, stack *auto.Stack) (auto.OutputMap, error) {
	return stack.Outputs(mCtx.Context())
}
