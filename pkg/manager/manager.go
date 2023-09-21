package manager

import (
	"context"

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
