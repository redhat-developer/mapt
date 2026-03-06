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
	rhoaiNamespace = "redhat-ods-operator"
)

var (
	dscGVR = schema.GroupVersionResource{
		Group:    "datasciencecluster.opendatahub.io",
		Version:  "v1",
		Resource: "datascienceclusters",
	}
)

// deployOpenShiftAI installs the RHOAI operator and creates a DataScienceCluster.
// The entire RHOAI installation is gated on prereqs (ServiceMesh v2, Authorino,
// and Serverless readiness outputs) so that when the operator starts and auto-creates
// the DSCI, it finds all dependencies already in place.
func deployOpenShiftAI(ctx *pulumi.Context, args *DeployArgs, prereqs []pulumi.StringOutput) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-rhoai-%s", args.Prefix, suffix)
	}

	// Gate the entire RHOAI installation on prerequisites.
	// The namespace name won't resolve until all prereqs are ready,
	// which delays the operator install until SM v2 + Authorino +
	// Serverless are fully operational.
	nsName := pulumi.String(rhoaiNamespace).ToStringOutput()
	for _, p := range prereqs {
		prev := nsName
		nsName = pulumi.All(prev, p).ApplyT(
			func(args []interface{}) string {
				return args[0].(string)
			}).(pulumi.StringOutput)
	}

	// Create Namespace (blocked until all prereqs resolve)
	ns, err := corev1.NewNamespace(ctx, rn("ns"),
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

	// Create OperatorGroup (AllNamespaces mode — no targetNamespaces)
	og, err := apiextensions.NewCustomResource(ctx, rn("og"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("operators.coreos.com/v1"),
			Kind:       pulumi.String("OperatorGroup"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("redhat-ods-operator-group"),
				Namespace: pulumi.String(rhoaiNamespace),
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
				Name:      pulumi.String("rhods-operator"),
				Namespace: pulumi.String(rhoaiNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":              "redhat-operators",
					"sourceNamespace":     "openshift-marketplace",
					"name":               "rhods-operator",
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

	// Wait for CSV to succeed (operator fully installed).
	csvReady := pulumi.All(sub.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, csvGVR,
				rhoaiNamespace, "rhods-operator",
				"", "Succeeded", 20*time.Minute, true); err != nil {
				return "", fmt.Errorf("waiting for RHOAI CSV: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	// Create DataScienceCluster CR after RHOAI CSV is ready.
	dscName := csvReady.ApplyT(func(_ string) string {
		return "default-dsc"
	}).(pulumi.StringOutput)

	dsc, err := apiextensions.NewCustomResource(ctx, rn("dsc"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("datasciencecluster.opendatahub.io/v1"),
			Kind:       pulumi.String("DataScienceCluster"),
			Metadata: &metav1.ObjectMetaArgs{
				Name: dscName,
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"components": map[string]interface{}{
						"dashboard":            map[string]interface{}{"managementState": "Managed"},
						"workbenches":          map[string]interface{}{"managementState": "Managed"},
						"datasciencepipelines": map[string]interface{}{"managementState": "Managed"},
						// Kserve depends on ServiceMesh and Serverless which are
						// deployed as implicit dependencies of the AI profile.
						"kserve":           map[string]interface{}{"managementState": "Managed"},
						"modelmeshserving": map[string]interface{}{"managementState": "Managed"},
						"ray":              map[string]interface{}{"managementState": "Managed"},
						// Kueue webhook fails on SNC due to missing endpoints
						"kueue":            map[string]interface{}{"managementState": "Removed"},
						"trustyai":         map[string]interface{}{"managementState": "Managed"},
						"codeflare":        map[string]interface{}{"managementState": "Managed"},
						"trainingoperator": map[string]interface{}{"managementState": "Removed"},
						"modelregistry":    map[string]interface{}{"managementState": "Removed"},
					},
				},
			},
		},
		pulumi.Provider(args.K8sProvider))
	if err != nil {
		return nil, err
	}

	// Wait for DataScienceCluster to be ready.
	dscReady := pulumi.All(dsc.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, dscGVR,
				"", "default-dsc",
				"Ready", "True", 40*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for DataScienceCluster: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("dscReady", dscReady)

	return dsc, nil
}
