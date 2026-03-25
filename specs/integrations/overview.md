# Integrations: Overview

Integrations allow any provisioned mapt target to register itself as a CI system agent
at boot, without manual setup. The integration is injected as a shell or PowerShell script
into the cloud-init `write_files` section.

Three services are supported — each has its own spec:
- `specs/integrations/github-actions.md` — GitHub Actions self-hosted runner
- `specs/integrations/cirrus-ci.md` — Cirrus CI persistent worker
- `specs/integrations/gitlab.md` — GitLab runner (uses Pulumi for registration)

---

## Shared Interface

**Package:** `github.com/redhat-developer/mapt/pkg/integrations`

### `IntegrationConfig`

```go
type IntegrationConfig interface {
    GetUserDataValues() *UserDataValues   // nil = integration disabled
    GetSetupScriptTemplate() string       // embedded shell/PS1 template string
}
```

Every service implementation implements this interface. Returning `nil` from
`GetUserDataValues()` is the zero-value — it means the integration was not configured
and `GetIntegrationSnippet` returns an empty string.

### `UserDataValues`

```go
type UserDataValues struct {
    CliURL   string   // download URL for the runner binary
    User     string   // OS username — set automatically by GetIntegrationSnippet
    Name     string   // runner/worker name (set to mCtx.RunID())
    Token    string   // registration/auth token
    Labels   string   // comma-separated labels or key=value pairs
    Port     string   // listen port (Cirrus only)
    RepoURL  string   // repository or GitLab instance URL
    Executor string   // executor type (GitLab only)
}
```

Not all fields are used by every service — see the per-service spec for which fields
are populated.

---

## Shared Functions

### `GetIntegrationSnippet`

```go
func GetIntegrationSnippet(intCfg IntegrationConfig, username string) (*string, error)
```

Renders the service's embedded script template with `UserDataValues`. Sets `User` from
`username` before rendering. Returns an empty string (not an error) when
`GetUserDataValues()` returns nil.

### `GetIntegrationSnippetAsCloudInitWritableFile`

```go
func GetIntegrationSnippetAsCloudInitWritableFile(intCfg IntegrationConfig, username string) (*string, error)
```

Same as `GetIntegrationSnippet` but indents every line by 6 spaces, ready to embed as
a `write_files` entry in a cloud-init YAML:

```yaml
write_files:
  - content: |
      #!/bin/bash
      # rendered snippet here — each line indented 6 spaces
```

---

## Config Flow

Integration args enter via `ContextArgs` at `mc.Init()` time, which calls each package's
`Init()` to store them as package-level state:

```go
// Caller sets one of (mutually exclusive in practice, but not validated):
mCtxArgs.GHRunnerArgs  = &github.GithubRunnerArgs{...}
mCtxArgs.CirrusPWArgs  = &cirrus.PersistentWorkerArgs{...}
mCtxArgs.GLRunnerArgs  = &gitlab.GitLabRunnerArgs{...}

// mc.Init() calls:
github.Init(ca.GHRunnerArgs)   // nil-safe; sets package-level runnerArgs
cirrus.Init(ca.CirrusPWArgs)
gitlab.Init(ca.GLRunnerArgs)
```

Cloud-config builders then retrieve via `<pkg>.GetRunnerArgs()` or
`<pkg>.GetIntegrationConfig()` and pass the result to `GetIntegrationSnippet`.

---

## Usage Pattern in a Cloud-Config Builder

```go
// In pkg/target/host/<target>/<target>.go:
snippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(
    github.GetRunnerArgs(),   // returns nil if not configured → empty snippet
    username,
)
// Embed snippet into the cloud-init write_files section
```

---

## Known Gaps

- No validation that at most one integration is configured (multiple could be set simultaneously)
- Runner versions are compile-time constants; upgrading requires a full rebuild and release
- The GitLab runner integration does not appear in the Tekton task templates (verify)

---

## When to Extend

Add a new file under `specs/integrations/` when:
- Adding a new CI system (e.g. Jenkins, TeamCity)
- Making runner versions runtime-configurable instead of compile-time
- Adding support for runner groups or additional registration parameters
