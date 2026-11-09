package ibmpower

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const reclaimScript = "reclaim-ipam.sh"

const loggerStub = `#!/bin/sh
printf '%s\n' "$*" >> "$LOGGER_OUT"
`

// podmanStub answers the two calls the reclaimer makes. Behaviour is driven by
// environment variables so a test only has to set the ones it cares about.
const podmanStub = `#!/bin/sh
case "$1" in
  ps)
    printf '%s\n' "${STUB_PS_STATES:-}"
    exit "${STUB_PS_EXIT:-0}"
    ;;
  events)
    printf 'call\n' >> "$STUB_EVENTS_CALLS"
    printf '%s\n' "$*" >> "$STUB_EVENTS_ARGS"
    if [ -n "${STUB_EVENTS_MUTATE_DB:-}" ]; then
      printf '10.88.9.9\000' >> "$STUB_EVENTS_MUTATE_DB"
    fi
    if [ -n "${STUB_EVENTS_DELETE_DB:-}" ]; then
      rm -f "$STUB_EVENTS_DELETE_DB"
    fi
    if [ -n "${STUB_EVENTS_STDERR:-}" ]; then
      printf '%s\n' "$STUB_EVENTS_STDERR" >&2
    fi
    ids="${STUB_EVENTS_IDS:-}"
    if [ -n "$ids" ] && [ -s "$ids" ]; then
      head -n 1 "$ids"
      tail -n +2 "$ids" > "$ids.tmp"
      mv "$ids.tmp" "$ids"
    fi
    exit "${STUB_EVENTS_EXIT:-0}"
    ;;
  network)
    case "$2" in
      ls)
        printf '%s\n' "${STUB_NETWORK_NAMES:-podman}"
        ;;
      inspect)
        case "$3" in
          podman)
            printf '%s\n' "${STUB_NETWORKS_PODMAN:-${STUB_NETWORKS:-192.168.100.0/24|}}"
            ;;
          *)
            printf '%s\n' "${STUB_NETWORKS_EXTRA:-${STUB_NETWORKS:-192.168.100.0/24|}}"
            ;;
        esac
        ;;
    esac
    exit "${STUB_NETWORKS_EXIT:-0}"
    ;;
esac
exit 1
`

// stubPath builds a PATH directory holding the stubbed commands, plus a
// `timeout` stub that drops its first argument and execs the rest (macOS has
// no GNU timeout) and a `sleep` stub that returns immediately so the
// re-check tests do not spend five seconds each. It prepends that directory to
// PATH.
func stubPath(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	stubs := map[string]string{
		"timeout": "#!/bin/sh\nshift\nexec \"$@\"\n",
		"sleep":   "#!/bin/sh\nexit 0\n",
		"logger":  loggerStub,
		"podman":  podmanStub,
	}
	for name, body := range stubs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("NETAVARK_LOCK", filepath.Join(t.TempDir(), "netavark.lock"))
}

// newIPAMDB writes a lease db holding n distinct leases and returns its path.
func newIPAMDB(t *testing.T, n int) string {
	t.Helper()
	return newIPAMDBIn(t, "192.168", n)
}

// newIPAMDBIn writes n leases under the given two-octet base, e.g. "10.88".
func newIPAMDBIn(t *testing.T, base string, n int) string {
	t.Helper()
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(base + "." + strconv.Itoa(i/254) + "." + strconv.Itoa(i%254+1) + "\x00")
	}
	return writeDB(t, b.String())
}

func writeDB(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ipam.db")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write ipam db: %v", err)
	}
	return path
}

