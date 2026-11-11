package gitlab

import (
	"strings"
	"testing"
)

// TestSnippetMatchesPlatform checks that each platform gets its own setup
// script and runner download, so Windows and macOS hosts don't run the Linux
// script.
func TestSnippetMatchesPlatform(t *testing.T) {
	cases := []struct {
		name     string
		platform Platform
		arch     Arch
		want     []string
		notWant  []string
	}{
		{"windows amd64", Windows, Amd64,
			[]string{"C:\\GitLab-Runner\\gitlab-runner.exe register", "gitlab-runner-windows-amd64.exe"},
			[]string{"sudo gitlab-runner", "podman.socket"}},
		{"macOS arm64", Darwin, Arm64,
			[]string{"sudo gitlab-runner register", "gitlab-runner-darwin-arm64"},
			[]string{"podman.socket", "gitlab-runner.exe"}},
		{"macOS amd64", Darwin, Amd64,
			[]string{"gitlab-runner-darwin-amd64"},
			[]string{"podman.socket", "gitlab-runner.exe"}},
		{"linux amd64", Linux, Amd64,
			[]string{"podman.socket", "gitlab-runner-linux-amd64"},
			[]string{"gitlab-runner.exe", "LaunchDaemon"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := newRunnerArgs("", tc.arch)
			platform := tc.platform
			args.Platform = &platform
			snippet := renderSnippet(t, args)
			for _, want := range tc.want {
				if !strings.Contains(snippet, want) {
					t.Errorf("snippet missing %q", want)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(snippet, notWant) {
					t.Errorf("snippet must not contain %q", notWant)
				}
			}
		})
	}
}
