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

// TestPiUserDataMountDataHomeWritesFstabForFreshVolume guards the fresh-volume
// fstab entry. The only UUID probe runs before mkfs.xfs, when a blank volume
// carries no UUID yet, so the entry can only be written if the helper re-reads
// the UUID after creating the filesystem. Live validation of #945 on a fresh
// PowerVS data volume showed the entry was skipped with "no UUID" and /home
// only became fstab-mounted on the next boot, once the filesystem existed.
func TestPiUserDataMountDataHomeWritesFstabForFreshVolume(t *testing.T) {
	script := extractMountDataHomeScript(t, decodePiUserData(t))

	mkfs := strings.Index(script, "mkfs.xfs -f -K")
	if mkfs < 0 {
		t.Fatal("expected mkfs.xfs in the helper")
	}
	appendEntry := strings.Index(script, `echo "UUID=$UUID /home $FS_TYPE defaults 0 2" >> /etc/fstab`)
	if appendEntry < 0 {
		t.Fatal("expected the /home fstab append in the helper")
	}
	// The fstab entry is written from $UUID, so a probe must exist between
	// mkfs.xfs and the append; without one a fresh volume never gets an
	// fstab entry and /home stays unmounted at boot until the service runs.
	segment := script[mkfs:appendEntry]
	if probe := strings.Index(segment, "UUID=$(blkid"); probe < 0 {
		t.Error("helper must re-read the UUID after mkfs.xfs; the first probe ran before the filesystem existed, so the fstab entry is skipped for fresh volumes")
	}
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

// decodePiUserData renders the default PowerVS user-data and returns the
// decoded cloud-config.
func decodePiUserData(t *testing.T) string {
	t.Helper()
	out, err := piUserData("10.0.0.1", nil, "", "")
	if err != nil {
		t.Fatalf("piUserData returned error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}
	return string(decoded)
}

// TestPiUserDataMountDataHomeBindGuard guards the graphroot bind. The drift
// paths exit before the bind, but a graphroot directory that lives inside a
// different mount must also be left alone: binding the (normally empty)
// /home/containers/storage over it hides a live store behind an empty
// directory, and the runner then pulls every image again.
func TestPiUserDataMountDataHomeBindGuard(t *testing.T) {
	script := extractMountDataHomeScript(t, decodePiUserData(t))
	bind := strings.Index(script, "mount --bind /home/containers/storage /var/lib/containers/storage")
	if bind < 0 {
		t.Fatal("expected the graphroot bind in the helper")
	}
	guard := strings.LastIndex(script[:bind], "findmnt -T /var/lib/containers/storage")
	if guard < 0 {
		t.Fatal("the bind must check where /var/lib/containers/storage is mounted first")
	}
	if !strings.Contains(script[guard:bind], `[ "$GRAPHROOT_MNT" = / ]`) {
		t.Error("the bind must only run when the graphroot directory is on the root filesystem")
	}
}

// TestPiUserDataMountDataHomeFstabHomeEntry closes the duplicate-/home gap: an
// fstab entry for /home that does not name this device is an existing layout.
// Appending another /home line would leave two entries for one mountpoint and
// make the boot-time result depend on fstab order.
func TestPiUserDataMountDataHomeFstabHomeEntry(t *testing.T) {
	script := extractMountDataHomeScript(t, decodePiUserData(t))
	mountHome := strings.Index(script, `mount "$DATA_DEV" /home`)
	if mountHome < 0 {
		t.Fatal("expected the /home mount in the helper")
	}
	lookup := strings.LastIndex(script[:mountHome], `$2 == "/home"`)
	if lookup < 0 {
		t.Fatal("helper must look up an existing /home fstab entry before mounting")
	}
	block := script[lookup:mountHome]
	if !strings.Contains(block, `[ "$FSTAB_HOME" != "UUID=${UUID}" ]`) {
		t.Error("helper must treat an fstab /home entry for another device as an existing layout")
	}
	if exit := strings.Index(block, "exit 0"); exit < 0 || !strings.Contains(block[:exit], "mkdir -p /home/containers/storage") {
		t.Error("fstab /home-conflict path must create /home/containers/storage before exiting")
	}
}

// TestPiUserDataMountDataHomeFailureStopsInstall guards visibility: cloud-init
// reports runcmd failure but keeps going, so a failed mount service used to
// leave a healthy-looking instance running on root-disk container storage.
// cloud-init runs all runcmd entries as one #!/bin/sh script, so this executes
// the rendered block with a stub systemctl to verify the abort, rather than
// asserting where "exit 1" happens to sit in the text.
func TestPiUserDataMountDataHomeFailureStopsInstall(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	cfg := decodePiUserData(t)
	start := strings.Index(cfg, "if ! systemctl start mount-data-home.service; then")
	if start < 0 {
		t.Fatal("cloud-init must fail the install when mount-data-home.service fails")
	}
	rest := cfg[start:]
	end := strings.Index(rest, "\n    fi")
	if end < 0 {
		t.Fatal("could not locate the end of the mount service check")
	}
	block := rest[:end+len("\n    fi")]

	dir := t.TempDir()
	stub := filepath.Join(dir, "systemctl")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit \"$STUB_RC\"\n"), 0o755); err != nil {
		t.Fatalf("write systemctl stub: %v", err)
	}
	for _, tc := range []struct {
		name      string
		rc        string
		wantAfter bool
	}{
		{"service failure aborts the rest", "1", false},
		{"service success continues", "0", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := "#!/bin/sh\n" + block + "\necho AFTER\n"
			cmd := exec.Command("sh", "-c", script)
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "STUB_RC="+tc.rc)
			out, _ := cmd.CombinedOutput()
			if got := strings.Contains(string(out), "AFTER"); got != tc.wantAfter {
				t.Errorf("subsequent commands ran = %v; want %v\n%s", got, tc.wantAfter, out)
			}
			if !tc.wantAfter && !strings.Contains(string(out), "journalctl -u mount-data-home.service") {
				t.Errorf("failure output must point at the service journal:\n%s", out)
			}
		})
	}
}

