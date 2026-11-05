package ibmgaudi

import (
	"encoding/base64"
	"slices"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations/otelcol"
	"gopkg.in/yaml.v3"
)

// decodeGaudiOutput extracts the base64-encoded cloud-config from the MIME
// envelope that gaudiUserData wraps it in, and returns the plain text.
func decodeGaudiOutput(t *testing.T, out string) string {
	t.Helper()
	_, payload, found := strings.Cut(out, "Content-Transfer-Encoding: base64\n\n")
	if !found {
		t.Fatalf("no base64 section in user data: %s", out)
	}
	encoded, _, found := strings.Cut(payload, "\n--MAPT-CLOUD-CONFIG")
	if !found {
		t.Fatalf("no closing boundary in user data: %s", out)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode MIME payload: %v", err)
	}
	return string(decoded)
}

// sshd rejects key logins once the login user's password ages out, so aging
// must be off whether or not otel is configured, and before otel installs.
func TestGaudiUserData_disablesPasswordAging(t *testing.T) {
	otelArgs := &otelcol.OtelcolArgs{
		AppCode:   "MYAPP",
		AuthToken: "tok",
		Endpoint:  "https://otel.example.com",
		Index:     "my-index",
		Arch:      otelcol.Amd64,
	}
	for name, args := range map[string]*otelcol.OtelcolArgs{
		"no otel":   nil,
		"with otel": otelArgs,
	} {
		t.Run(name, func(t *testing.T) {
			out, err := gaudiUserData(args)
			if err != nil {
				t.Fatalf("gaudiUserData returned error: %v", err)
			}
			cfg := decodeGaudiOutput(t, out)
			var parsed struct {
				Runcmd []string `yaml:"runcmd"`
			}
			if err := yaml.Unmarshal([]byte(cfg), &parsed); err != nil {
				t.Fatalf("cloud-config is not valid YAML: %v", err)
			}
			want := []string{"chage -M -1 -I -1 -E -1 " + defaultUser}
			if args != nil {
				want = append(want, "bash /opt/install-otelcol.sh")
			}
			if !slices.Equal(parsed.Runcmd, want) {
				t.Fatalf("expected runcmd %q, got %q", want, parsed.Runcmd)
			}
		})
	}
}
