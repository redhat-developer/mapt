# CLI Params Layer

**Package:** `github.com/redhat-developer/mapt/cmd/mapt/cmd/params`
**File:** `cmd/mapt/cmd/params/params.go`

Central registry for all reusable CLI flags. Every flag that appears on more than one
`create` command is defined here, not in the individual cmd files. Individual cmd files
only define flags that are unique to that target.

---

## The Three-Part Pattern

Every flag group follows the same structure:

### 1. Constants

```go
// Exported: used by cmd files to read values via viper
const FlagName string = "flag-name"
const FlagNameDesc string = "human readable description"
const FlagNameDefault string = "default-value"  // optional

// Unexported: only used within params.go
const internalFlag string = "internal-flag-name"
```

Use **exported** constants when the cmd file needs to call `viper.GetX(params.FlagName)`
directly. Use **unexported** when the value is only read inside a `*Args()` helper in
this package.

### 2. `Add*Flags(fs *pflag.FlagSet)`

Registers flags on the flagset passed in. Called once per `create` command that needs
this group:

```go
func AddSpotFlags(fs *pflag.FlagSet) {
    fs.Bool(spot, false, spotDesc)
    fs.StringP(spotTolerance, "", spotToleranceDefault, spotToleranceDesc)
    fs.StringSliceP(spotExcludedHostedZones, "", []string{}, spotExcludedHostedZonesDesc)
}
```

### 3. `*Args() *SomeType`

Reads values from viper and returns a populated struct (or `nil` if the feature is not
enabled). Called inside the cmd's `RunE` when building the action args:

```go
func SpotArgs() *spotTypes.SpotArgs {
    if viper.IsSet(spot) {
        return &spotTypes.SpotArgs{ ... }
    }
    return nil  // nil = feature not requested
}
```

Returning `nil` is the canonical "not configured" signal — action code checks for nil
before using the result.

---

## How Viper Binding Works

Each `create` command binds its flagset to viper at the start of `RunE`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    if err := viper.BindPFlags(cmd.Flags()); err != nil {
        return err
    }
    // now viper.GetX(flagName) works for all registered flags
    ...
}
```

After binding, all flag values are accessible via `viper.GetString`, `viper.GetBool`,
`viper.GetInt32`, `viper.GetStringSlice`, `viper.IsSet`, etc.

---

## Existing Flag Groups

### Common (every command)

```go
func AddCommonFlags(fs *pflag.FlagSet)
```

| Flag | Type | Description |
|---|---|---|
| `project-name` | string | Pulumi project name |
| `backed-url` | string | State backend URL (`file://`, `s3://`, `azblob://`) |

Added to the parent command's `PersistentFlags` so it applies to all subcommands.

---

### Debug

```go
func AddDebugFlags(fs *pflag.FlagSet)
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `debug` | bool | false | Enable debug traces |
| `debug-level` | uint | 3 | Verbosity 1–9 |

---

### Compute Request

```go
func AddComputeRequestFlags(fs *pflag.FlagSet)
func ComputeRequestArgs() *cr.ComputeRequestArgs
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `cpus` | int32 | 8 | vCPU count |
| `memory` | int32 | 64 | RAM in GiB |
| `gpus` | int32 | 0 | GPU count |
| `gpu-manufacturer` | string | — | e.g. `NVIDIA` |
| `nested-virt` | bool | false | Require nested virtualisation support |
| `compute-sizes` | []string | — | Override selector; comma-separated instance types |
| `arch` | string | `x86_64` | `x86_64` or `arm64` |

`ComputeRequestArgs()` maps `arch` to `cr.Amd64` / `cr.Arm64`. When `--snc` is set,
`NestedVirt` is forced true regardless of `--nested-virt`.

See `specs/api/provider-interfaces.md` for `ComputeRequestArgs` type.

---

### Spot

