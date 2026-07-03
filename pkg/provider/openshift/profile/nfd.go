package profile

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	nfdNamespace   = "openshift-nfd"
	nfdOGName      = "openshift-nfd"
	nfdPackageName = "nfd"
	nfdCRName      = "nfd-instance"
	nfdAPIVersion  = "nfd.openshift.io/v1"
	nfdKind        = "NodeFeatureDiscovery"
)

// deployNFD installs the Node Feature Discovery operator and creates a
// NodeFeatureDiscovery CR. NFD is a prerequisite for the NVIDIA GPU Operator
// as it labels nodes with hardware features (e.g. PCI vendor 10de for NVIDIA GPUs).
// Returns a StringOutput that resolves when the NFD operator CSV is ready.
func deployNFD(ctx *pulumi.Context, args *DeployArgs) (pulumi.StringOutput, error) {
	rn := func(suffix string) string {
		return fmt.Sprintf("%s-nfd-%s", args.Prefix, suffix)
	}

	// Install the NFD operator
	csvReady, err := installOperator(ctx, args, operatorInstall{
		resourcePrefix: rn(""),
		namespace:      nfdNamespace,
		ogName:         nfdOGName,
		ogTargetNS:     []string{nfdNamespace},
		subName:        nfdPackageName,
		packageName:    nfdPackageName,
		csvPrefix:      nfdPackageName,
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Create NodeFeatureDiscovery CR after CSV is ready
	nfdName := csvReady.ApplyT(func(_ string) string {
		return nfdCRName
	}).(pulumi.StringOutput)

	if _, err := apiextensions.NewCustomResource(ctx, rn("cr"),
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(nfdAPIVersion),
			Kind:       pulumi.String(nfdKind),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      nfdName,
				Namespace: pulumi.String(nfdNamespace),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"operand": map[string]interface{}{
						"imagePullPolicy": "Always",
					},
					"workerConfig": map[string]interface{}{
						"configData": "core:\n  sleepInterval: 60s\n",
					},
				},
			},
		},
		args.k8sOpts()...); err != nil {
		return pulumi.StringOutput{}, err
	}

	return csvReady, nil
}
