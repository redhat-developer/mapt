package manager

import (
	"context"
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
	Background bool
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
		if len(opts) == 1 && opts[0].Background {
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
		if len(opts) == 1 && opts[0].Background {
			return auto.UpResult{}, err
		}
		return auto.UpResult{}, fmt.Errorf("failed to update stack: %w", err)
	}

	return result, nil
}

func DestroyStack(mCtx *mc.Context, targetStack Stack, opts ...ManagerOptions) error {
	logging.Debugf("destroying stack %s", targetStack.StackName)
	ctx := mCtx.Context()

	objectStack, err := getStackForDestroy(ctx, mCtx, targetStack)
	if err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Background {
			return err
		}
		return fmt.Errorf("failed to get stack: %w", err)
	}
	return DestroySelectedStack(mCtx, &objectStack, DestroyOptions{
		Background: len(opts) == 1 && opts[0].Background,
	})
}

// DestroyOptions controls an opt-in phased teardown. Targeted phases retain the
// stack so later phases continue from the same state.
type DestroyOptions struct {
	Targets          []string
	TargetDependents bool
	ContinueOnError  bool
	// Background returns the destroy error unwrapped, as the caller runs the
	// destroy in a routine and reports the error itself.
	Background bool
}

// SelectStackForDestroy opens existing state without refreshing providers or
// creating an empty stack when the requested stack does not exist.
func SelectStackForDestroy(mCtx *mc.Context, target Stack) (*auto.Stack, error) {
	program := target.DeployFunc
	if program == nil {
		// Targeted destroy still resolves the project runtime. A no-op program
		// makes the Automation API serve the inline language runtime, so the
		// CLI does not try to load a Go program that does not exist.
		program = func(*pulumi.Context) error { return nil }
	}
	s, err := auto.SelectStackInlineSource(mCtx.Context(), target.StackName,
		target.ProjectName, program, getOpts(target)...)
	if err != nil {
		return nil, err
	}
	if err := setStackCredentials(mCtx.Context(), mCtx, target, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func DestroySelectedStack(mCtx *mc.Context, stack *auto.Stack, options DestroyOptions) error {
	ctx := mCtx.Context()

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

	return destroyPhase(ctx, stack, func(ctx context.Context) error {
		return stack.Workspace().RemoveStack(ctx, stack.Name())
	}, options, mOpts...)
}

type stackDestroyer interface {
	Destroy(context.Context, ...optdestroy.Option) (auto.DestroyResult, error)
}

func destroyPhase(ctx context.Context, stack stackDestroyer, remove func(context.Context) error, options DestroyOptions, opts ...optdestroy.Option) error {
	if len(options.Targets) > 0 {
		opts = append(opts, optdestroy.Target(options.Targets))
		if options.TargetDependents {
			opts = append(opts, optdestroy.TargetDependents())
		}
	}
	if options.ContinueOnError {
		opts = append(opts, optdestroy.ContinueOnError())
	}
	if _, err := stack.Destroy(ctx, opts...); err != nil {
		logging.Error(err)
		if options.Background {
			return err
		}
		return fmt.Errorf("failed to destroy stack resources: %w", err)
	}
	if len(options.Targets) > 0 {
		return nil
	}
	if err := remove(ctx); err != nil {
		logging.Error(err)
		if options.Background {
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