```go
func AddSpotFlags(fs *pflag.FlagSet)
func SpotArgs() *spotTypes.SpotArgs  // returns nil when --spot not set
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `spot` | bool | false | Enable spot selection |
| `spot-eviction-tolerance` | string | `lowest` | `lowest`/`low`/`medium`/`high`/`highest` |
| `spot-increase-rate` | int | 30 | Bid price % above current price |
| `spot-excluded-regions` | []string | — | Regions to skip |

Returns `nil` when `--spot` is not set — this signals on-demand to allocation.

See `specs/api/provider-interfaces.md` for `SpotArgs` type.

---

### Network (to be added — `specs/features/aws/vpc-endpoints.md`)

```go
func AddNetworkFlags(fs *pflag.FlagSet)
func NetworkEndpoints() []string
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `endpoints` | []string | — | VPC endpoints to create: `s3`, `ecr`, `ssm` |

---

### GitHub Actions Runner

```go
func AddGHActionsFlags(fs *pflag.FlagSet)
func GithubRunnerArgs() *github.GithubRunnerArgs  // returns nil when token not set
```

| Flag | Type | Description |
|---|---|---|
| `ghactions-runner-token` | string | Registration token |
| `ghactions-runner-repo` | string | Repository or org URL |
| `ghactions-runner-labels` | []string | Runner labels |

Returns `nil` when `--ghactions-runner-token` is not set.
Platform and arch are derived from `--arch`; not user-configurable at CLI level.

---

### Cirrus CI Persistent Worker

```go
func AddCirrusFlags(fs *pflag.FlagSet)
func CirrusPersistentWorkerArgs() *cirrus.PersistentWorkerArgs  // returns nil when token not set
```

| Flag | Type | Description |
|---|---|---|
| `it-cirrus-pw-token` | string | Cirrus registration token |
| `it-cirrus-pw-labels` | map[string]string | Labels as `key=value` pairs |

Returns `nil` when `--it-cirrus-pw-token` is not set.

---

### GitLab Runner

```go
func AddGitLabRunnerFlags(fs *pflag.FlagSet)
func GitLabRunnerArgs() *gitlab.GitLabRunnerArgs  // returns nil when token not set
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `glrunner-token` | string | — | GitLab Personal Access Token |
| `glrunner-project-id` | string | — | Project ID (mutually exclusive with group ID) |
| `glrunner-group-id` | string | — | Group ID (mutually exclusive with project ID) |
| `glrunner-url` | string | `https://gitlab.com` | GitLab instance URL |
| `glrunner-tags` | []string | — | Runner tags |

Returns `nil` when `--glrunner-token` is not set.

---

### Serverless / Destroy

```go
// No Add* helper — these are registered directly in each destroy command
```

| Flag | Type | Description | Command |
|---|---|---|---|
| `timeout` | string | Go duration string — schedules self-destruct | create |
| `serverless` | bool | Use role-based credentials (ECS context) | destroy |
| `force-destroy` | bool | Destroy even if locked | destroy |
| `keep-state` | bool | Keep Pulumi state in S3 after destroy | destroy |

---

## Arch Conversion Helpers

Each integration has its own `Platform`/`Arch` type. Params provides private converters:

```go
func linuxArchAsGithubActionsArch(arch string) *github.Arch  // "x86_64" → &Amd64
func linuxArchAsCirrusArch(arch string) *cirrus.Arch
func linuxArchAsGitLabArch(arch string) *gitlab.Arch

// Exported variants for MAC commands (different arch string convention):
func MACArchAsCirrusArch(arch string) *cirrus.Arch   // "x86" → &Amd64
func MACArchAsGitLabArch(arch string) *gitlab.Arch
```

---

## How to Add a New Flag Group

1. **Add constants** in the `const` block — unexported flag name, exported description
2. **Add `Add*Flags(fs *pflag.FlagSet)`** — register each flag with the appropriate type
   (`Bool`, `StringP`, `StringSliceP`, `Int32P`, `StringToStringP`)
3. **Add `*Args() *SomeType`** — read from viper and return a populated struct or `nil`
4. **Call `Add*Flags`** in each cmd create function that needs the group
5. **Call `*Args()`** in the `RunE` body when building the action args struct

For a single-target flag (not shared), define it with a local constant in the target's
cmd file instead, and read it directly with `viper.GetX(localConst)`.

---

## When to Extend This File

Update this spec when:
- Adding a new shared flag group (e.g. `AddNetworkFlags` for VPC endpoints)
- Adding flags to an existing group
- Adding a new arch conversion helper for a new integration
