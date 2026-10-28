package ibmpower

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
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

// TestPiUserData_withPowerRunnerRegion mirrors the deploy() path for a PowerVS
// GitLab runner: it renders the real Linux setup script with PowerVSRegion set
// (as deploy() does via GetIntegrationSnippetAsCloudInitWritableFile), wraps it
// into cloud-init user-data through piUserData, and asserts the Power-specific
// podman networking settings survive the whole chain (indent + base64 +
// cloud-config).
func TestPiUserData_withPowerRunnerRegion(t *testing.T) {
	platform := gitlab.Linux
	arch := gitlab.Ppc64le
	args := &gitlab.GitLabRunnerArgs{
		URL:           "https://gitlab.example.com",
		Name:          "test-runner",
		AuthToken:     "tok",
		Platform:      &platform,
		Arch:          &arch,
		PowerVSRegion: "us-south",
	}
	gitlab.Init(args)
	glSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(args, "root")
	if err != nil {
		t.Fatalf("rendering gitlab snippet: %v", err)
	}
	out, err := piUserData("10.0.0.1", nil, *glSnippet, "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	cfg := string(decoded)
	for _, want := range []string{
		"install-glrunner.sh",
		"10.130.98.132",                    // us-south resolver
		`default_subnet = "172.20.0.0/24"`, // non-conflicting nested subnet
		`--docker-network-mode "podman"`,   // shared host network
	} {
		if !strings.Contains(cfg, want) {
			t.Errorf("Power user-data missing %q", want)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
