package eks

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func destroyEKS(mCtx *mc.Context) error {
	target, err := aws.PrepareDestroyStack(mCtx, aws.DestroyStackRequest{Stackname: stackName})
	if err != nil {
		return err
	}
	stack, err := manager.SelectStackForDestroy(mCtx, target)
	if auto.IsSelectStack404Error(err) {
		return nil
	}
	if err != nil {
		return err
	}
	state, err := stack.Export(mCtx.Context())
	if err != nil {
		return err
	}
	identity, clusterURN, err := cleanupIdentityFromCheckpoint(state.Deployment)
	if err != nil {
		return err
	}
	switch {
	case identity != nil:
		logging.Infof("EKS teardown: account=%s region=%s cluster=%s vpc=%s", identity.AccountID, identity.Region, identity.ClusterName, identity.VPCID)
	case createCompleted(state.Deployment):
		// A finished create without a cluster resource means an earlier destroy
		// already removed the cluster, possibly without its controller resources.
		logging.Warn("EKS cluster is no longer in state; skipping controller resource cleanup. Load balancers from an earlier destroy may block VPC deletion")
	default:
		logging.Debug("no EKS cluster was created; skipping controller resource cleanup")
	}
	kubernetesTargets, err := kubernetesTeardownTargets(state.Deployment, clusterURN)
	if err != nil {
		return err
	}
	if len(kubernetesTargets) > 0 {
		logging.Debug("EKS teardown: removing Kubernetes resources and providers before the cluster")
		err := withDeleteUnreachable(stack.Workspace(), func() error {
			return manager.DestroySelectedStack(mCtx, stack, manager.DestroyOptions{
				Targets: kubernetesTargets, ContinueOnError: true,
			})
		})
		if err != nil {
			return fmt.Errorf("destroying kubernetes resources: %w", err)
		}
	}
	return runEKSTeardown(mCtx.Context(), clusterURN,
		func(targets []string) error {
			return manager.DestroySelectedStack(mCtx, stack, manager.DestroyOptions{
				Targets: targets, TargetDependents: true, ContinueOnError: true,
			})
		},
		func(ctx context.Context) error {
			if identity == nil {
				return nil
			}
			cleaner, err := newAWSCleaner(ctx, *identity)
			if err != nil {
				return err
			}
			return cleaner.cleanup(ctx)
		})
}

// The final phase is deliberately unreachable until both worker/cluster
// destruction and external-resource cleanup have succeeded.
func runEKSTeardown(ctx context.Context, clusterURN string, destroy func([]string) error, cleanup func(context.Context) error) error {
	if clusterURN != "" {
		if err := destroy([]string{clusterURN}); err != nil {
			return fmt.Errorf("destroying EKS cluster and dependents: %w", err)
		}
	}
	if err := cleanup(ctx); err != nil {
		return err
	}
	return destroy(nil)
}

const deleteUnreachableEnv = "PULUMI_K8S_DELETE_UNREACHABLE"

type workspaceEnv interface {
	SetEnvVar(string, string)
	UnsetEnvVar(string)
}

// withDeleteUnreachable lets the Kubernetes provider drop resources it can no
// longer reach, such as after its short-lived token expired. This is only safe
// while the cluster is still scheduled for deletion, which removes those
// resources anyway. The workspace environment reaches the CLI and its plugins.
func withDeleteUnreachable(ws workspaceEnv, destroy func() error) error {
	ws.SetEnvVar(deleteUnreachableEnv, "true")
	defer ws.UnsetEnvVar(deleteUnreachableEnv)
	return destroy()
}
