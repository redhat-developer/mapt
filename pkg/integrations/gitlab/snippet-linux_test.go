package gitlab

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations"
)

// renderSnippet renders the Linux setup script for the given runner args.
//
// GetSetupScriptTemplate reads the package-global runnerArgs set by Init,
// so tests must initialize the global with the args they render.
func renderSnippet(t *testing.T, args *GitLabRunnerArgs) string {
	t.Helper()
	Init(args)
	snippet, err := integrations.GetIntegrationSnippet(args, "testuser")
	if err != nil {
		t.Fatalf("failed to render snippet: %v", err)
	}
	return *snippet
}

func newRunnerArgs(region string, arch Arch) *GitLabRunnerArgs {
	platform := Linux
	return &GitLabRunnerArgs{
		URL:           "https://gitlab.example.com",
		Name:          "test-runner",
		ProjectID:     "123",
		GitLabToken:   "glpat-test",
		Platform:      &platform,
		Arch:          &arch,
		PowerVSRegion: region,
	}
}

// regionResolvers is the IBM Cloud per-region DNS resolver table from
// https://github.com/redhat-developer/mapt/issues/894 (the issue's "lon"
// rows are the IBM eu-gb region).
var regionResolvers = map[string][]string{
	"us-south": {"10.130.98.132", "10.130.98.5", "10.130.98.69"},
	"us-east":  {"10.130.64.4", "10.130.64.70", "10.130.64.134"},
	"eu-de":    {"10.130.128.5", "10.130.128.69", "10.130.128.133"},
	"eu-gb":    {"10.130.129.4", "10.130.129.68", "10.130.129.132"},
}

// assertPowerNetworking checks the Power-only networking settings from
// mapt issue #894 (config validated on live dal14 ppc64le runners):
// non-conflicting nested subnet, shared host network, public DNS fallback,
// and reduced DNS retry count.
func assertPowerNetworking(t *testing.T, snippet string) {
	t.Helper()
	for _, want := range []string{
		`_pow_subnet="172.20.0.0/24"`,    // non-conflicting nested subnet
		`--docker-network-mode "podman"`, // share the host's default podman network
		"attempts:3",                     // fewer DNS retries
		"_pow_apply_dns_fallback",        // public fallback kept within glibc's 3-nameserver limit
		"_pow_check_subnet_overlap",      // CIDR-aware host overlap guard
		"_pow_set_executor_subnet",       // merge [network] rather than appending a second table
		"podman network create podman",   // network the runner registers against must exist
	} {
		if !strings.Contains(snippet, want) {
			t.Errorf("Power snippet missing %q", want)
		}
	}
	if strings.Contains(snippet, "attempts:5") {
		t.Errorf("Power snippet must not use attempts:5")
	}
	if strings.Contains(snippet, `default_subnet = "192.168.100.0/24"`) {
		t.Errorf("Power snippet must not use the conflicting 192.168.100.0/24 subnet")
	}
}