// setPodman configures what the podman stub reports. eventsIDs are handed to
// successive `podman events` calls, one per call; an empty entry means that
// call reports no events. It returns the path of the events call log.
func setPodman(t *testing.T, psStates string, psExit int, eventsIDs []string, eventsExit int) string {
	t.Helper()
	calls := filepath.Join(t.TempDir(), "events.calls")
	t.Setenv("STUB_PS_STATES", psStates)
	t.Setenv("STUB_PS_EXIT", strconv.Itoa(psExit))
	t.Setenv("STUB_EVENTS_CALLS", calls)
	t.Setenv("STUB_EVENTS_ARGS", filepath.Join(t.TempDir(), "events.args"))
	t.Setenv("STUB_EVENTS_EXIT", strconv.Itoa(eventsExit))
	ids := filepath.Join(t.TempDir(), "events.ids")
	if len(eventsIDs) > 0 {
		if err := os.WriteFile(ids, []byte(strings.Join(eventsIDs, "\n")+"\n"), 0o600); err != nil {
			t.Fatalf("write event ids: %v", err)
		}
	}
	t.Setenv("STUB_EVENTS_IDS", ids)
	return calls
}

// setPool configures the CIDR the podman network stub reports and its exit code.
func setPool(t *testing.T, cidr string, exitCode int) {
	t.Helper()
	t.Setenv("STUB_NETWORKS", cidr+"|")
	t.Setenv("STUB_NETWORK_NAMES", "podman")
	t.Setenv("STUB_NETWORKS_EXIT", strconv.Itoa(exitCode))
}

// runReclaimer runs the script against the stubs and returns its exit code and
// everything the logger stub captured.
func runReclaimer(t *testing.T) (int, string) {
	t.Helper()
	logged := filepath.Join(t.TempDir(), "logger.out")
	t.Setenv("LOGGER_OUT", logged)
	cmd := exec.Command("bash", reclaimScript)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("run reclaimer: %v (stderr: %s)", err, stderr.String())
	}
	out, readErr := os.ReadFile(logged)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("read logger output: %v", readErr)
	}
	return code, string(out)
}

func assertDBExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected lease db %s to still exist: %v", path, err)
	}
}

func countEvents(t *testing.T, calls string) int {
	t.Helper()
	out, err := os.ReadFile(calls)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("read events call log: %v", err)
	}
	return strings.Count(string(out), "call")
}

func eventsArgs(t *testing.T) string {
	t.Helper()
	path := os.Getenv("STUB_EVENTS_ARGS")
	if path == "" {
		t.Fatal("events args path not configured")
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read events args: %v", err)
	}
	return string(out)
}

func TestReclaimIPAM_noDB(t *testing.T) {
	stubPath(t)
	setPodman(t, "", 0, nil, 0)
	db := filepath.Join(t.TempDir(), "ipam.db")
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Error("reclaimer created the lease db")
	}
	if logged != "" {
		t.Errorf("expected no log line, got %q", logged)
	}
}

func TestReclaimIPAM_belowThreshold(t *testing.T) {
	stubPath(t)
	setPodman(t, "", 0, nil, 0)
	db := newIPAMDB(t, 5)
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if logged != "" {
		t.Errorf("expected no log line below the threshold, got %q", logged)
	}
}

func TestReclaimIPAM_runningContainer(t *testing.T) {
	stubPath(t)
	setPodman(t, "running", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "busy") {
		t.Errorf("expected a busy log line while over the threshold, got %q", logged)
	}
}

func TestReclaimIPAM_transientContainerState(t *testing.T) {
	stubPath(t)
	setPodman(t, "stopping", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
}

func TestReclaimIPAM_eventsDoesNotStream(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if got := eventsArgs(t); !strings.Contains(got, "--stream=false") {
		t.Errorf("podman events must not stream, got %q", got)
	}
}

func TestReclaimIPAM_countsAnyPool(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	// podman's default network is 10.88.0.0/16, not the /24 the issue quotes.
	setPool(t, "10.88.0.0/22", 0)
	db := newIPAMDBIn(t, "10.88", 900) // 85% of a /22 is 868
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Error("expected a full pool in any subnet to be reclaimed")
	}
}

