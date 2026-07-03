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
	cnvNamespace = "openshift-cnv"
)

var (
	hcoGVR = schema.GroupVersionResource{
		Group:    "hco.kubevirt.io",
		Version:  "v1beta1",
		Resource: "hyperconvergeds",
	}
)

func deployVirtualization(ctx *pulumi.Context, args *DeployArgs) (pulumi.Resource, error) {
	goCtx := ctx.Context()
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-virt-%s", args.Prefix, suffix)
	}

	// Install the CNV operator
	csvReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      cnvNamespace,
		ogName:         "kubevirt-hyperconverged-group",
		ogTargetNS:     []string{cnvNamespace},
		subName:        "hco-operatorhub",
		packageName:    "kubevirt-hyperconverged",
		csvPrefix:      "kubevirt-hyperconverged-operator",
	})
	if err != nil {
		return nil, err
	}

	// Create HyperConverged CR — the Name depends on the CSV wait completing.
	hcoName := csvReady.ApplyT(func(_ string) string {
		return "kubevirt-hyperconverged"
	}).(pulumi.StringOutput)

	hco, err := apiextensions.NewCustomResource(ctx, rn("hco"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("hco.kubevirt.io/v1beta1"),
			Kind:       pulumi.String("HyperConverged"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      hcoName,
				Namespace: pulumi.String(cnvNamespace),
			},
		},
		args.k8sOpts()...)
	if err != nil {
		return nil, err
	}

	// Wait for HyperConverged to be ready
	hcoReady := pulumi.All(hco.ID(), args.Kubeconfig).ApplyT(
		func(allArgs []interface{}) (string, error) {
			kc := allArgs[1].(string)
			if err := waitForCRCondition(goCtx, kc, hcoGVR,
				cnvNamespace, "kubevirt-hyperconverged",
				"", "Available", "True", 20*time.Minute, false); err != nil {
				return "", fmt.Errorf("waiting for HyperConverged: %w", err)
			}
			return "ready", nil
		}).(pulumi.StringOutput)

	ctx.Export("hcoReady", hcoReady)

	return hco, nil
}