// TestPowerExecutorSubnetMergesInPlace exercises the [network] merge against
// the containers.conf shapes a target image can present. Appending a second
// [network] table is a TOML parse error, so an existing table must be updated
// rather than duplicated.
func TestPowerExecutorSubnetMergesInPlace(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	snippet := renderSnippet(t, newRunnerArgs("us-south", Ppc64le))
	fn := shellFunction(t, snippet, "_pow_set_executor_subnet")
	// The function targets the real path and uses sudo; run the same logic
	// against a temporary file with sudo stripped.
	fn = strings.ReplaceAll(fn, "/etc/containers/executor-containers.conf", "$CONF")
	fn = strings.ReplaceAll(fn, "sudo ", "")

	cases := []struct {
		name  string
		input string
	}{
		{"no network table", "[containers]\npids_limit = -1\n"},
		{"network table with default_subnet", "[containers]\n\n[network]\ndefault_subnet = \"10.88.0.0/16\"\nfirewall_driver = \"none\"\n"},
		{"network table without default_subnet", "[network]\nfirewall_driver = \"none\"\n"},
		{"indented default_subnet", "[network]\n  default_subnet = \"10.88.0.0/16\"\n"},
		{"indented network header", "  [network]\n  firewall_driver = \"none\"\n"},
		{"default_subnet_pools only", "[network]\ndefault_subnet_pools = [{\"base\" = \"10.90.0.0/15\", \"size\" = 24}]\n"},
		// A later table with the same key name must not be rewritten.
		{"table after network", "[network]\ndefault_subnet = \"10.88.0.0/16\"\n\n[other]\ndefault_subnet = \"untouched\"\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conf := filepath.Join(t.TempDir(), "executor-containers.conf")
			if err := os.WriteFile(conf, []byte(tc.input), 0o600); err != nil {
				t.Fatalf("write conf: %v", err)
			}
			script := "set -euo pipefail\nCONF=\"" + conf + "\"\n" + fn + "\n_pow_set_executor_subnet\n"
			if out, err := exec.Command("bash", "-c", script).CombinedOutput(); err != nil {
				t.Fatalf("running merge: %v\n%s", err, out)
			}
			got, err := os.ReadFile(conf)
			if err != nil {
				t.Fatalf("read conf: %v", err)
			}
			result := string(got)
			if n := strings.Count(result, "[network]"); n != 1 {
				t.Errorf("expected exactly one [network] table, found %d:\n%s", n, result)
			}
			if !strings.Contains(result, `default_subnet = "172.20.0.0/24"`) {
				t.Errorf("expected the nested subnet to be set:\n%s", result)
			}
			if strings.Contains(result, `default_subnet = "10.88.0.0/16"`) {
				t.Errorf("expected the previous default_subnet to be replaced:\n%s", result)
			}
			if n := len(activeDefaultSubnet.FindAllString(networkTable(result), -1)); n != 1 {
				t.Errorf("expected exactly one default_subnet key in [network], found %d:\n%s", n, result)
			}
			if strings.Contains(tc.input, `"untouched"`) && !strings.Contains(result, `default_subnet = "untouched"`) {
				t.Errorf("expected keys in later tables to survive:\n%s", result)
			}
			if strings.Contains(tc.input, "  default_subnet") && !strings.Contains(result, `  default_subnet = "172.20.0.0/24"`) {
				t.Errorf("expected the replaced key to keep its indentation:\n%s", result)
			}
			if strings.Contains(tc.input, "default_subnet_pools") && !strings.Contains(result, "default_subnet_pools = ") {
				t.Errorf("expected default_subnet_pools to survive:\n%s", result)
			}
			if strings.Contains(tc.input, "firewall_driver") && !strings.Contains(result, `firewall_driver = "none"`) {
				t.Errorf("expected unrelated [network] keys to survive:\n%s", result)
			}
			if strings.Contains(tc.input, "pids_limit") && !strings.Contains(result, "pids_limit = -1") {
				t.Errorf("expected other tables to survive:\n%s", result)
			}
		})
	}
}

// activeDefaultSubnet matches an uncommented default_subnet key, but not
// default_subnet_pools.
var activeDefaultSubnet = regexp.MustCompile(`(?m)^[ \t]*default_subnet[ \t]*=`)

// networkTable returns the body of the [network] table, up to the next table
// header.
func networkTable(conf string) string {
	var body []string
	in := false
	for _, line := range strings.Split(conf, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			in = strings.HasPrefix(trimmed, "[network]")
			continue
		}
		if in {
			body = append(body, line)
		}
	}
	return strings.Join(body, "\n")
}

// shellFunction returns the definition of the named shell function from the
// rendered snippet, from its signature line through its closing brace.
func shellFunction(t *testing.T, snippet, name string) string {
	t.Helper()
	lines := strings.Split(snippet, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == name+"() {" {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("function %s not found in rendered snippet", name)
	}
	for i := start + 1; i < len(lines); i++ {
		// A shell function closes with a brace at column 0; nested awk
		// blocks close with an indented one.
		if lines[i] == "}" {
			return strings.Join(lines[start:i+1], "\n")
		}
	}
	t.Fatalf("function %s has no closing brace", name)
	return ""
}

// extractCaseBlock returns the region-selection case block for the region.
func extractCaseBlock(t *testing.T, snippet, region string) string {
	t.Helper()
	start := strings.Index(snippet, "case \""+region+"\" in")
	if start < 0 {
		t.Fatalf("region case block for %q not found in snippet", region)
	}
	end := strings.Index(snippet[start:], "\nesac\n")
	if end < 0 {
		t.Fatalf("end of region case block for %q not found", region)
	}
	return snippet[start:start+end] + "\nesac"
}

// extractBashFunction returns the definition of the named function from the
// rendered snippet, from "name() {" through the closing brace line.
func extractBashFunction(t *testing.T, snippet, name string) string {
	t.Helper()
	start := strings.Index(snippet, name+"() {")
	if start < 0 {
		t.Fatalf("function %s not found in snippet", name)
	}
	end := strings.Index(snippet[start:], "\n}\n")
	if end < 0 {
		t.Fatalf("end of function %s not found", name)
	}
	return snippet[start : start+end+len("\n}")]
}

// runDNSSelection executes the region-selection case block from the rendered
// snippet and returns the resulting _dns_servers value.
func runDNSSelection(t *testing.T, region, snippet string) string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	script := "set -e\n_dns_servers=\"\"\n" + extractCaseBlock(t, snippet, region) + "\nprintf '%s' \"$_dns_servers\"\n"
	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		t.Fatalf("executing DNS selection for %q: %v", region, err)
	}
	return string(out)
}

