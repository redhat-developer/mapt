# Integration: GitLab Runner

**Package:** `github.com/redhat-developer/mapt/pkg/integrations/gitlab`

Registers the provisioned machine as a GitLab runner. Unlike GitHub Actions and Cirrus CI,
GitLab registration requires creating a runner resource in GitLab itself to obtain an auth
token. mapt uses the Pulumi GitLab provider to create the runner as a Pulumi resource
inside the deploy stack — the auth token is resolved at provision time and injected into
the setup script.

See `specs/integrations/overview.md` for the shared interface and config flow.

---

## Type

```go
type GitLabRunnerArgs struct {
    GitLabPAT string    // Personal Access Token for the Pulumi GitLab provider
    ProjectID string    // GitLab project ID — mutually exclusive with GroupID
    GroupID   string    // GitLab group ID — mutually exclusive with ProjectID
    URL       string    // GitLab instance URL (e.g. "https://gitlab.com")
    Tags      []string  // Runner tags for job routing; empty = accepts untagged jobs
    Name      string    // Runner description — set to mCtx.RunID() by the action
    Platform  *Platform // Target OS: Linux | Darwin | Windows
    Arch      *Arch     // Target arch: Amd64 | Arm64 | Arm
    User      string    // OS user to run as
    AuthToken string    // Set by Pulumi after CreateRunner(); not caller-supplied
}
```

### Platform / Arch constants

```go
var (
    Windows Platform = "windows"
    Linux   Platform = "linux"
    Darwin  Platform = "darwin"

    Arm64 Arch = "arm64"
    Amd64 Arch = "amd64"
    Arm   Arch = "arm"
)
```

---

## Runner Version

```go
var version = "18.8.0"  // overridden at build time via linker flag
```

Makefile variable: `GITLAB_RUNNER`
Linker target: `pkg/integrations/gitlab.version`

---

## Download URL Pattern

```
https://gitlab-runner-downloads.s3.amazonaws.com/v{version}/binaries/gitlab-runner-{platform}-{arch}
https://gitlab-runner-downloads.s3.amazonaws.com/v{version}/binaries/gitlab-runner-{platform}-{arch}.exe  (Windows)
```

---

## Pulumi Registration (key difference from other integrations)

GitLab runners must be registered in GitLab before deployment. mapt handles this inside the
Pulumi deploy stack by calling `CreateRunner()`:

```go
func CreateRunner(ctx *pulumi.Context, args *GitLabRunnerArgs) (pulumi.StringOutput, error)
```

This creates a `gitlab.UserRunner` Pulumi resource via the `pulumi-gitlab` provider,
authenticated with `GitLabPAT`. The resource returns an `AuthToken` as a `pulumi.StringOutput`.

The returned token is then wired via `ApplyT` into the userdata generation so it is available
when the cloud-init script is rendered:

```go
token, err := gitlab.CreateRunner(ctx, glArgs)
// token is a pulumi.StringOutput resolved during stack apply
token.ApplyT(func(t string) string {
    gitlab.SetAuthToken(t)
    // generate userdata here using GetIntegrationSnippet
    return t
})
```

Exports added to the stack: `gitlab-runner-id`, `gitlab-runner-type`.

### Project vs Group runner

Exactly one of `ProjectID` or `GroupID` must be set — `CreateRunner` returns an error if
both or neither are provided:

| Field set | Runner type | GitLab API |
|---|---|---|
| `ProjectID` | `project_type` | Scoped to a single project |
| `GroupID` | `group_type` | Shared across all projects in the group |

---

## Functions

```go
func Init(args *GitLabRunnerArgs)           // stores args as package-level state
func GetRunnerArgs() *GitLabRunnerArgs      // returns nil if not configured
func GetToken() string                      // returns AuthToken or "" if not configured
func SetAuthToken(token string)             // called inside ApplyT after CreateRunner
func CreateRunner(ctx *pulumi.Context, args *GitLabRunnerArgs) (pulumi.StringOutput, error)
```

---

## UserDataValues populated

| Field | Source |
|---|---|
| `CliURL` | `downloadURL()` — version + platform + arch |
| `Name` | `GitLabRunnerArgs.Name` |
| `Token` | `GitLabRunnerArgs.AuthToken` — set by Pulumi, not caller |
| `RepoURL` | `GitLabRunnerArgs.URL` |
| `User` | Set by `GetIntegrationSnippet` from `username` arg |
| `Labels`, `Port`, `Executor` | Not used |

---

## Script Templates

Embedded at compile time:
- `snippet-linux.sh` — downloads binary, registers runner, starts as systemd service
- `snippet-darwin.sh` — same flow for macOS
- `snippet-windows.ps1` — downloads `.exe`, installs as Windows service

Template selection is based on `GitLabRunnerArgs.Platform`.

---

## Known Gaps

- No Tekton task template includes the GitLab runner flags (verify and add)
- Tags are not surfaced in the setup script — only the Pulumi resource carries them
