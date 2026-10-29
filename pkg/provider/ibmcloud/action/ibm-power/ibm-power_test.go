package ibmpower

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
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

// TestPiUserDataMountDataHomeSyntax extracts the mount-data-home.sh helper
// from the rendered user-data and syntax-checks it with bash -n. The helper
// runs via mount-data-home.service on first boot; a bash syntax error in it
// silently disables the data-volume mount (the loop structure was broken in
// 465926fc5 and never parsed — found while validating #894).
func TestPiUserDataMountDataHomeSyntax(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	out, err := piUserData("10.0.0.1", nil, "", "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	script := extractMountDataHomeScript(t, string(decoded))
	path := filepath.Join(t.TempDir(), "mount-data-home.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write helper: %v", err)
	}
	if out, err := exec.Command("bash", "-n", path).CombinedOutput(); err != nil {
		t.Errorf("bash -n failed for mount-data-home.sh: %v\n%s", err, out)
	}
}

// TestPiUserDataMountDataHomeAdoptsExistingFilesystem guards the data-safety
// property of the helper: only format the data volume when it carries no
// filesystem, and record the detected type in fstab. Repairing the retry loop
// (#945) made the previously inert format path reachable, so on a host whose
// mpath device already holds live data an unconditional mkfs.xfs would destroy
// it.
func TestPiUserDataMountDataHomeAdoptsExistingFilesystem(t *testing.T) {
	out, err := piUserData("10.0.0.1", nil, "", "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	script := extractMountDataHomeScript(t, string(decoded))

	// -p probes the device directly instead of trusting the blkid cache.
	if !strings.Contains(script, "blkid -p -s TYPE") {
		t.Error("helper must probe the device for an existing filesystem type before formatting")
	}
	mkfs := strings.Index(script, "mkfs.xfs -f -K")
	if mkfs < 0 {
		t.Fatal("expected mkfs.xfs in the helper")
	}
	if guard := strings.LastIndex(script[:mkfs], `if [ -z "$FS_TYPE" ]`); guard < 0 {
		t.Error("mkfs.xfs must only run when no filesystem type was detected")
	}
	if strings.Contains(script, "/home xfs defaults") {
		t.Error("fstab entry must use the detected filesystem type, not a hardcoded xfs")
	}
	if !strings.Contains(script, "/home $FS_TYPE defaults") {
		t.Error("expected fstab entry to use the detected filesystem type")
	}
	// blkid exits 2 when nothing is detected, but also when the device cannot
	// be read, and exits 0 with no TYPE for a non-filesystem signature such as
	// a partition table. Neither may fall through to formatting.
	if !strings.Contains(script, "BLKID_RC") {
		t.Error("helper must distinguish a failed blkid probe from a missing signature")
	}
	if guard := strings.LastIndex(script[:mkfs], `dd if="$DATA_DEV" of=/dev/null`); guard < 0 {
		t.Error("helper must confirm the device is readable before treating it as blank")
	}
	if guard := strings.LastIndex(script[:mkfs], `[ "$BLKID_RC" -eq 0 ] && [ -z "$FS_TYPE" ]`); guard < 0 {
		t.Error("helper must refuse to format a device carrying a non-filesystem signature")
	}
	// A device already mounted elsewhere belongs to an existing layout and
	// must be left alone rather than mounted a second time at /home.
	mounted := strings.LastIndex(script[:mkfs], "lsblk -dnro MOUNTPOINT")
	if mounted < 0 {
		t.Fatal("helper must skip a data device that is already mounted")
	}
	// cloud-init runs restorecon on /home/containers/storage after the
	// service, so the early exit must still leave the directory in place.
	skip := script[mounted:mkfs]
	if exit := strings.Index(skip, "exit 0"); exit < 0 || !strings.Contains(skip[:exit], "mkdir -p /home/containers/storage") {
		t.Error("already-mounted path must create /home/containers/storage before exiting")
	}
	// An fstab entry for the same UUID at another path is an existing layout:
	// it must be detected before the device is mounted at /home, and the
	// helper must still leave the storage directory for restorecon.
	mountHome := strings.Index(script, `mount "$DATA_DEV" /home`)
	if mountHome < 0 {
		t.Fatal("expected the /home mount in the helper")
	}
	lookup := strings.LastIndex(script[:mountHome], `awk -v u="UUID=${UUID}" '$1 == u {print $2; exit}' /etc/fstab`)
	if lookup < 0 {
		t.Fatal("helper must look up an existing fstab entry before mounting /home")
	}
	conflict := script[lookup:mountHome]
	if !strings.Contains(conflict, `[ "$FSTAB_TARGET" != /home ]`) {
		t.Error("helper must leave the layout alone when fstab targets another path")
	}
	if exit := strings.Index(conflict, "exit 0"); exit < 0 || !strings.Contains(conflict[:exit], "mkdir -p /home/containers/storage") {
		t.Error("fstab-conflict path must create /home/containers/storage before exiting")
	}
	if !strings.Contains(script[mountHome:], `elif [ "$FSTAB_TARGET" = /home ]; then`) {
		t.Error("helper must not append a duplicate /home fstab entry for the same UUID")
	}
	// Migrating the root filesystem's /home onto an adopted volume would
	// overwrite paths that already hold live data.
	if !strings.Contains(script, "FRESH_FS") {
		t.Error("helper must track whether it created the filesystem")
	}
	homeCopy := strings.Index(script, "cp -a /home/. /mnt/home-tmp/")
	if homeCopy < 0 {
		t.Fatal("expected the /home migration copy in the helper")
	}
	if guard := strings.LastIndex(script[:homeCopy], `if [ "$FRESH_FS" -eq 1 ]`); guard < 0 {
		t.Error("the /home migration copy must only run for a newly created filesystem")
	}
}

// extractMountDataHomeScript pulls the mount-data-home.sh helper out of the
// rendered cloud-config. The helper is the write_files content block for
// /usr/local/bin/mount-data-home.sh: six-space indented lines between its
// "content: |" and the next write_files entry.
func extractMountDataHomeScript(t *testing.T, cfg string) string {
	t.Helper()
	var script []string
	inHelper := false
	inContent := false
	for _, line := range strings.Split(cfg, "\n") {
		switch {
		case strings.Contains(line, "path: /usr/local/bin/mount-data-home.sh"):
			inHelper = true
		case inHelper && !inContent:
			if strings.Contains(line, "content: |") {
				inContent = true
			}
		case inContent && strings.HasPrefix(line, "  - path:"):
			// next write_files entry ends the helper
		case inContent && strings.HasPrefix(line, "      "):
			script = append(script, strings.TrimPrefix(line, "      "))
		case inContent && strings.TrimSpace(line) == "":
			script = append(script, "")
		}
		if inContent && strings.HasPrefix(line, "  - path:") {
			break
		}
	}
	if len(script) == 0 {
		t.Fatal("could not extract mount-data-home.sh from rendered user-data")
	}
	return strings.Join(script, "\n") + "\n"
}