func TestSnippetLinuxPowerRegionalDNS(t *testing.T) {
	for region, resolvers := range regionResolvers {
		t.Run(region, func(t *testing.T) {
			snippet := renderSnippet(t, newRunnerArgs(region, Ppc64le))
			want := strings.Join(resolvers, " ")
			if got := runDNSSelection(t, region, snippet); got != want {
				t.Errorf("region %s selected resolvers %q; expected %q", region, got, want)
			}
			for _, ip := range resolvers {
				if !strings.Contains(snippet, ip) {
					t.Errorf("snippet for %s missing resolver %s", region, ip)
				}
			}
			assertPowerNetworking(t, snippet)
		})
	}
}

func TestSnippetLinuxPowerUnlistedRegionFallsBackToDetection(t *testing.T) {
	snippet := renderSnippet(t, newRunnerArgs("not-a-region", Ppc64le))
	// Execute the actual region-selection block: an unlisted region must leave
	// _dns_servers empty so the host detection cascade runs.
	if got := runDNSSelection(t, "not-a-region", snippet); got != "" {
		t.Errorf("unlisted region selected resolvers %q; expected none (host detection fallback)", got)
	}
	// The host detection cascade must remain in the script.
	for _, want := range []string{
		"resolvectl", "nmcli dev show", "/run/systemd/resolve/resolv.conf",
	} {
		if !strings.Contains(snippet, want) {
			t.Errorf("host DNS detection missing %q", want)
		}
	}
	assertPowerNetworking(t, snippet)
}

// TestSnippetLinuxPowerDNSFinalOrder executes the complete selection and
// fallback logic from the rendered script and verifies the final resolver
// order: glibc reads at most three nameservers from resolv.conf, so the
// public fallback must sit within the first three entries.
func TestSnippetLinuxPowerDNSFinalOrder(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	for region, resolvers := range regionResolvers {
		t.Run(region, func(t *testing.T) {
			snippet := renderSnippet(t, newRunnerArgs(region, Ppc64le))
			script := "set -e\n_dns_servers=\"\"\n" +
				extractCaseBlock(t, snippet, region) + "\n" +
				extractBashFunction(t, snippet, "_pow_apply_dns_fallback") + "\n" +
				"_pow_apply_dns_fallback\nprintf '%s' \"$_dns_servers\"\n"
			out, err := exec.Command("bash", "-c", script).Output()
			if err != nil {
				t.Fatalf("executing DNS selection and fallback for %q: %v", region, err)
			}
			want := strings.Join(resolvers[:2], " ") + " 8.8.8.8"
			if got := string(out); got != want {
				t.Errorf("final DNS order = %q; want %q (the public fallback must be within the first three entries)", got, want)
			}
		})
	}
}

