package gitlab

type Platform string
type Arch string

var (
	Windows Platform = "windows"
	Linux   Platform = "linux"
	Darwin  Platform = "darwin"

	Arm64 Arch = "arm64"
	Amd64 Arch = "amd64"
	Arm   Arch = "arm"
)

type GitLabRunnerArgs struct {
	GitLabPAT string   // Personal Access Token for Pulumi GitLab provider (input)
	ProjectID string   // GitLab project ID for project runner registration (mutually exclusive with GroupID)
	GroupID   string   // GitLab group ID for group runner registration (mutually exclusive with ProjectID)
	URL       string   // GitLab instance URL (e.g., https://gitlab.com)
	Tags      []string // Runner tags for job routing (e.g., ["linux", "aws", "spot"])
	Name      string   // Runner name/description (auto-generated from run ID)
	Platform  *Platform // Target platform
	Arch      *Arch     // Target architecture
	User      string    // OS user to run as
	AuthToken string    // Runner authentication token (set by Pulumi during deployment)
}