func TestReclaimIPAM_largePoolIsNotTouched(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	setPool(t, "10.88.0.0/16", 0) // 85% of a /16 is 55651
	db := newIPAMDBIn(t, "10.88", 250)
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if logged != "" {
		t.Errorf("expected no log line far below capacity, got %q", logged)
	}
}

func TestReclaimIPAM_unknownPoolCapacity(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	setPool(t, "", 1)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "cannot determine the podman subnet") {
		t.Errorf("expected a log line about the unknown subnet, got %q", logged)
	}
}

func TestReclaimIPAM_eventsReadErrorSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	// podman 5.4 exits 0 after a journald read error and only logs it.
	t.Setenv("STUB_EVENTS_STDERR", "Failed to read event: input/output error")
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	code, _ := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
}

func TestReclaimIPAM_thresholdBoundary(t *testing.T) {
	// 85% of a /24 is 215 leases.
	for _, tc := range []struct {
		name          string
		leases        int
		expectCleared bool
	}{
		{"below", 214, false},
		{"at", 215, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stubPath(t)
			setPodman(t, "exited", 0, nil, 0)
			db := newIPAMDB(t, tc.leases)
			t.Setenv("IPAM_DB", db)

			if code, _ := runReclaimer(t); code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			_, err := os.Stat(db)
			cleared := os.IsNotExist(err)
			if cleared != tc.expectCleared {
				t.Errorf("with %d leases: cleared=%v, want %v", tc.leases, cleared, tc.expectCleared)
			}
		})
	}
}

func TestReclaimIPAM_recentCreateEvent(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, []string{"abc123"}, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
}

func TestReclaimIPAM_rechecksBeforeDelete(t *testing.T) {
	stubPath(t)
	calls := setPodman(t, "exited", 0, []string{"", "abc123"}, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if got := countEvents(t, calls); got != 2 {
		t.Errorf("expected the idle check to run twice, got %d", got)
	}
}

func TestReclaimIPAM_idleClearsDB(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Error("expected the lease db to be cleared")
	}
	if !strings.Contains(logged, "cleared") {
		t.Errorf("expected a clearing log line, got %q", logged)
	}
	if !strings.Contains(logged, "250") {
		t.Errorf("expected the lease count in the log line, got %q", logged)
	}
}

func TestReclaimIPAM_podmanErrorSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "", 1, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
}

func TestReclaimIPAM_eventsErrorSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 1)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
}

func TestReclaimIPAM_dbChangedDuringChecks(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)
	// Model a job starting mid-check: the stub records a lease in the db while
	// the reclaimer is asking podman for events.
	t.Setenv("STUB_EVENTS_MUTATE_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "db changed") {
		t.Errorf("expected a log line about the db changing, got %q", logged)
	}
}

func TestReclaimIPAM_unparsableLeaseCount(t *testing.T) {
	stubPath(t)
	setPodman(t, "", 0, nil, 0)
	db := writeDB(t, "no leases in here\x00")
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "no leases") {
		t.Errorf("expected a diagnostic log line when the count is unreadable, got %q", logged)
	}
}

// holdNetavarkLock takes the same fcntl lock the reclaimer uses, the way podman
// does, and returns a release function.
func holdNetavarkLock(t *testing.T, path string) func() {
	t.Helper()
	cmd := exec.Command("python3", "-c", `
import fcntl, os, sys, time
fd = os.open(sys.argv[1], os.O_RDWR | os.O_CREAT, 0o600)
fcntl.lockf(fd, fcntl.LOCK_EX)
print("locked", flush=True)
time.sleep(60)
`, path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("lock helper pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start lock helper: %v", err)
	}
	buf := make([]byte, 8)
	if _, err := out.Read(buf); err != nil {
		t.Fatalf("lock helper never locked: %v", err)
	}
	return func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

func TestReclaimIPAM_lockContentionSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)
	lock := filepath.Join(t.TempDir(), "netavark.lock")
	t.Setenv("NETAVARK_LOCK", lock)
	t.Setenv("LOCK_TIMEOUT", "1")

	release := holdNetavarkLock(t, lock)
	defer release()

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "could not take") {
		t.Errorf("expected a lock-timeout log line, got %q", logged)
	}
}

