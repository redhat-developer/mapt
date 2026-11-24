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
	destroyTargets := func(options manager.DestroyOptions) error {
		return manager.DestroySelectedStack(mCtx, stack, options)
	}
	phases := teardownPhases{
		cleanup: func(ctx context.Context) error {
			if identity == nil {
				return nil
			}
			cleaner, err := newAWSCleaner(ctx, *identity)
			if err != nil {
				return withPermissionsHint(err)
			}
			return withPermissionsHint(cleaner.cleanup(ctx))
		},
		final: func() error {
			return destroyTargets(manager.DestroyOptions{})
		},
	}
	if len(kubernetesTargets) > 0 {
		phases.kubernetes = func() error {
			logging.Debug("EKS teardown: removing Kubernetes resources and providers before the cluster")
			return withDeleteUnreachable(stack.Workspace(), func() error {
				return destroyTargets(manager.DestroyOptions{
					Targets: kubernetesTargets, ContinueOnError: true,
				})
			})
		}
	}
	if clusterURN != "" {
		phases.cluster = func() error {
			return destroyTargets(manager.DestroyOptions{
				Targets: []string{clusterURN}, TargetDependents: true, ContinueOnError: true,
			})
		}
	}
	return runEKSTeardown(mCtx.Context(), phases)
}

// teardownPhases holds the destroy steps: Kubernetes resources, then the
// cluster and its dependents, then the external resources the cluster
// controllers created, then everything that is left. A nil step is skipped.
type teardownPhases struct {
	kubernetes func() error // nil when the checkpoint has no Kubernetes resources
	cluster    func() error // nil when the cluster is no longer in state
	cleanup    func(context.Context) error
	final      func() error
}

// runEKSTeardown runs the phases in order and stops at the first error.
// ContinueOnError still reports an error when any deletion failed, so a phase
// starts only after the previous ones removed every resource they own. The
// final phase therefore never sees Kubernetes resources and does not need
// PULUMI_K8S_DELETE_UNREACHABLE.
func runEKSTeardown(ctx context.Context, phases teardownPhases) error {
	for _, step := range []struct {
		name string
		run  func() error
	}{
		{"destroying kubernetes resources", phases.kubernetes},
		{"destroying EKS cluster and dependents", phases.cluster},
		{"cleaning up controller resources", func() error {
			if phases.cleanup == nil {
				return nil
			}
			return phases.cleanup(ctx)
		}},
		{"destroying network and remaining resources", phases.final},
	} {
		if step.run == nil {
			continue
		}
		if err := step.run(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return nil
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
