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

var (
	smcpGVR = schema.GroupVersionResource{
		Group:    "maistra.io",
		Version:  "v2",
		Resource: "servicemeshcontrolplanes",
	}
)

// deployServiceMeshV2 installs OpenShift Service Mesh v2 (Maistra/Istio) and the
// Authorino operator, both required by RHOAI for Kserve. It creates an SMCP named
// "data-science-smcp" in istio-system, matching the DSCI defaults.
// Returns a StringOutput that resolves when both SMCP and Authorino are ready.
func deployServiceMeshV2(ctx *pulumi.Context, args *DeployArgs) (pulumi.Resource, pulumi.StringOutput, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-smeshv2-%s", args.Prefix, suffix)
	}

	// Create istio-system namespace
	ns, err := corev1.NewNamespace(ctx, rn("ns"),
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(istioSystemNamespace),
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// --- Service Mesh v2 operator ---

	// Create Subscription (openshift-operators is a pre-existing global namespace
	// with an OperatorGroup, no need to create one).
	smSub, err := apiextensions.NewCustomResource(ctx, rn("sub"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("servicemeshoperator"),
				Namespace: pulumi.String("openshift-operators"),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               "servicemeshoperator",
					"channel":            "stable",
					"installPlanApproval": "Automatic",
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Wait for the Service Mesh v2 CSV to succeed
	smCSVReady := pulumi.All(smSub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				"openshift-operators", "servicemeshoperator",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for Service Mesh v2 CSV: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Create ServiceMeshControlPlane — "data-science-smcp" matches the DSCI default
	smcpName := smCSVReady.ApplyT(func(_ string) string {
		return "data-science-smcp"
	}).(pulumi.StringOutput)

	smcp, err := apiextensions.NewCustomResource(ctx, rn("smcp"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("maistra.io/v2"),
			Kind:       pulumi.String("ServiceMeshControlPlane"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      smcpName,
				Namespace: pulumi.String(istioSystemNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"version": "v2.6",
					"tracing": map[string]interface{}{
						"type": "None",
					},
					"security": map[string]interface{}{
						"dataPlane": map[string]interface{}{
							"mtls": true,
						},
					},
					"addons": map[string]interface{}{
						"kiali": map[string]interface{}{
							"enabled": false,
						},
						"grafana": map[string]interface{}{
							"enabled": false,
						},
						"prometheus": map[string]interface{}{
							"enabled": false,
						},
					},
				},
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Wait for SMCP to be ready
	smcpReady := pulumi.All(smcp.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, smcpGVR,
				istioSystemNamespace, "data-science-smcp",
				"Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for SMCP: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// --- Authorino operator (required by RHOAI for ServiceMesh authorization) ---

	authSub, err := apiextensions.NewCustomResource(ctx, rn("authorino-sub"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1alpha1"),
			Kind:       pulumi.String("Subscription"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("authorino-operator"),
				Namespace: pulumi.String("openshift-operators"),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               "authorino-operator",
					"channel":            "stable",
					"installPlanApproval": "Automatic",
				},
			},
		},
		pulumi.Provider(args.K8sProvider),
		pulumi.DependsOn(args.Deps))
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Wait for Authorino CSV to succeed
	authReady := pulumi.All(authSub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				"openshift-operators", "authorino-operator",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for Authorino CSV: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Combine SMCP + Authorino readiness into a single output
	allReady := pulumi.All(smcpReady, authReady).ApplyT(
		func(_ []interface{}) string {
			return "ready"
		}).(pulumi.StringOutput)

	ctx.Export("smcpReady", allReady)

	return smcp, allReady, nil
}
