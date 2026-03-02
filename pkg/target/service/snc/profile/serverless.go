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

// deployServerless installs the Serverless operator and deploys the requested
// Knative components (Serving, Eventing, or both).
func deployServerless(ctx *pulumi.Context, args *DeployArgs, serving, eventing bool) error {
	operatorReady, err := deployServerlessOperator(ctx, args)
	if err != nil {
		return err
	}
	if serving {
		if _, err := deployKnativeServing(ctx, args, operatorReady); err != nil {
			return err
		}
	}
	if eventing {
		if _, err := deployKnativeEventing(ctx, args, operatorReady); err != nil {
			return err
		}
	}
	return nil
}

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

// knativeCRArgs describes a Knative custom resource to deploy.
type knativeCRArgs struct {
	suffix    string // short identifier used in Pulumi resource names (e.g. "ks", "ke")
	namespace string
	kind      string // e.g. "KnativeServing"
	crName    string // e.g. "knative-serving"
	gvr       schema.GroupVersionResource
	exportKey string // Pulumi export name for readiness
}

// deployKnativeCR is the shared implementation for deploying a Knative CR
// (Serving or Eventing). It creates the target namespace, the CR, and waits
// for it to become ready.
func deployKnativeCR(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput, cr knativeCRArgs) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(s string) string {
		return fmt.Sprintf("%s-serverless-%s", args.Prefix, s)
	}

	// Thread the operator-ready wait into the namespace name via ApplyT
	nsName := operatorReady.ApplyT(func(_ string) string {
		return cr.namespace
	}).(pulumi.StringOutput)

	// Create target namespace
	ns, err := corev1.NewNamespace(ctx, rn(cr.suffix+"-ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: nsName,
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, err
	}

	// Create the Knative CR
	res, err := apiextensions.NewCustomResource(ctx, rn(cr.suffix),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operator.knative.dev/v1beta1"),
			Kind:       pulumi.String(cr.kind),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(cr.crName),
				Namespace: pulumi.String(cr.namespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	// Wait for the CR to be ready
	ready := pulumi.All(res.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, cr.gvr,
				cr.namespace, cr.crName,
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for %s: %w", cr.kind, err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export(cr.exportKey, ready)

	return res, nil
}

// deployKnativeServing creates a KnativeServing CR and waits for it to be ready.
func deployKnativeServing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, error) {
	return deployKnativeCR(ctx, args, operatorReady, knativeCRArgs{
		suffix:    "ks",
		namespace: knativeServingNamespace,
		kind:      "KnativeServing",
		crName:    "knative-serving",
		gvr:       knativeServingGVR,
		exportKey: "knativeServingReady",
	})
}

// deployKnativeEventing creates a KnativeEventing CR and waits for it to be ready.
func deployKnativeEventing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, error) {
	return deployKnativeCR(ctx, args, operatorReady, knativeCRArgs{
		suffix:    "ke",
		namespace: knativeEventingNamespace,
		kind:      "KnativeEventing",
		crName:    "knative-eventing",
		gvr:       knativeEventingGVR,
		exportKey: "knativeEventingReady",
	})
}
