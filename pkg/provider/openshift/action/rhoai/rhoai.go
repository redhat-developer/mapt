package rhoai

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/openshift"
	"github.com/redhat-developer/mapt/pkg/provider/openshift/profile"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	StackName     = "stackOpenshiftRHOAI"
	defaultPrefix = "rhoai"
)

type RHOAIArgs struct {
	KubeconfigPath string
	Profiles       []string
	Prefix         string
}

type rhoaiRequest struct {
	mCtx           *mc.Context
	prefix         string
	kubeconfigPath string
	profiles       []string
}

func Create(mCtxArgs *mc.ContextArgs, args *RHOAIArgs) error {
	mCtx, err := mc.Init(mCtxArgs, openshift.Provider())
	if err != nil {
		return err
	}
	if err := profile.Validate(args.Profiles); err != nil {
		return err
	}
	prefix := args.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}
	r := &rhoaiRequest{
		mCtx:           mCtx,
		prefix:         prefix,
		kubeconfigPath: args.KubeconfigPath,
		profiles:       args.Profiles,
	}
	return r.deploy()
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Run openshift rhoai destroy")
	mCtx, err := mc.Init(mCtxArgs, openshift.Provider())
	if err != nil {
		return err
	}
	return openshift.DestroyStack(mCtx, StackName)
}

func (r *rhoaiRequest) deploy() error {
	cs := manager.Stack{
		StackName:           r.mCtx.StackNameByProject(StackName),
		ProjectName:         r.mCtx.ProjectName(),
		BackedURL:           r.mCtx.BackedURL(),
		ProviderCredentials: openshift.NoCredentials,
		DeployFunc:          r.pulumiProgram,
	}
	_, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return fmt.Errorf("stack creation failed: %w", err)
	}
	return nil
}

func (r *rhoaiRequest) pulumiProgram(ctx *pulumi.Context) error {
	kcBytes, err := os.ReadFile(r.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("reading kubeconfig from %s: %w", r.kubeconfigPath, err)
	}
	kubeconfig := pulumi.String(string(kcBytes)).ToStringOutput()

	k8sProvider, err := profile.NewK8sProvider(ctx, "k8s-provider", kubeconfig)
	if err != nil {
		return err
	}

	return profile.Deploy(ctx, r.profiles, &profile.DeployArgs{
		K8sProvider: k8sProvider,
		Kubeconfig:  kubeconfig,
		Prefix:      r.prefix,
	})
}
