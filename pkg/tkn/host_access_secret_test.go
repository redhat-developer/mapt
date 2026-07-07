package tkn

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostAccessSecretEncoding(t *testing.T) {
	root := moduleRoot(t)

	for _, dir := range []string{"tkn", filepath.Join("tkn", "template")} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if !isInfraTask(name) {
				continue
			}
			checkFile(t, filepath.Join(root, dir, name))
		}
	}
}

func isInfraTask(name string) bool {
	return strings.HasSuffix(name, ".yaml") &&
		(strings.HasPrefix(name, "infra-aws-") || strings.HasPrefix(name, "infra-azure-"))
}

func checkFile(t *testing.T, path string) {
	t.Helper()

	sanitizeFields := map[string]struct{}{
		"host":             {},
		"username":         {},
		"bastion-host":     {},
		"bastion-username": {},
		"adminusername":    {},
	}
	preserveFields := map[string]struct{}{
		"id_rsa":            {},
		"bastion-id_rsa":    {},
		"userpassword":      {},
		"adminuserpassword": {},
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.Contains(line, "$(cat /opt/host-info/") {
			continue
		}
		field, ok := hostInfoField(line)
		if !ok {
			continue
		}
		if _, ok := preserveFields[field]; ok {
			if strings.Contains(line, "tr -d") {
				t.Errorf("%s: %s must not use tr -d: %s", path, field, strings.TrimSpace(line))
			}
			continue
		}
		if _, ok := sanitizeFields[field]; ok && !strings.Contains(line, `tr -d '\n\r'`) {
			t.Errorf("%s: %s must use tr -d: %s", path, field, strings.TrimSpace(line))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func hostInfoField(line string) (string, bool) {
	line = strings.TrimSpace(line)
	field, _, ok := strings.Cut(line, ":")
	if !ok {
		return "", false
	}
	return field, true
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
