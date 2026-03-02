package snc

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
	cnvNamespace = "openshift-cnv"
)

var (
	csvGVR = schema.GroupVersionResource{
		Group:    "operators.coreos.com",
		Version:  "v1alpha1",
		Resource: "clusterserviceversions",
	}
	hcoGVR = schema.GroupVersionResource{
		Group:    "hco.kubevirt.io",
		Version:  "v1beta1",
		Resource: "hyperconvergeds",
	}
)

func deployVirtualization(ctx *pulumi.Context, args *ProfileDeployArgs) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-virt-%s", args.Prefix, suffix)
	}

	// Create Namespace
	ns, err := corev1.NewNamespace(ctx, rn("ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(cnvNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, err
	}

	// Create OperatorGroup
	og, err := apiextensions.NewCustomResource(ctx, rn("og"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1"),
			Kind:       pulumi.String("OperatorGroup"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("kubevirt-hyperconverged-group"),
				Namespace: pulumi.String(cnvNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"targetNamespaces": []string{cnvNamespace},
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	// Create Subscription
	sub, err := apiextensions.NewCustomResource(ctx, rn("sub"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("hco-operatorhub"),
				Namespace: pulumi.String(cnvNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               "kubevirt-hyperconverged",
					"channel":            "stable",
					"installPlanApproval": "Automatic",
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{og}))
	if err != nil {
		return nil, err
	}

	// Wait for CSV to succeed (operator fully installed) using client-go.
	// We thread the wait into the HCO resource name via ApplyT so Pulumi
	// knows the dependency ordering.
	hcoName := pulumi.All(sub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				cnvNamespace, "kubevirt-hyperconverged-operator",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for CNV CSV: %w", err)
			}
			return "kubevirt-hyperconverged", nil
		}).(pulumi.StringOutput)

	// Create HyperConverged CR — the Name depends on the CSV wait completing.
	hco, err := apiextensions.NewCustomResource(ctx, rn("hco"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("hco.kubevirt.io/v1beta1"),
			Kind:       pulumi.String("HyperConverged"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      hcoName,
				Namespace: pulumi.String(cnvNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Wait for HyperConverged to be ready using client-go.
	// Capture the output so Pulumi tracks the wait and propagates errors.
	hcoReady := pulumi.All(hco.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, hcoGVR,
				cnvNamespace, "kubevirt-hyperconverged",
				"Available", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for HyperConverged: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Export the readiness output so Pulumi waits for it to resolve.
	ctx.Export("hcoReady", hcoReady)

	return hco, nil
}
