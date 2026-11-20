package params

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
)

func TestMACArchAsRunnerArch(t *testing.T) {
	cases := []struct {
		mac    string
		github github.Arch
		gitlab gitlab.Arch
	}{
		{"x86", github.Amd64, gitlab.Amd64},
		{"m1", github.Arm64, gitlab.Arm64},
		{"m2", github.Arm64, gitlab.Arm64},
	}
	for _, tc := range cases {
		if got := *MACArchAsGithubArch(tc.mac); got != tc.github {
			t.Errorf("MACArchAsGithubArch(%q) = %q; want %q", tc.mac, got, tc.github)
		}
		if got := *MACArchAsGitLabArch(tc.mac); got != tc.gitlab {
			t.Errorf("MACArchAsGitLabArch(%q) = %q; want %q", tc.mac, got, tc.gitlab)
		}
	}
}

func TestGithubRunnerArgsKeepsPlatform(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set(ghActionsRunnerToken, "test-token")
	viper.Set(ghActionsRunnerRepo, "https://github.com/example/repo")

	for _, platform := range []github.Platform{github.Linux, github.Windows, github.Darwin} {
		args := GithubRunnerArgs(&platform, &github.Amd64)
		if args == nil {
			t.Fatalf("GithubRunnerArgs(%q) returned nil", platform)
		}
		if *args.Platform != platform {
			t.Errorf("GithubRunnerArgs(%q).Platform = %q", platform, *args.Platform)
		}
	}
}

func TestGitLabRunnerArgsKeepsPlatform(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set(glRunnerToken, "glpat-test")
	viper.Set(glRunnerProjectID, "123")

	for _, platform := range []gitlab.Platform{gitlab.Linux, gitlab.Windows, gitlab.Darwin} {
		args := GitLabRunnerArgs(&platform, &gitlab.Amd64)
		if args == nil {
			t.Fatalf("GitLabRunnerArgs(%q) returned nil", platform)
		}
		if *args.Platform != platform {
			t.Errorf("GitLabRunnerArgs(%q).Platform = %q", platform, *args.Platform)
		}
	}
}

func writeJITConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jit.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGithubRunnerArgsFromJITConfig(t *testing.T) {
	t.Cleanup(viper.Reset)
	t.Setenv("GITHUB_TOKEN", "")
	viper.Set(ghActionsRunnerJITConfigFile, writeJITConfig(t, "  jit-config-value\n"))

	args := GithubRunnerArgs(&github.Linux, &github.Amd64)
	if args == nil {
		t.Fatal("GithubRunnerArgs returned nil")
	}
	if args.JITConfig != "jit-config-value" {
		t.Errorf("JITConfig = %q; want the trimmed file content", args.JITConfig)
	}
	if args.Token != "" {
		t.Errorf("Token = %q; a just-in-time runner needs none", args.Token)
	}
}

func TestReadJITConfigRejects(t *testing.T) {
	cases := []struct {
		name    string
		set     func()
		content string
	}{
		{"token", func() { viper.Set(ghActionsRunnerToken, "t") }, "jit"},
		{"app id", func() { viper.Set(ghActionsAppID, "1") }, "jit"},
		{"app installation id", func() { viper.Set(ghActionsAppInstallationID, "2") }, "jit"},
		{"app private key", func() { viper.Set(ghActionsAppPrivateKey, "key.pem") }, "jit"},
		{"labels", func() { viper.Set(ghActionsRunnerLabels, []string{"gpu"}) }, "jit"},
		{"ephemeral", func() { viper.Set(ghActionsRunnerEphemeral, true) }, "jit"},
		{"empty file", func() {}, " \n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			tc.set()
			if _, err := readJITConfig(writeJITConfig(t, tc.content)); err == nil {
				t.Error("readJITConfig succeeded; want an error")
			}
		})
	}
	t.Run("missing file", func(t *testing.T) {
		t.Cleanup(viper.Reset)
		if _, err := readJITConfig(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Error("readJITConfig succeeded; want an error")
		}
	})
}
