package gitlab

type Platform string
type Arch string

var (
	Windows Platform = "windows"
	Linux   Platform = "linux"
	Darwin  Platform = "darwin"

	Arm64   Arch = "arm64"
	Amd64   Arch = "amd64"
	Arm     Arch = "arm"
	Ppc64le Arch = "ppc64le"
	S390x   Arch = "s390x"
)

type GitLabRunnerArgs struct {
	GitLabToken string    // Token with create_runner scope (PAT, group/project access token, or service account token)
	ProjectID   string    // GitLab project ID for project runner registration (mutually exclusive with GroupID)
	GroupID     string    // GitLab group ID for group runner registration (mutually exclusive with ProjectID)
	URL         string    // GitLab instance URL (e.g., https://gitlab.com)
	Tags        []string  // Runner tags for job routing (e.g., ["linux", "aws", "spot"])
	Name        string    // Runner name/description (auto-generated from run ID)
	Platform    *Platform // Target platform
	Arch        *Arch     // Target architecture
	User        string    // OS user to run as (only used when Unsecure is true)
	AuthToken   string    // Runner authentication token (set by Pulumi during deployment)
	Unsecure      bool // When false (default) a dedicated gitlab-runner system user is created; when true the runner service runs as User
	Concurrent    int  // Maximum number of concurrent jobs (written to config.toml; 0 means leave at default of 1)
	LogToJournald bool // When true, sets Podman log_driver=journald so CI job output is captured by systemd journal for OTel correlation
}
