package github

import (
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations"
)

func renderSnippet(t *testing.T, platform Platform, arch Arch) string {
	t.Helper()
	args := &GithubRunnerArgs{
		Token:    "test-token",
		RepoURL:  "https://github.com/example/repo",
		Name:     "test-runner",
		Platform: &platform,
		Arch:     &arch,
	}
	Init(args)
	snippet, err := integrations.GetIntegrationSnippet(args, "testuser")
	if err != nil {
		t.Fatalf("failed to render snippet: %v", err)
	}
	return *snippet
}

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
		{"linux x64", Linux, Amd64,
			[]string{"sudo ./bin/installdependencies.sh", "actions-runner-linux-x64-" + runnerVersion + ".tar.gz"},
			[]string{"config.cmd", "plutil"}},
		{"linux arm64", Linux, Arm64,
			[]string{"sudo ./bin/installdependencies.sh", "actions-runner-linux-arm64-" + runnerVersion + ".tar.gz"},
			[]string{"config.cmd", "plutil"}},
		{"windows x64", Windows, Amd64,
			[]string{"Invoke-WebRequest", "./config.cmd", "actions-runner-win-x64-" + runnerVersion + ".zip"},
			[]string{"installdependencies.sh", "plutil", ".tar.gz"}},
		{"macOS arm64", Darwin, Arm64,
			[]string{"plutil -replace UserName", "actions-runner-osx-arm64-" + runnerVersion + ".tar.gz"},
			[]string{"installdependencies.sh", "config.cmd"}},
		{"macOS x64", Darwin, Amd64,
			[]string{"plutil -replace UserName", "actions-runner-osx-x64-" + runnerVersion + ".tar.gz"},
			[]string{"installdependencies.sh", "config.cmd"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snippet := renderSnippet(t, tc.platform, tc.arch)
			for _, want := range tc.want {
				if !strings.Contains(snippet, want) {
					t.Errorf("snippet missing %q:\n%s", want, snippet)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(snippet, notWant) {
					t.Errorf("snippet must not contain %q:\n%s", notWant, snippet)
				}
			}
		})
	}
}
