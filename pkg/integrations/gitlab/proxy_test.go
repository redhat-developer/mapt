package gitlab

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProxyEnv(t *testing.T) {
	env, err := ProxyEnv("http://proxy.example.com:3128", "localhost, .internal.example.com,10.0.0.5,::1,*")
	if err != nil {
		t.Fatalf("ProxyEnv: %v", err)
	}
	want := []string{
		"HTTP_PROXY=http://proxy.example.com:3128",
		"http_proxy=http://proxy.example.com:3128",
		"HTTPS_PROXY=http://proxy.example.com:3128",
		"https_proxy=http://proxy.example.com:3128",
		"NO_PROXY=localhost,.internal.example.com,10.0.0.5,::1,*",
		"no_proxy=localhost,.internal.example.com,10.0.0.5,::1,*",
	}
	if !reflect.DeepEqual(env, want) {
		t.Errorf("ProxyEnv =\n%q\nwant\n%q", env, want)
	}
}

func TestProxyEnvWithoutNoProxy(t *testing.T) {
	env, err := ProxyEnv("https://proxy.example.com", "")
	if err != nil {
		t.Fatalf("ProxyEnv: %v", err)
	}
	if len(env) != 4 || strings.Contains(strings.Join(env, " "), "NO_PROXY") {
		t.Errorf("ProxyEnv = %q; want only the four proxy variables", env)
	}
}

func TestProxyEnvUnset(t *testing.T) {
	env, err := ProxyEnv("", "")
	if err != nil || env != nil {
		t.Errorf("ProxyEnv(\"\", \"\") = %q, %v; want nil, nil", env, err)
	}
}

func TestProxyEnvErrors(t *testing.T) {
	cases := map[string]struct {
		httpProxy, noProxy, wantErr string
	}{
		"CIDR in no-proxy":       {"http://proxy:3128", "localhost,192.168.0.0/16", "192.168.0.0/16\" is a CIDR range"},
		"IPv6 CIDR in no-proxy":  {"http://proxy:3128", "fd00::/8", "is a CIDR range"},
		"invalid no-proxy entry": {"http://proxy:3128", "bad host", "is not a host name"},
		"no-proxy without proxy": {"", "localhost", "needs --glrunner-http-proxy"},
		"proxy without scheme":   {"proxy.example.com:3128", "", "--glrunner-http-proxy"},
		"proxy with bad scheme":  {"ftp://proxy.example.com", "", "must use http or https"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ProxyEnv(tc.httpProxy, tc.noProxy)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("ProxyEnv(%q, %q) error = %v; want it to contain %q", tc.httpProxy, tc.noProxy, err, tc.wantErr)
			}
		})
	}
}

// registerArgs runs the rendered `gitlab-runner register` command with a fake
// gitlab-runner and returns the arguments it received, one per line.
func registerArgs(t *testing.T, snippet string) []string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	start := strings.Index(snippet, "sudo gitlab-runner register")
	if start < 0 {
		t.Fatal("snippet has no gitlab-runner register command")
	}
	var lines []string
	for _, line := range strings.Split(snippet[start:], "\n") {
		lines = append(lines, line)
		if !strings.HasSuffix(strings.TrimSpace(line), `\`) {
			break
		}
	}
	register := strings.Replace(strings.Join(lines, "\n"), "sudo gitlab-runner", "fake_runner", 1)
	script := "fake_runner() { printf '%s\\n' \"$@\"; }\n_docker_dns_args=()\n" + register + "\n"
	path := filepath.Join(t.TempDir(), "register.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("write script: %v", err)
	}
	out, err := exec.Command("bash", path).CombinedOutput()
	if err != nil {
		t.Fatalf("register command failed: %v\n%s", err, out)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

func envArgs(args []string) []string {
	var env []string
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--env" {
			env = append(env, args[i+1])
		}
	}
	return env
}

func TestSnippetRegistersProxyEnv(t *testing.T) {
	// A proxy password may contain shell characters; they must reach the runner unchanged.
	proxy := `http://user:p'a$s@proxy.example.com:3128`
	env, err := ProxyEnv(proxy, "localhost,.example.com")
	if err != nil {
		t.Fatalf("ProxyEnv: %v", err)
	}

	for _, platform := range []Platform{Linux, Darwin} {
		t.Run(string(platform), func(t *testing.T) {
			args := newRunnerArgs("", Amd64)
			args.Platform = &platform
			args.Env = env
			got := envArgs(registerArgs(t, renderSnippet(t, args)))
			if !reflect.DeepEqual(got, env) {
				t.Errorf("register --env values =\n%q\nwant\n%q", got, env)
			}
		})
	}
}

func TestSnippetLinuxPowerRegistersProxyEnv(t *testing.T) {
	env, err := ProxyEnv("http://proxy.example.com:3128", "localhost")
	if err != nil {
		t.Fatalf("ProxyEnv: %v", err)
	}
	args := newPowerArgs("us-south", testResolvers, 1450)
	args.Env = env
	registered := registerArgs(t, renderSnippet(t, args))
	if got := envArgs(registered); !reflect.DeepEqual(got, env) {
		t.Errorf("register --env values =\n%q\nwant\n%q", got, env)
	}
	if !strings.Contains(strings.Join(registered, "\n"), "--docker-volumes") {
		t.Errorf("the arguments after --env were lost: %q", registered)
	}
}

func TestSnippetWithoutProxyHasNoEnv(t *testing.T) {
	for _, platform := range []Platform{Linux, Darwin, Windows} {
		args := newRunnerArgs("", Amd64)
		args.Platform = &platform
		if snippet := renderSnippet(t, args); strings.Contains(snippet, "--env") {
			t.Errorf("%s snippet without a proxy must not pass --env", platform)
		}
	}
}

func TestSnippetWindowsQuotesProxyEnv(t *testing.T) {
	windows := Windows
	args := newRunnerArgs("", Amd64)
	args.Platform = &windows
	args.Env = []string{"HTTP_PROXY=http://us'er@proxy:3128"}
	snippet := renderSnippet(t, args)
	want := "  --env 'HTTP_PROXY=http://us''er@proxy:3128' `\n  --executor \"shell\""
	if !strings.Contains(snippet, want) {
		t.Errorf("Windows snippet missing %q:\n%s", want, snippet)
	}
}
