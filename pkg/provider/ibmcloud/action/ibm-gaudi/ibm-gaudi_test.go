package ibmgaudi

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations/otelcol"
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
			chage := "- chage -M -1 -I -1 -E -1 " + defaultUser + "\n"
			at := strings.Index(cfg, chage)
			if at == -1 {
				t.Fatalf("expected %q in cloud-config:\n%s", chage, cfg)
			}
			otelCmd := strings.Index(cfg, "- bash /opt/install-otelcol.sh")
			if (args != nil) != (otelCmd != -1) {
				t.Errorf("otel install step does not match otel config:\n%s", cfg)
			}
			if otelCmd != -1 && at > otelCmd {
				t.Error("expected password aging to be disabled before the otel install")
			}
		})
	}
}
