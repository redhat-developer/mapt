package tkn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tasks that have a compute-sizes conditional in their script and must also
// conditionally pass --compute-families in the else branch.
var tasksWithComputeFamiliesScript = map[string]struct{}{
	"infra-aws-rhel.yaml":            {},
	"infra-aws-ocp-snc.yaml":         {},
	"infra-aws-fedora.yaml":          {},
	"infra-aws-rhel-ai.yaml":         {},
	"infra-aws-eks.yaml":             {},
	"infra-aws-kind.yaml":            {},
	"infra-aws-windows-server.yaml":  {},
}

// mac uses dedicated host provisioning — CLI does not accept --compute-families.
var tasksWithoutComputeFamiliesParam = map[string]struct{}{
	"infra-aws-mac.yaml": {},
}

func TestComputeFamiliesParamDefined(t *testing.T) {
	root := moduleRoot(t)
	for _, dir := range []string{"tkn", filepath.Join("tkn", "template")} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if !isAWSInfraTask(name) {
				continue
			}
			if _, skip := tasksWithoutComputeFamiliesParam[name]; skip {
				continue
			}
			path := filepath.Join(root, dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "name: compute-families") {
				t.Errorf("%s: missing 'compute-families' param definition", path)
			}
		}
	}
}

func TestComputeFamiliesPassedInScript(t *testing.T) {
	root := moduleRoot(t)
	for _, dir := range []string{"tkn", filepath.Join("tkn", "template")} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if _, ok := tasksWithComputeFamiliesScript[name]; !ok {
				continue
			}
			path := filepath.Join(root, dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "--compute-families") {
				t.Errorf("%s: missing '--compute-families' flag in script", path)
			}
		}
	}
}

func isAWSInfraTask(name string) bool {
	return strings.HasSuffix(name, ".yaml") && strings.HasPrefix(name, "infra-aws-")
}
