package profile

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	serverlessNamespace      = "openshift-serverless"
	knativeServingNamespace  = "knative-serving"
	knativeEventingNamespace = "knative-eventing"
)

var (
	knativeServingGVR = schema.GroupVersionResource{
		Group:    "operator.knative.dev",
		Version:  "v1beta1",
		Resource: "knativeservings",
	}

	knativeEventingGVR = schema.GroupVersionResource{
		Group:    "operator.knative.dev",
		Version:  "v1beta1",
		Resource: "knativeeventings",
	}
)

// deployServerlessOperator installs the OpenShift Serverless operator and waits
// for the CSV to succeed. It returns a pulumi.StringOutput that resolves after
// the operator is ready, suitable for threading namespace names through ApplyT.
func deployServerlessOperator(ctx *pulumi.Context, args *DeployArgs) (pulumi.StringOutput, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-serverless-%s", args.Prefix, suffix)
	}

	// Create openshift-serverless namespace
	ns, err := corev1.NewNamespace(ctx, rn("ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(serverlessNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Create OperatorGroup (AllNamespaces — empty spec)
	og, err := apiextensions.NewCustomResource(ctx, rn("og"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1"),
			Kind:       pulumi.String("OperatorGroup"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("serverless-operators"),
				Namespace: pulumi.String(serverlessNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Create Subscription
	sub, err := apiextensions.NewCustomResource(ctx, rn("sub"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("serverless-operator"),
				Namespace: pulumi.String(serverlessNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":                "serverless-operator",
					"channel":             "stable",
					"installPlanApproval": "Automatic",
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{og}))
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Wait for CSV to succeed (operator fully installed).
	operatorReady := pulumi.All(sub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				serverlessNamespace, "serverless-operator",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for Serverless CSV: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	return operatorReady, nil
}

// deployKnativeServing creates a KnativeServing CR and waits for it to be ready.
// The operatorReady output is used to chain the dependency on the operator installation.
func deployKnativeServing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-serverless-%s", args.Prefix, suffix)
	}

	// Thread the wait into the namespace name via ApplyT
	ksNSName := operatorReady.ApplyT(func(_ string) string {
		return knativeServingNamespace
	}).(pulumi.StringOutput)

	// Create knative-serving namespace
	ksNS, err := corev1.NewNamespace(ctx, rn("ks-ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: ksNSName,
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Create KnativeServing CR
	ks, err := apiextensions.NewCustomResource(ctx, rn("ks"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operator.knative.dev/v1beta1"),
			Kind:       pulumi.String("KnativeServing"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("knative-serving"),
				Namespace: pulumi.String(knativeServingNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ksNS}))
	if err != nil {
		return nil, err
	}

	// Wait for KnativeServing to be ready.
	ksReady := pulumi.All(ks.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, knativeServingGVR,
				knativeServingNamespace, "knative-serving",
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for KnativeServing: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("knativeServingReady", ksReady)

	return ks, nil
}

// deployKnativeEventing creates a KnativeEventing CR and waits for it to be ready.
// The operatorReady output is used to chain the dependency on the operator installation.
func deployKnativeEventing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-serverless-%s", args.Prefix, suffix)
	}

	// Thread the wait into the namespace name via ApplyT
	keNSName := operatorReady.ApplyT(func(_ string) string {
		return knativeEventingNamespace
	}).(pulumi.StringOutput)

	// Create knative-eventing namespace
	keNS, err := corev1.NewNamespace(ctx, rn("ke-ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: keNSName,
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Create KnativeEventing CR
	ke, err := apiextensions.NewCustomResource(ctx, rn("ke"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operator.knative.dev/v1beta1"),
			Kind:       pulumi.String("KnativeEventing"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("knative-eventing"),
				Namespace: pulumi.String(knativeEventingNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{keNS}))
	if err != nil {
		return nil, err
	}

	// Wait for KnativeEventing to be ready.
	keReady := pulumi.All(ke.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, knativeEventingGVR,
				knativeEventingNamespace, "knative-eventing",
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for KnativeEventing: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("knativeEventingReady", keReady)

	return ke, nil
}
