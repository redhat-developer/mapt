package ibmpower

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations/otelcol"
)

func TestPiUserData_noRunner(t *testing.T) {
	out, err := piUserData("10.0.0.1", nil, "", "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	cfg := string(decoded)
	if !strings.HasPrefix(cfg, "#cloud-config") {
		t.Errorf("expected #cloud-config header, got: %s", cfg[:min(len(cfg), 40)])
	}
	if strings.Contains(cfg, "install-glrunner") {
		t.Error("expected no GitLab runner section when script is empty")
	}
	if !strings.Contains(cfg, "mount-data-home") {
		t.Error("expected mount-data-home service in write_files regardless of runner/otel config")
	}
}

func TestPiUserData_withRunner(t *testing.T) {
	script := "      #!/bin/bash\n      echo hello"
	out, err := piUserData("10.0.0.1", nil, script, "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	cfg := string(decoded)
	if !strings.Contains(cfg, "install-glrunner.sh") {
		t.Error("expected install-glrunner.sh in write_files")
	}
	if !strings.Contains(cfg, "bash /opt/install-glrunner.sh") {
		t.Error("expected runcmd entry to execute the runner script")
	}
	if !strings.Contains(cfg, "write_files") {
		t.Error("expected write_files section")
	}
}

func TestPiUserData_withOtelAndRunner(t *testing.T) {
	script := "      #!/bin/bash\n      echo hello"
	args := &otelcol.OtelcolArgs{
		AppCode:             "MYAPP",
		AuthToken:           "tok",
		Endpoint:            "https://otel.example.com",
		Index:               "my-index",
		Arch:                otelcol.Ppc64le,
		SyslogPath:          "/var/log/messages",
		SecurePath:          "/var/log/secure",
		MonitorGitLabRunner: true,
	}
	out, err := piUserData("10.0.0.1", args, script, "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	cfg := string(decoded)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
