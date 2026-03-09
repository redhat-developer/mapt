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
	istioSystemNamespace = "istio-system"
	istioCNINamespace    = "istio-cni"
)

var (
	sailCSVGVR = schema.GroupVersionResource{
		Group:    "operators.coreos.com",
		Version:  "v1alpha1",
		Resource: "clusterserviceversions",
	}
	istioGVR = schema.GroupVersionResource{
		Group:    "sailoperator.io",
		Version:  "v1",
		Resource: "istios",
	}
	istioCNIGVR = schema.GroupVersionResource{
		Group:    "sailoperator.io",
		Version:  "v1",
		Resource: "istiocnis",
	}
)

func deployServiceMesh(ctx *pulumi.Context, args *ProfileDeployArgs) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-smesh-%s", args.Prefix, suffix)
	}

	// Create istio-system namespace
	nsSystem, err := corev1.NewNamespace(ctx, rn("ns-system"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(istioSystemNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, err
	}

	// Create istio-cni namespace
	nsCNI, err := corev1.NewNamespace(ctx, rn("ns-cni"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(istioCNINamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, err
	}

	// Create Subscription for the OpenShift Service Mesh 3 operator
	sub, err := apiextensions.NewCustomResource(ctx, rn("sub"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("servicemeshoperator3"),
				Namespace: pulumi.String("openshift-operators"),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               "servicemeshoperator3",
					"channel":            "stable",
					"installPlanApproval": "Automatic",
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{nsSystem, nsCNI}))
	if err != nil {
		return nil, err
	}

	// Wait for the Service Mesh operator CSV to succeed
	csvReady := pulumi.All(sub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, sailCSVGVR,
				"openshift-operators", "servicemeshoperator3",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for Service Mesh operator CSV: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Create IstioCNI CR
	istioCNIName := csvReady.ApplyT(func(_ string) string {
		return "default"
	}).(pulumi.StringOutput)

	// IstioCNI is cluster-scoped
	cni, err := apiextensions.NewCustomResource(ctx, rn("istiocni"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("sailoperator.io/v1"),
			Kind:       pulumi.String("IstioCNI"),
			Metadata: &metav1.ObjectMetaArgs{
				Name: istioCNIName,
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"namespace": istioCNINamespace,
					"profile":   "openshift",
				},
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Wait for IstioCNI to be ready (cluster-scoped, empty namespace)
	cniReady := pulumi.All(cni.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, istioCNIGVR,
				"", "default",
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for IstioCNI: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Create Istio CR (cluster-scoped, depends on CNI being ready)
	istioName := cniReady.ApplyT(func(_ string) string {
		return "default"
	}).(pulumi.StringOutput)

	istio, err := apiextensions.NewCustomResource(ctx, rn("istio"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("sailoperator.io/v1"),
			Kind:       pulumi.String("Istio"),
			Metadata: &metav1.ObjectMetaArgs{
				Name: istioName,
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"namespace": istioSystemNamespace,
				},
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Wait for Istio to be ready (cluster-scoped, empty namespace)
	istioReady := pulumi.All(istio.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, istioGVR,
				"", "default",
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for Istio: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("istioReady", istioReady)

	return istio, nil
}
