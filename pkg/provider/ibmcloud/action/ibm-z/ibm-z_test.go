package ibmz

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations/otelcol"
)

// decodeIzOutput extracts the base64-encoded cloud-config from the MIME
// envelope that izUserData wraps it in, and returns the plain text.
func decodeIzOutput(t *testing.T, out string) string {
	t.Helper()
	lines := strings.Split(out, "\n")
	inContent := false
	var b64Lines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "Content-Transfer-Encoding: base64") {
			inContent = true
			continue
		}
		if inContent {
			if line == "" && len(b64Lines) == 0 {
				continue
			}
			if strings.HasPrefix(line, "--MAPT-CLOUD-CONFIG") {
				break
			}
			b64Lines = append(b64Lines, line)
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.Join(b64Lines, ""))
	if err != nil {
		t.Fatalf("failed to decode MIME payload: %v", err)
	}
	return string(decoded)
}

func TestIzUserData_noRunner(t *testing.T) {
	out, err := izUserData(nil, "", "")
	if err != nil {
		t.Fatalf("izUserData returned error: %v", err)
	}
	cfg := decodeIzOutput(t, out)
	if strings.Contains(cfg, "install-glrunner") {
		t.Error("expected no GitLab runner section when script is empty")
	}
	if strings.Contains(cfg, "write_files") {
		t.Error("expected no write_files when otel and runner are both absent")
	}
	if !strings.Contains(cfg, "apt-get install") {
		t.Error("expected package install in runcmd")
	}
}

func TestIzUserData_withRunner(t *testing.T) {
	script := "      #!/bin/bash\n      echo hello"
	out, err := izUserData(nil, script, "")
	if err != nil {
		t.Fatalf("izUserData returned error: %v", err)
	}
	cfg := decodeIzOutput(t, out)
	if !strings.Contains(cfg, "install-glrunner.sh") {
		t.Error("expected install-glrunner.sh in write_files")
	}
	if !strings.Contains(cfg, "bash /opt/install-glrunner.sh") {
		t.Error("expected runcmd entry to execute the runner script")
	}
}

func TestIzUserData_withOtelAndRunner(t *testing.T) {
	script := "      #!/bin/bash\n      echo hello"
	args := &otelcol.OtelcolArgs{
		AppCode:             "MYAPP",
		AuthToken:           "tok",
		Endpoint:            "https://otel.example.com",
		Index:               "my-index",
		Arch:                otelcol.S390x,
		SyslogPath:          "/var/log/syslog",
		SecurePath:          "/var/log/auth.log",
		MonitorGitLabRunner: true,
	}
	out, err := izUserData(args, script, "")
	if err != nil {
		t.Fatalf("izUserData returned error: %v", err)
	}
	cfg := decodeIzOutput(t, out)
	if !strings.Contains(cfg, "otelcol-contrib") {
		t.Error("expected otel section")
	}
	if !strings.Contains(cfg, "install-glrunner.sh") {
		t.Error("expected GitLab runner section")
	}
	if !strings.Contains(cfg, "filelog/gitlab-runner") {
		t.Error("expected gitlab-runner filelog receiver in otelcol config")
	}
	if !strings.Contains(cfg, "/var/log/gitlab-runner/runner.log") {
		t.Error("expected runner log path in otelcol config")
	}
	if !strings.Contains(cfg, "filelog/syslog, filelog/secure, filelog/audit, filelog/gitlab-runner") {
		t.Error("expected gitlab-runner included in otelcol pipeline receivers")
	}
	if !strings.Contains(cfg, "StandardOutput=append:/var/log/gitlab-runner") {
		t.Error("expected systemd drop-in to redirect runner stdout to log file")
	}
}
