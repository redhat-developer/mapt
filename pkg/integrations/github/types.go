package github

type Platform string
type Arch string

var (
	Windows Platform = "win"
	Linux   Platform = "linux"
	Darwin  Platform = "osx"

	Arm64 Arch = "arm64"
	Amd64 Arch = "x64"
	Arm   Arch = "arm"
)

type GithubRunnerArgs struct {
	Token    string
	RepoURL  string
	Name     string
	Platform *Platform
	Arch     *Arch
	Labels   []string
	User     string
}
