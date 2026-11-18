package params

import (
	"reflect"
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

func TestGitLabRunnerArgsProxyEnv(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set(glRunnerToken, "glpat-test")
	viper.Set(glRunnerProjectID, "123")
	viper.Set(glRunnerHTTPProxy, "http://proxy.example.com:3128")
	viper.Set(glRunnerNoProxy, "localhost,.example.com")

	linux := gitlab.Linux
	args := GitLabRunnerArgs(&linux, &gitlab.Amd64)
	if args == nil {
		t.Fatal("GitLabRunnerArgs returned nil")
	}
	want := []string{
		"HTTP_PROXY=http://proxy.example.com:3128",
		"http_proxy=http://proxy.example.com:3128",
		"HTTPS_PROXY=http://proxy.example.com:3128",
		"https_proxy=http://proxy.example.com:3128",
		"NO_PROXY=localhost,.example.com",
		"no_proxy=localhost,.example.com",
	}
	if !reflect.DeepEqual(args.Env, want) {
		t.Errorf("Env =\n%q\nwant\n%q", args.Env, want)
	}
}

func TestGitLabRunnerArgsWithoutProxy(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set(glRunnerToken, "glpat-test")
	viper.Set(glRunnerProjectID, "123")

	linux := gitlab.Linux
	args := GitLabRunnerArgs(&linux, &gitlab.Amd64)
	if args == nil || args.Env != nil {
		t.Errorf("GitLabRunnerArgs() = %+v; want args with no Env", args)
	}
}

func TestGitLabRunnerArgsProxyWithoutToken(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set(glRunnerHTTPProxy, "http://proxy.example.com:3128")

	linux := gitlab.Linux
	if args := GitLabRunnerArgs(&linux, &gitlab.Amd64); args != nil {
		t.Errorf("GitLabRunnerArgs() = %+v; want nil without a token", args)
	}
}
