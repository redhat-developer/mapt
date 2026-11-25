package windowsserver

import (
	"strings"
	"testing"

	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

func render(t *testing.T, udv userDataValues) string {
	t.Helper()
	userdata, err := file.Template(udv, string(BootstrapScript))
	if err != nil {
		t.Fatalf("failed to render the userdata: %v", err)
	}
	return userdata
}

func snippet(t *testing.T, cfg integrations.IntegrationConfig) string {
	t.Helper()
	s, err := integrations.GetIntegrationSnippet(cfg, "testuser")
	if err != nil {
		t.Fatalf("failed to render the snippet: %v", err)
	}
	return *s
}

func githubUserdata(t *testing.T, token, jitConfig string) string {
	t.Helper()
	platform, arch := github.Windows, github.Amd64
	args := &github.GithubRunnerArgs{
		Token:     token,
		RepoURL:   "https://github.com/o/r",
		Name:      "test-runner",
		Platform:  &platform,
		Arch:      &arch,
		JITConfig: jitConfig,
	}
	github.Init(args)
	t.Cleanup(func() { github.Init(nil) })
	return render(t, userDataValues{ActionsRunnerSnippet: snippet(t, args), RunnerToken: github.GetToken()})
}

func gitlabUserdata(t *testing.T, token string) string {
	t.Helper()
	platform, arch := gitlab.Windows, gitlab.Amd64
	args := &gitlab.GitLabRunnerArgs{
		URL:       "https://gitlab.example.com",
		Name:      "test-runner",
		ProjectID: "123",
		Platform:  &platform,
		Arch:      &arch,
	}
	gitlab.Init(args)
	gitlab.SetAuthToken(token)
	t.Cleanup(func() {
		gitlab.SetAuthToken("")
		gitlab.Init(nil)
	})
	return render(t, userDataValues{GitLabSnippet: snippet(t, args), GitLabToken: token})
}

// TestRunnerTokensAreQuotedOnTheirOwnLine checks that each token is a quoted
// PowerShell string and that the runner snippet starts on the next line.
func TestRunnerTokensAreQuotedOnTheirOwnLine(t *testing.T) {
	cases := []struct {
		name     string
		userdata string
		want     string
	}{
		{"github", githubUserdata(t, "AABBCCDDEEFF0011", ""),
			"\n$ghToken = 'AABBCCDDEEFF0011'\nNew-Item -Path C:\\actions-runner -Type Directory"},
		{"github jit config", githubUserdata(t, "", "jit-config-value"),
			"\n$ghToken = ''\nNew-Item -Path C:\\actions-runner -Type Directory"},
		{"gitlab", gitlabUserdata(t, "glrt-test_token"),
			"\n$gitlabToken = 'glrt-test_token'\n# Download GitLab Runner\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.userdata, tc.want) {
				t.Errorf("userdata missing %q:\n%s", tc.want, tc.userdata)
			}
		})
	}
}

func TestNoRunnerSetsNoToken(t *testing.T) {
	userdata := render(t, userDataValues{})
	for _, notWant := range []string{"$ghToken", "$gitlabToken"} {
		if strings.Contains(userdata, notWant) {
			t.Errorf("userdata must not contain %q:\n%s", notWant, userdata)
		}
	}
}
