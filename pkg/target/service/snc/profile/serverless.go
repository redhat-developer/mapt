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

// deployServerlessWithPrereqs installs the Serverless operator and deploys the
// requested Knative components. When needAI is true, the serving readiness
// output is appended to aiPrereqs for the AI profile to chain on.
func deployServerlessWithPrereqs(ctx *pulumi.Context, args *DeployArgs, serving, eventing, needAI bool, aiPrereqs *[]pulumi.StringOutput) error {
	operatorReady, err := deployServerlessOperator(ctx, args)
	if err != nil {
		return err
	}
	if serving {
		_, ksReady, err := deployKnativeServing(ctx, args, operatorReady)
		if err != nil {
			return err
		}
		if needAI {
			*aiPrereqs = append(*aiPrereqs, ksReady)
		}
	}
	if eventing {
		if _, _, err := deployKnativeEventing(ctx, args, operatorReady); err != nil {
			return err
		}
	}
	return nil
}

// deployServerlessOperator installs the OpenShift Serverless operator and waits
// for the CSV to succeed.
func deployServerlessOperator(ctx *pulumi.Context, args *DeployArgs) (pulumi.StringOutput, error) {
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-serverless-%s", args.Prefix, suffix)
	}
	return installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      serverlessNamespace,
		ogName:         "serverless-operators",
		subName:        "serverless-operator",
		packageName:    "serverless-operator",
		csvPrefix:      "serverless-operator",
	})
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
// for it to become ready. Returns the resource and a readiness output.
func deployKnativeCR(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput, cr knativeCRArgs) (pulumi.Resource, pulumi.StringOutput, error) {
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
		args.k8sOpts(pulumi.DependsOn(args.Deps))...)
	if err != nil {
		return nil, pulumi.StringOutput{}, err
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
		args.k8sOpts(pulumi.DependsOn([]pulumi.Resource{ns}))...)
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Wait for the CR to be ready
	ready := pulumi.All(res.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, cr.gvr,
				cr.namespace, cr.crName,
				"", "Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for %s: %w", cr.kind, err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export(cr.exportKey, ready)

	return res, ready, nil
}

// deployKnativeServing creates a KnativeServing CR and waits for it to be ready.
func deployKnativeServing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, pulumi.StringOutput, error) {
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
func deployKnativeEventing(ctx *pulumi.Context, args *DeployArgs, operatorReady pulumi.StringOutput) (pulumi.Resource, pulumi.StringOutput, error) {
	return deployKnativeCR(ctx, args, operatorReady, knativeCRArgs{
		suffix:    "ke",
		namespace: knativeEventingNamespace,
		kind:      "KnativeEventing",
		crName:    "knative-eventing",
		gvr:       knativeEventingGVR,
		exportKey: "knativeEventingReady",
	})
}
