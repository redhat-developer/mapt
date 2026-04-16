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
		args.k8sOpts(pulumi.DependsOn(args.Deps))...)
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Install Service Mesh v2 operator (into openshift-operators)
	smCSVReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      "openshift-operators",
		subName:        "servicemeshoperator",
		packageName:    "servicemeshoperator",
		csvPrefix:      "servicemeshoperator",
		extraDeps:      []pulumi.Resource{ns},
	})
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

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
		args.k8sOpts()...)
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Wait for SMCP to be ready
	smcpReady := pulumi.All(smcp.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, smcpGVR,
				istioSystemNamespace, "data-science-smcp",
				"", "Ready", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for SMCP: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Create ServiceMeshMemberRoll to enroll namespaces used by RHOAI model serving.
	// SM v2 requires explicit namespace enrollment unlike SM v3 which is cluster-wide.
	smmrName := smcpReady.ApplyT(func(_ string) string {
		return "default"
	}).(pulumi.StringOutput)

	if _, err := apiextensions.NewCustomResource(ctx, rn("smmr"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("maistra.io/v1"),
			Kind:       pulumi.String("ServiceMeshMemberRoll"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      smmrName,
				Namespace: pulumi.String(istioSystemNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"members": []string{
						"knative-serving",
						"redhat-ods-applications",
					},
				},
			},
		},
		args.k8sOpts()...); err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Install Authorino operator (into openshift-operators)
	authReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn("authorino-"),
		namespace:      "openshift-operators",
		subName:        "authorino-operator",
		packageName:    "authorino-operator",
		csvPrefix:      "authorino-operator",
	})
	if err != nil {
		return nil, pulumi.StringOutput{}, err
	}

	// Combine SMCP + Authorino readiness into a single output
	allReady := pulumi.All(smcpReady, authReady).ApplyT(
		func(_ []interface{}) string {
			return "ready"
		}).(pulumi.StringOutput)

	ctx.Export("smcpReady", allReady)

	return smcp, allReady, nil
}
