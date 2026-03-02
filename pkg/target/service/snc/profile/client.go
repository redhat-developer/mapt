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

// waitForCRCondition polls a custom resource until a nested field matches the expected value.
// jsonPath is a dot-separated path into the object (e.g. "status.phase" or
// "status.conditions[?(@.type==\"Available\")].status") but here we use explicit
// field traversal for the two cases we need:
//   - CSV:  status.phase == "Succeeded"  (single scalar)
//   - HCO:  status.conditions where type==Available → status == "True"
//
// When prefixMatch is true, the name is treated as a prefix and the first
// resource whose name starts with that prefix is used (useful for CSVs
// whose names include a version suffix, e.g. "kubevirt-hyperconverged-operator.v4.21.0").
//
// The function blocks until the condition is met or the timeout expires.
func waitForCRCondition(ctx context.Context, kubeconfig string, gvr schema.GroupVersionResource,
	namespace, name, condType, expected string, timeout time.Duration, prefixMatch bool) error {

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
			return fmt.Errorf("timed out waiting for %s/%s condition %q to be %q", namespace, name, condType, expected)
		}

		obj, err := findResource(ctx, dc, gvr, namespace, name, prefixMatch)
		if err != nil {
			logging.Debugf("waiting for %s/%s: %v", namespace, name, err)
		} else if obj != nil && matchesCondition(obj, condType, expected) {
			return nil
		}
		time.Sleep(15 * time.Second)
	}
}

// findResource returns a single resource by exact name or by name prefix.
func findResource(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource,
	namespace, name string, prefixMatch bool) (*unstructured.Unstructured, error) {

	if !prefixMatch {
		return dc.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}

	list, err := dc.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
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

// matchesCondition checks whether the unstructured object satisfies the condition.
// When condType is empty it checks .status.phase directly.
// Otherwise it searches .status.conditions for a matching type and checks its status field.
func matchesCondition(obj *unstructured.Unstructured, condType, expected string) bool {
	status, ok := obj.Object["status"].(map[string]interface{})
	if !ok {
		return false
	}
	if condType == "" {
		// Scalar field check (e.g. CSV phase)
		phase, _ := status["phase"].(string)
		return phase == expected
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
