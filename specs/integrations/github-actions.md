# Integration: GitHub Actions Self-Hosted Runner

**Package:** `github.com/redhat-developer/mapt/pkg/integrations/github`

Registers the provisioned machine as a GitHub Actions self-hosted runner at boot.
The runner binary is downloaded and installed by the injected setup script.

See `specs/integrations/overview.md` for the shared interface and config flow.

---

## Type

```go
type GithubRunnerArgs struct {
    Token    string     // GitHub runner registration token (required)
    RepoURL  string     // Repository or organisation URL to register against (required)
    Name     string     // Runner name — set to mCtx.RunID() by the action
    Platform *Platform  // Target OS: Linux | Darwin | Windows
    Arch     *Arch      // Target arch: Amd64 | Arm64 | Arm
    Labels   []string   // Runner labels, comma-joined before injection
    User     string     // OS user to run as (set by cloud-config builder)
}
```

### Platform / Arch constants

```go
var (
    Windows Platform = "win"
    Linux   Platform = "linux"
    Darwin  Platform = "osx"

    Arm64 Arch = "arm64"
    Amd64 Arch = "x64"
    Arm   Arch = "arm"
)
```

---

## Runner Version

```go
var runnerVersion = "2.317.0"  // overridden at build time via linker flag
```

Makefile variable: `GITHUB_RUNNER`
Linker target: `pkg/integrations/github.runnerVersion`

---

## Download URL Pattern

```
https://github.com/actions/runner/releases/download/v{version}/actions-runner-{platform}-{arch}-{version}.tar.gz
https://github.com/actions/runner/releases/download/v{version}/actions-runner-{platform}-{arch}-{version}.zip  (Windows)
```

The URL is built by `downloadURL()` and injected as `UserDataValues.CliURL`.

---

## Functions

```go
func Init(args *GithubRunnerArgs)        // stores args as package-level state
func GetRunnerArgs() *GithubRunnerArgs   // returns nil if not configured
func GetToken() string                   // returns token or "" if not configured
```

`GetRunnerArgs()` implements `IntegrationConfig` (via pointer receiver methods on
`*GithubRunnerArgs`) — pass directly to `GetIntegrationSnippet`.

---

## UserDataValues populated

| Field | Source |
|---|---|
| `CliURL` | `downloadURL()` — version + platform + arch |
| `Name` | `GithubRunnerArgs.Name` |
| `Token` | `GithubRunnerArgs.Token` |
| `Labels` | `GithubRunnerArgs.Labels` joined with `,` |
| `RepoURL` | `GithubRunnerArgs.RepoURL` |
| `User` | Set by `GetIntegrationSnippet` from `username` arg |
| `Port`, `Executor` | Not used |

---

## Script Templates

Embedded at compile time:
- `snippet-linux.sh` — downloads `.tar.gz`, extracts, configures, starts as systemd service
- `snippet-darwin.sh` — same flow for macOS
- `snippet-windows.ps1` — downloads `.zip`, extracts, registers as Windows service

Template selection is based on `GithubRunnerArgs.Platform`.
