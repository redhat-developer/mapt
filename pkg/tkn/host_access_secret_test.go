package tkn

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAWSHostAccessSecretEncoding(t *testing.T) {
	root := moduleRoot(t)

	for _, dir := range []string{"tkn", filepath.Join("tkn", "template")} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, "infra-aws-") || !strings.HasSuffix(name, ".yaml") {
				continue
			}
			checkFile(t, filepath.Join(root, dir, name))
		}
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.Contains(line, "$(cat /opt/host-info/") {
			continue
		}
		if strings.Contains(line, "id_rsa:") {
			if strings.Contains(line, "tr -d") {
				t.Errorf("%s: id_rsa must not use tr -d: %s", path, strings.TrimSpace(line))
			}
			continue
		}
		if !strings.Contains(line, `tr -d '\n\r'`) {
			t.Errorf("%s: host/username must use tr -d: %s", path, strings.TrimSpace(line))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
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
