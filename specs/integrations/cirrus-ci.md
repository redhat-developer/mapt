# Integration: Cirrus CI Persistent Worker

**Package:** `github.com/redhat-developer/mapt/pkg/integrations/cirrus`

Registers the provisioned machine as a Cirrus CI persistent worker at boot.
The cirrus-cli binary is downloaded and configured as a long-running service.

See `specs/integrations/overview.md` for the shared interface and config flow.

---

## Type

```go
type PersistentWorkerArgs struct {
    Name     string            // Worker name — set to mCtx.RunID() by the action
    Token    string            // Cirrus CI registration token (required)
    Platform *Platform         // Target OS: Linux | Darwin | Windows
    Arch     *Arch             // Target arch: Amd64 | Arm64
    Labels   map[string]string // Worker labels as key=value pairs
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
)
```

---

## Persistent Worker Version

```go
var version = "v0.135.0"  // overridden at build time via linker flag
```

Makefile variable: `CIRRUS_CLI`
Linker target: `pkg/integrations/cirrus.version`

---

## Download URL Pattern

```
https://github.com/cirruslabs/cirrus-cli/releases/download/{version}/cirrus-{platform}-{arch}
https://github.com/cirruslabs/cirrus-cli/releases/download/{version}/cirrus-{platform}-{arch}.exe  (Windows)
```

---

## Listen Port

```go
var cirrusPort = "3010"
```

The worker listens on port `3010`. This port must be opened in the security group when
Cirrus integration is enabled — callers use `cirrus.CirrusPort()` to conditionally add
the ingress rule:

```go
func CirrusPort() (*int, error)  // returns nil, nil if Cirrus not configured
```

This is the only integration that requires an additional inbound port.

---

## Functions

```go
func Init(args *PersistentWorkerArgs)        // stores args as package-level state
func GetRunnerArgs() *PersistentWorkerArgs   // returns nil if not configured
func GetToken() string                       // returns token or "" if not configured
func CirrusPort() (*int, error)              // returns port int or nil if not configured
```

---

## UserDataValues populated

| Field | Source |
|---|---|
| `CliURL` | `downloadURL()` — version + platform + arch |
| `Name` | `PersistentWorkerArgs.Name` |
| `Token` | `PersistentWorkerArgs.Token` |
| `Labels` | Map entries formatted as `key=value`, joined with `,` |
| `Port` | `"3010"` (fixed) |
| `User` | Set by `GetIntegrationSnippet` from `username` arg |
| `RepoURL`, `Executor` | Not used |

---

## Script Templates

Embedded at compile time:
- `snippet-linux.sh` — downloads binary, installs as systemd service
- `snippet-darwin.sh` — same flow for macOS
- `snippet-windows.ps1` — downloads `.exe`, installs as Windows service

Template selection is based on `PersistentWorkerArgs.Platform`.