// TestSnippetLinuxPowerDNSFallbackInvariants executes the fallback function
// from the rendered script against synthetic resolver lists.
func TestSnippetLinuxPowerDNSFallbackInvariants(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	snippet := renderSnippet(t, newRunnerArgs("us-south", Ppc64le))
	fn := extractBashFunction(t, snippet, "_pow_apply_dns_fallback")
	cases := []struct{ name, in, want string }{
		{"long list keeps two plus fallback", "1.1.1.1 2.2.2.2 3.3.3.3 4.4.4.4", "1.1.1.1 2.2.2.2 8.8.8.8"},
		{"existing fallback deduplicated", "8.8.8.8 1.1.1.1 2.2.2.2", "1.1.1.1 2.2.2.2 8.8.8.8"},
		{"single resolver", "1.1.1.1", "1.1.1.1 8.8.8.8"},
		{"detection fallback only", "8.8.8.8 8.8.4.4", "8.8.4.4 8.8.8.8"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			script := fn + "\n_dns_servers=\"" + tc.in + "\"\n_pow_apply_dns_fallback\nprintf '%s' \"$_dns_servers\"\n"
			out, err := exec.Command("bash", "-c", script).Output()
			if err != nil {
				t.Fatalf("executing fallback: %v", err)
			}
			if got := string(out); got != tc.want {
				t.Errorf("fallback(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestSnippetLinuxPowerSubnetOverlapGuard executes the subnet-overlap guard
// function from the rendered script with a stubbed ip command, covering the
// covering, contained, disjoint, and failed-inspection cases.
func TestSnippetLinuxPowerSubnetOverlapGuard(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	snippet := renderSnippet(t, newRunnerArgs("us-south", Ppc64le))
	fn := extractBashFunction(t, snippet, "_pow_check_subnet_overlap")

	cases := []struct {
		name    string
		routes  string
		addrs   string
		ipFails bool
		accept  bool
		errMsg  string
	}{
		{
			name:   "covering supernet rejected",
			routes: "172.16.0.0/12 dev eth0 proto kernel scope link src 172.16.0.9\ndefault via 10.0.0.1 dev eth0\n",
			addrs:  "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450\n    inet 172.16.0.9/12 brd 172.31.255.255 scope global eth0\n",
			accept: false,
			errMsg: "overlaps",
		},
		{
			name:   "typed blackhole route rejected",
			routes: "blackhole 172.16.0.0/12 proto static metric 100\ndefault via 10.0.0.1 dev eth0\n",
			addrs:  "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>\n    inet 10.0.0.5/24 brd 10.0.0.255 scope global eth0\n",
			accept: false,
			errMsg: "overlaps",
		},
		{
			name:   "typed unreachable route rejected",
			routes: "unreachable 172.20.0.0/24 proto static metric 100\ndefault via 10.0.0.1 dev eth0\n",
			addrs:  "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>\n    inet 10.0.0.5/24 brd 10.0.0.255 scope global eth0\n",
			accept: false,
			errMsg: "overlaps",
		},
		{
			name:   "contained subnet rejected",
			routes: "default via 10.0.0.1 dev eth0\n10.88.0.0/16 dev podman0 proto kernel scope link src 10.88.0.1\n",
			addrs:  "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>\n    inet 172.20.0.128/25 brd 172.20.0.255 scope global eth0\n",
			accept: false,
			errMsg: "overlaps",
		},
		{
			name:   "disjoint networks accepted",
			routes: "default via 10.0.0.1 dev eth0\n172.20.1.0/24 dev eth1 proto kernel scope link src 172.20.1.9\n",
			addrs:  "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>\n    inet 192.168.100.10/24 brd 192.168.100.255 scope global eth0\n",
			accept: true,
		},
		{
			name:    "inspection failure rejected",
			ipFails: true,
			accept:  false,
			errMsg:  "cannot inspect",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubDir := t.TempDir()
			routesFile := filepath.Join(stubDir, "routes")
			addrsFile := filepath.Join(stubDir, "addrs")
			if err := os.WriteFile(routesFile, []byte(tc.routes), 0o600); err != nil {
				t.Fatalf("write routes fixture: %v", err)
			}
			if err := os.WriteFile(addrsFile, []byte(tc.addrs), 0o600); err != nil {
				t.Fatalf("write addrs fixture: %v", err)
			}
			stub := filepath.Join(stubDir, "ip")
			stubScript := `#!/bin/bash
case "$*" in
"-4 route show")
    if [ -n "$STUB_FAIL" ]; then exit 1; fi
    cat "$STUB_ROUTES"
    ;;
"-4 addr show")
    if [ -n "$STUB_FAIL" ]; then exit 1; fi
    cat "$STUB_ADDRS"
    ;;
*) exit 1 ;;
esac
`
			if err := os.WriteFile(stub, []byte(stubScript), 0o755); err != nil {
				t.Fatalf("write ip stub: %v", err)
			}
			script := fn + "\n_pow_check_subnet_overlap\nprintf 'guard-exit=%s' \"$?\"\n"
			cmd := exec.Command("bash", "-c", script)
			cmd.Env = append(os.Environ(),
				"PATH="+stubDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"STUB_ROUTES="+routesFile,
				"STUB_ADDRS="+addrsFile)
			if tc.ipFails {
				cmd.Env = append(cmd.Env, "STUB_FAIL=1")
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("running guard: %v\n%s", err, out)
			}
			got := string(out)
			wantExit := "guard-exit=0"
			if !tc.accept {
				wantExit = "guard-exit=1"
			}
			if !strings.Contains(got, wantExit) {
				t.Errorf("guard output = %q; want %q", got, wantExit)
			}
			if tc.errMsg != "" && !strings.Contains(got, tc.errMsg) {
				t.Errorf("guard output = %q; want error containing %q", got, tc.errMsg)
			}
			if tc.accept && strings.Contains(got, "ERROR") {
				t.Errorf("guard output = %q; accepted case must not report errors", got)
			}
		})
	}
}

func TestSnippetLinuxNonPowerUnchanged(t *testing.T) {
	for _, arch := range []Arch{Amd64, S390x} {
		t.Run(string(arch), func(t *testing.T) {
			snippet := renderSnippet(t, newRunnerArgs("", arch))
			for _, want := range []string{
				`default_subnet = "192.168.100.0/24"`,        // existing nested subnet
				"attempts:5",                                 // existing DNS retry count
				"10.88.0.0/16",                               // existing masquerade rule
				"if command -v resolvectl &>/dev/null; then", // existing detection guard
			} {
				if !strings.Contains(snippet, want) {
					t.Errorf("non-Power snippet missing %q", want)
				}
			}
			for _, notWant := range []string{
				`--docker-network-mode`,
				`default_subnet = "172.20.0.0/24"`,
				"10.130.98.132",
				"attempts:3",
			} {
				if strings.Contains(snippet, notWant) {
					t.Errorf("non-Power snippet must not contain %q", notWant)
				}
			}
		})
	}
}

var updateGolden = flag.Bool("update-golden", false, "update golden files for rendered snippet tests")

// TestSnippetLinuxNonPowerGolden pins the non-Power rendering byte-for-byte:
// the PowerVSRegion plumbing must not change the script emitted for any other
// target (AWS x86_64, IBM Z s390x, macOS hosts).
func TestSnippetLinuxNonPowerGolden(t *testing.T) {
	// downloadURL() bakes the runner version into the script; pin it so the
	// golden does not depend on ldflags (make test injects a different
	// version than the glrunner.go fallback).
	origVersion := version
	version = "golden-version"
	defer func() { version = origVersion }()

	snippet := renderSnippet(t, newRunnerArgs("", Amd64))
	goldenPath := filepath.Join("testdata", "snippet-linux-nonpower.golden")
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(snippet), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden (run with -update-golden to create): %v", err)
	}
	if string(want) != snippet {
		t.Errorf("non-Power snippet differs from golden; verify the diff is intentional, then re-run with -update-golden")
	}
}

// TestSnippetLinuxConfigTomlSettings pins the config.toml shaping in the
// script: the registration flags that determine [runners.docker] content and
// the post-registration sed edits (including the plan requirement to preserve
// existing volume mounts and PID limits). The TOML file itself is written by
// gitlab-runner register at runtime, so its actual content is validated on a
// live runner (final validation for mapt#894), not in unit tests.
func TestSnippetLinuxConfigTomlSettings(t *testing.T) {
	args := newRunnerArgs("us-south", Ppc64le)
	args.Concurrent = 2
	snippet := renderSnippet(t, args)
	for _, want := range []string{
		`--executor "docker"`,
		`--docker-host "unix:///run/podman/podman.sock"`,
		`--docker-privileged`,
		`--docker-network-mode "podman"`,
		`--docker-volumes "/etc/containers/executor-containers.conf:/etc/containers/containers.conf:ro"`,
		`sed -i "s/^concurrent = .*/concurrent = 2/" /etc/gitlab-runner/config.toml`,
		`output_limit = 65536`,
		`pids_limit = -1`,
	} {
		if !strings.Contains(snippet, want) {
			t.Errorf("Power snippet missing config.toml setting %q", want)
		}
	}
}

// TestSnippetLinuxBashSyntax guards the rendered script's bash syntax for
// every configuration (Power and non-Power). It is a regression guard: the
// current script must already pass it.
func TestSnippetLinuxBashSyntax(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	for _, region := range []string{"", "us-south", "eu-gb", "not-a-region"} {
		snippet := renderSnippet(t, newRunnerArgs(region, Ppc64le))
		path := filepath.Join(t.TempDir(), "snippet.sh")
		if err := os.WriteFile(path, []byte(snippet), 0o600); err != nil {
			t.Fatalf("write snippet: %v", err)
		}
		cmd := exec.Command("bash", "-n", path)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("bash -n failed for region %q: %v\n%s", region, err, out)
		}
	}
}
