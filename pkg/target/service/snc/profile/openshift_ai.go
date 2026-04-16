package profile

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
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

	// Install the RHOAI operator (gated on prereqs via nsName)
	csvReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      rhoaiNamespace,
		nsName:         nsName,
		ogName:         "redhat-ods-operator-group",
		subName:        "rhods-operator",
		packageName:    "rhods-operator",
		csvPrefix:      "rhods-operator",
	})
	if err != nil {
		return nil, err
	}

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
						"kserve":               map[string]interface{}{"managementState": "Managed"},
						"modelmeshserving":     map[string]interface{}{"managementState": "Removed"},
						"ray":                  map[string]interface{}{"managementState": "Managed"},
						"kueue":                map[string]interface{}{"managementState": "Removed"},
						"trustyai":             map[string]interface{}{"managementState": "Managed"},
						"codeflare":            map[string]interface{}{"managementState": "Managed"},
						"trainingoperator":     map[string]interface{}{"managementState": "Removed"},
						"modelregistry":        map[string]interface{}{"managementState": "Removed"},
					},
				},
			},
		},
		args.k8sOpts()...)
	if err != nil {
		return nil, err
	}

	// Wait for DataScienceCluster to be ready.
	dscReady := pulumi.All(dsc.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, dscGVR,
				"", "default-dsc",
				"", "Ready", "True", 40*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for DataScienceCluster: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("dscReady", dscReady)

	return dsc, nil
}