// TestPiUserDataMountDataHomeFstabLookup exercises the awk program that finds
// an existing /home entry against realistic fstab shapes. The guard is only
// correct if comments are ignored and other mount points do not match.
func TestPiUserDataMountDataHomeFstabLookup(t *testing.T) {
	script := extractMountDataHomeScript(t, decodePiUserData(t))
	var line string
	for _, l := range strings.Split(script, "\n") {
		if strings.Contains(l, `/^#/ && $2 == "/home"`) {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatal("helper has no /home fstab lookup")
	}
	start := strings.Index(line, "awk '")
	if start < 0 {
		t.Fatalf("could not find the awk program in %q", line)
	}
	rest := line[start+len("awk '"):]
	end := strings.Index(rest, "' /etc/fstab")
	if end < 0 {
		t.Fatalf("could not find the end of the awk program in %q", line)
	}
	program := rest[:end]

	cases := []struct{ name, fstab, want string }{
		{"plain entry", "UUID=aaaa /home xfs defaults 0 2\n", "UUID=aaaa"},
		{"tab separated", "UUID=aaaa\t/home\txfs\tdefaults\t0 2\n", "UUID=aaaa"},
		{"commented entry", "# UUID=aaaa /home xfs defaults 0 2\n", ""},
		{"commented without space", "#UUID=aaaa /home xfs defaults 0 2\n", ""},
		{"different mount point", "UUID=aaaa /var/lib/containers xfs defaults 0 2\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fstab := filepath.Join(t.TempDir(), "fstab")
			if err := os.WriteFile(fstab, []byte(tc.fstab), 0o600); err != nil {
				t.Fatalf("write fstab: %v", err)
			}
			out, err := exec.Command("awk", program, fstab).Output()
			if err != nil {
				t.Fatalf("running awk: %v", err)
			}
			if got := strings.TrimSpace(string(out)); got != tc.want {
				t.Errorf("lookup = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestPiUserDataMountDataHomeSkipsResetOnAdoptedLayout guards the unconditional
// `podman system reset --force`. The reset deletes the images and containers at
// the effective graphroot, so it must only run when this boot created the
// filesystem; on an adopted volume or a drifted layout it would destroy live
// container storage.
func TestPiUserDataMountDataHomeSkipsResetOnAdoptedLayout(t *testing.T) {
	cfg := decodePiUserData(t)
	script := extractMountDataHomeScript(t, cfg)

	touch := strings.Index(script, "touch /run/mount-data-home-fresh")
	if touch < 0 {
		t.Fatal("helper must record that it created a fresh filesystem")
	}
	if guard := strings.LastIndex(script[:touch], `if [ "$FRESH_FS" -eq 1 ]`); guard < 0 {
		t.Error("the fresh-filesystem marker must only be written when the helper formatted the volume")
	}

	reset := strings.Index(cfg, "podman system reset --force")
	if reset < 0 {
		t.Fatal("expected podman system reset in cloud-init")
	}
	if gate := strings.LastIndex(cfg[:reset], "-f /run/mount-data-home-fresh"); gate < 0 {
		t.Error("podman system reset must be gated on a freshly created filesystem")
	}
}

func TestPiUserDataMountDataHomeResetRequiresFreshGraphroot(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	cfg := decodePiUserData(t)
	helper := extractMountDataHomeScript(t, cfg)
	start := strings.Index(cfg, "if [ -f /run/mount-data-home-fresh ]; then")
	if start < 0 {
		t.Fatal("reset authorization block not found")
	}
	end := strings.Index(cfg[start:], "\n    fi")
	if end < 0 {
		t.Fatal("reset authorization block is incomplete")
	}
	reset := cfg[start : start+end+len("\n    fi")]
	cases := []struct {
		name             string
		fresh            string
		homeMounted      string
		graphrootMounted string
		graphrootMount   string
		bindRC           string
		staleMarker      bool
		wantBind         bool
		wantReset        bool
		wantFailed       bool
	}{
		{"fresh store bound to graphroot", "1", "0", "0", "/", "0", false, true, true, false},
		{"existing store beneath another mount", "1", "0", "0", "/var/lib/containers", "0", false, false, false, false},
		{"unknown graphroot mount", "1", "0", "0", "", "0", false, false, false, false},
		{"existing graphroot mountpoint", "1", "0", "1", "", "0", false, false, false, false},
		{"adopted filesystem", "0", "0", "0", "/", "0", false, true, false, false},
		{"adopted filesystem with stale marker", "0", "0", "0", "/", "0", true, true, false, false},
		{"mounted home with stale marker", "0", "1", "0", "/", "0", true, true, false, false},
		{"failed bind clears stale marker", "1", "0", "0", "/", "1", true, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			home := filepath.Join(dir, "home")
			graphroot := filepath.Join(dir, "graphroot")
			marker := filepath.Join(dir, "fresh-marker")
			bindCalls := filepath.Join(dir, "bind-calls")
			resetCalls := filepath.Join(dir, "reset-calls")
			for _, path := range []string{home, graphroot} {
				if err := os.MkdirAll(path, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range []string{filepath.Join(dir, "fstab"), bindCalls, resetCalls} {
				if err := os.WriteFile(path, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.staleMarker {
				if err := os.WriteFile(marker, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			replace := strings.NewReplacer(
				"/var/lib/containers/storage", graphroot,
				"/run/mount-data-home-fresh", marker,
				"/mnt/home-tmp", filepath.Join(dir, "migration"),
				"/etc/fstab", filepath.Join(dir, "fstab"),
				"/home", home,
			)
			// Execute the complete helper and its real runcmd consumer. All
			// device operations are stubbed; real file operations stay in dir.
			script := `set -euo pipefail
systemctl() { return 0; }
sleep() { return 0; }
udevadm() { return 0; }
mountpoint() {
    if [ "$2" = "$TEST_HOME" ]; then [ "$TEST_HOME_MOUNTED" = 1 ];
    elif [ "$2" = "$TEST_GRAPHROOT" ]; then [ "$TEST_GRAPHROOT_MOUNTED" = 1 ];
    else return 99; fi
}
lsblk() { if [ "$1" = -rnpo ]; then printf '/dev/mapper/test-volume mpath\n'; fi; }
blkid() {
    case " $* " in
        *' TYPE '*) if [ "$TEST_FRESH" = 1 ]; then return 2; else printf 'xfs\n'; fi ;;
        *' UUID '*) if [ "$TEST_FRESH" = 0 ] || [ -f "$FORMATTED" ]; then printf 'test-uuid\n'; fi ;;
        *) return 99 ;;
    esac
}
dd() { return 0; }
mkfs.xfs() { touch "$FORMATTED"; }
mount() {
    if [ "$1" = --bind ]; then
        printf '%s\n' "$*" >> "$BIND_CALLS"
        return "$TEST_BIND_RC"
    fi
}
findmnt() { if [ -n "$TEST_GRAPHROOT_MOUNT" ]; then printf '%s\n' "$TEST_GRAPHROOT_MOUNT"; else return 1; fi; }
podman() { printf '%s\n' "$*" >> "$RESET_CALLS"; }
` + replace.Replace(helper+"\n"+reset)
			cmd := exec.Command("bash", "-c", script)
			cmd.Env = append(os.Environ(),
				"TEST_HOME="+home, "TEST_GRAPHROOT="+graphroot,
				"TEST_FRESH="+tc.fresh, "TEST_HOME_MOUNTED="+tc.homeMounted,
				"TEST_GRAPHROOT_MOUNTED="+tc.graphrootMounted,
				"TEST_GRAPHROOT_MOUNT="+tc.graphrootMount, "TEST_BIND_RC="+tc.bindRC,
				"FORMATTED="+filepath.Join(dir, "formatted"),
				"BIND_CALLS="+bindCalls, "RESET_CALLS="+resetCalls,
			)
			out, runErr := cmd.CombinedOutput()
			if (runErr != nil) != tc.wantFailed {
				t.Fatalf("helper failure=%v, want %v: %v\n%s", runErr != nil, tc.wantFailed, runErr, out)
			}
			binds, err := os.ReadFile(bindCalls)
			if err != nil {
				t.Fatal(err)
			}
			if (len(binds) > 0) != tc.wantBind {
				t.Errorf("bind attempted=%v, want %v\n%s", len(binds) > 0, tc.wantBind, out)
			}
			resets, err := os.ReadFile(resetCalls)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if tc.wantReset {
				want = "system reset --force\n"
			}
			if string(resets) != want {
				t.Errorf("reset calls=%q, want %q\n%s", resets, want, out)
			}
			_, markerErr := os.Stat(marker)
			if (markerErr == nil) != tc.wantReset {
				t.Errorf("reset marker exists=%v, want %v\n%s", markerErr == nil, tc.wantReset, out)
			}
		})
	}
}
