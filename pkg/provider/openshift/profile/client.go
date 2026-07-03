package profile

import (
	"context"
	"fmt"
	"strings"
	"time"

	logging "github.com/redhat-developer/mapt/pkg/util/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// csvGVR is used by multiple profiles to wait for OLM ClusterServiceVersions.
var csvGVR = schema.GroupVersionResource{
	Group:    "operators.coreos.com",
	Version:  "v1alpha1",
	Resource: "clusterserviceversions",
}

// waitForCRCondition polls a custom resource until a status field matches the expected value.
//
// It supports two modes controlled by statusField and condType:
//
//   - Scalar field: set statusField to a top-level key under .status (e.g. "phase", "state").
//     The function checks status[statusField] == expected.
//   - Conditions array: leave statusField empty and set condType to the condition type
//     (e.g. "Available"). The function searches .status.conditions for a matching
//     type and checks its status field.
//
// When prefixMatch is true, the name is treated as a prefix and the first
// resource whose name starts with that prefix is used (useful for CSVs
// whose names include a version suffix, e.g. "kubevirt-hyperconverged-operator.v4.21.0").
//
// The function blocks until the condition is met or the timeout expires.
func waitForCRCondition(ctx context.Context, kubeconfig string, gvr schema.GroupVersionResource,
	namespace, name, statusField, condType, expected string, timeout time.Duration, prefixMatch bool) error {

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("building REST config: %w", err)
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("creating dynamic client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if time.Now().After(deadline) {
			field := condType
			if statusField != "" {
				field = statusField
			}
			return fmt.Errorf("timed out waiting for %s/%s %q to be %q", namespace, name, field, expected)
		}

		obj, err := findResource(ctx, dc, gvr, namespace, name, prefixMatch)
		if err != nil {
			logging.Debugf("waiting for %s/%s: %v", namespace, name, err)
		} else if obj != nil && matchesCondition(obj, statusField, condType, expected) {
			return nil
		}
		time.Sleep(15 * time.Second)
	}
}

// findResource returns a single resource by exact name or by name prefix.
// When namespace is empty, the resource is looked up at cluster scope.
func findResource(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource,
	namespace, name string, prefixMatch bool) (*unstructured.Unstructured, error) {

	var ri dynamic.ResourceInterface
	if namespace == "" {
		ri = dc.Resource(gvr)
	} else {
		ri = dc.Resource(gvr).Namespace(namespace)
	}

	if !prefixMatch {
		return ri.Get(ctx, name, metav1.GetOptions{})
	}

	list, err := ri.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range list.Items {
		if strings.HasPrefix(list.Items[i].GetName(), name) {
			return &list.Items[i], nil
		}
	}
	return nil, fmt.Errorf("no resource with prefix %q found in %s", name, namespace)
}

// matchesCondition checks whether the unstructured object satisfies the expected status.
//
// When statusField is non-empty it checks .status[statusField] == expected (scalar check).
// When statusField is empty it searches .status.conditions for a matching condType
// and checks its status field.
func matchesCondition(obj *unstructured.Unstructured, statusField, condType, expected string) bool {
	status, ok := obj.Object["status"].(map[string]interface{})
	if !ok {
		return false
	}
	if statusField != "" {
		// Scalar field check (e.g. status.phase, status.state)
		val, _ := status[statusField].(string)
		return val == expected
	}
	// Condition array check (e.g. HCO Available)
	conditions, ok := status["conditions"].([]interface{})
	if !ok {
		return false
	}
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if cond["type"] == condType && cond["status"] == expected {
			return true
		}
	}
	return false
}
