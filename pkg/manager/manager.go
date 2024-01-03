package manager

import (
	"context"
	"os"

	"github.com/adrianriobo/qenvs/pkg/manager/credentials"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

func UpStack(targetStack Stack, opts ...ManagerOptions) (auto.UpResult, error) {
	logging.Debugf("Creating stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := getStack(ctx, targetStack)
	// TODO add when loglevel debug control in place
	w := logging.GetWritter()
	defer w.Close()
	stdoutStreamer := optup.ProgressStreams(w)
	r, err := objectStack.Up(ctx, stdoutStreamer)
	if err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return auto.UpResult{}, err
		}
		os.Exit(1)
	}
	return r, nil
}

func DestroyStack(targetStack Stack, opts ...ManagerOptions) (err error) {
	logging.Debugf("Destroying stack %s", targetStack.StackName)
	ctx := context.Background()
	objectStack := getStack(ctx, targetStack)
	w := logging.GetWritter()
	defer w.Close()
	stdoutStreamer := optdestroy.ProgressStreams(w)
	if _, err := objectStack.Destroy(ctx, stdoutStreamer); err != nil {
		logging.Error(err)
		os.Exit(1)
	}
	if err := objectStack.Workspace().RemoveStack(ctx, targetStack.StackName); err != nil {
		logging.Error(err)
		if len(opts) == 1 && opts[0].Baground {
			return err
		}
		os.Exit(1)
	}
	return nil
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

func GetOutputs(stack *auto.Stack) (auto.OutputMap, error) {
	return stack.Outputs(context.Background())
}