func TestReclaimIPAM_fingerprintFailureSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)
	// python3 is required for the fingerprint and the lock.
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, "python3"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write broken python3: %v", err)
	}
	t.Setenv("PATH", broken+string(os.PathListSeparator)+os.Getenv("PATH"))

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "cannot fingerprint") {
		t.Errorf("expected a fingerprint failure log line, got %q", logged)
	}
}

func TestReclaimIPAM_unreadableDB(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("file permissions do not restrict root; cannot make the db unreadable")
	}
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	if err := os.Chmod(db, 0o000); err != nil {
		t.Fatalf("chmod db: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(db, 0o600) })
	t.Setenv("IPAM_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "cannot read") {
		t.Errorf("expected a read failure log line, got %q", logged)
	}
}

func TestReclaimIPAM_capacityAcrossNetworks(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	// The /24 lives on a second network; the threshold has to come from it.
	t.Setenv("STUB_NETWORK_NAMES", "podman extra")
	t.Setenv("STUB_NETWORKS_PODMAN", "10.88.0.0/16|")
	t.Setenv("STUB_NETWORKS_EXTRA", "192.168.100.0/24|")
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	if code, _ := runReclaimer(t); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Error("expected the /24 network to set the threshold and trigger reclamation")
	}
}

func TestReclaimIPAM_leaseRangeCapacity(t *testing.T) {
	// A 21-address range inside a /16 exhausts at 21, so 85% is 17.
	for _, tc := range []struct {
		leases        int
		expectCleared bool
	}{
		{16, false},
		{17, true},
	} {
		t.Run(strconv.Itoa(tc.leases), func(t *testing.T) {
			stubPath(t)
			setPodman(t, "exited", 0, nil, 0)
			t.Setenv("STUB_NETWORKS", "10.88.0.0/16|10.88.0.10-10.88.0.30")
			t.Setenv("STUB_NETWORK_NAMES", "podman")
			db := newIPAMDB(t, tc.leases)
			t.Setenv("IPAM_DB", db)

			if code, _ := runReclaimer(t); code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			_, err := os.Stat(db)
			if cleared := os.IsNotExist(err); cleared != tc.expectCleared {
				t.Errorf("with %d lease strings: cleared=%v, want %v", tc.leases, cleared, tc.expectCleared)
			}
		})
	}
}

func TestReclaimIPAM_partialLeaseReadSkips(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)

	// grep emits matches and then fails, the way a truncated read can.
	broken := t.TempDir()
	grepStub := `#!/bin/sh
if [ "$1" = "-aoE" ]; then
  printf '192.168.100.1\n192.168.100.2\n'
  exit 2
fi
exec /usr/bin/grep "$@"
`
	if err := os.WriteFile(filepath.Join(broken, "grep"), []byte(grepStub), 0o755); err != nil {
		t.Fatalf("write grep stub: %v", err)
	}
	t.Setenv("PATH", broken+string(os.PathListSeparator)+os.Getenv("PATH"))

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	assertDBExists(t, db)
	if !strings.Contains(logged, "cannot read") {
		t.Errorf("expected a read failure log line, got %q", logged)
	}
}

func TestReclaimIPAM_dbVanishesBeforeDelete(t *testing.T) {
	stubPath(t)
	setPodman(t, "exited", 0, nil, 0)
	db := newIPAMDB(t, 250)
	t.Setenv("IPAM_DB", db)
	// The final check runs under the lock and can no longer read the db.
	t.Setenv("STUB_EVENTS_DELETE_DB", db)

	code, logged := runReclaimer(t)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(logged, "could not read or remove") {
		t.Errorf("expected a locked-check failure log line, got %q", logged)
	}
}
