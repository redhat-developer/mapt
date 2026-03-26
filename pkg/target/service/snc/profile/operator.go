package profile

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// operatorInstall describes how to install an OLM operator on the cluster.
type operatorInstall struct {
	// resourcePrefix is used for Pulumi resource naming (e.g. "main-virt-").
	resourcePrefix string
	// namespace where the Subscription (and optional OG) are created.
	namespace string
	// nsName is a dynamic namespace name for prerequisite gating via ApplyT.
	// When nil, pulumi.String(namespace) is used.
	nsName pulumi.StringInput
	// ogName is the OperatorGroup name. When non-empty, a dedicated Namespace
	// and OperatorGroup are created for this operator.
	ogName string
	// ogTargetNS are the OG targetNamespaces. nil means AllNamespaces mode.
	ogTargetNS []string
	// subName is the Subscription metadata.name.
	subName string
	// packageName is the operator package name in the catalog.
	packageName string
	// channel is the subscription channel (defaults to "stable").
	channel string
	// csvPrefix is the CSV name prefix used for prefix-matching during wait.
	csvPrefix string
	// csvNamespace overrides where to look for the CSV; empty uses namespace.
	csvNamespace string
	// extraDeps are additional Pulumi resources the Subscription depends on.
	extraDeps []pulumi.Resource
}

// installOperator creates the namespace (optional), OperatorGroup (optional),
// Subscription, and waits for the CSV to succeed.
// Returns a StringOutput that resolves when the operator is fully installed.
func installOperator(ctx *pulumi.Context, args *DeployArgs, oi operatorInstall) (pulumi.StringOutput, error) {
	goCtx := ctx.Context()
	channel := oi.channel
	if channel == "" {
		channel = "stable"
	}

	deps := append([]pulumi.Resource{}, args.Deps...)
	deps = append(deps, oi.extraDeps...)

	// If ogName is provided, create a dedicated namespace and OperatorGroup.
	if oi.ogName != "" {
		nsName := oi.nsName
		if nsName == nil {
			nsName = pulumi.String(oi.namespace)
		}
		ns, err := corev1.NewNamespace(ctx, oi.resourcePrefix+"ns",
			&corev1.NamespaceArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name: nsName,
				},
			},
			args.k8sOpts(pulumi.DependsOn(deps))...)
		if err != nil {
			return pulumi.StringOutput{}, err
		}

		ogSpec := map[string]interface{}{}
		if oi.ogTargetNS != nil {
			ogSpec["targetNamespaces"] = oi.ogTargetNS
		}
		og, err := apiextensions.NewCustomResource(ctx, oi.resourcePrefix+"og",
			&apiextensions.CustomResourceArgs{
				ApiVersion: pulumi.String("operators.coreos.com/v1"),
				Kind:       pulumi.String("OperatorGroup"),
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(oi.ogName),
					Namespace: pulumi.String(oi.namespace),
				},
				OtherFields: map[string]interface{}{
					"spec": ogSpec,
				},
			},
			args.k8sOpts(pulumi.DependsOn([]pulumi.Resource{ns}))...)
		if err != nil {
			return pulumi.StringOutput{}, err
		}
		deps = []pulumi.Resource{og}
	}

	// Create Subscription
	sub, err := apiextensions.NewCustomResource(ctx, oi.resourcePrefix+"sub",
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(oi.subName),
				Namespace: pulumi.String(oi.namespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               oi.packageName,
					"channel":            channel,
					"installPlanApproval": "Automatic",
				},
			},
		},
		args.k8sOpts(pulumi.DependsOn(deps))...)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Wait for CSV to succeed
	csvNS := oi.csvNamespace
	if csvNS == "" {
		csvNS = oi.namespace
	}
	ready := pulumi.All(sub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				csvNS, oi.csvPrefix,
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for %s CSV: %w", oi.csvPrefix, err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	return ready, nil
}
