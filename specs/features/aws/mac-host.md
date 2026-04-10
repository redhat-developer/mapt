# Spec: AWS Mac Host (Single)

## Status
Implemented

## Context
Provisions a single macOS instance on an AWS Dedicated Host. Entry point:
`pkg/provider/aws/action/mac/`. Modules: `pkg/provider/aws/modules/mac/`.
CLI: `cmd/mapt/cmd/aws/hosts/mac.go`.

AWS Dedicated Hosts for Mac have a hard constraint: minimum 24-hour tenancy before release.
The mac module handles host allocation, machine setup (via root-volume replacement), and
graceful release respecting the 24h window.

## Problem
This feature is implemented. This spec documents behaviour and the 24h constraint implications.

## Requirements
- [x] Allocate an AWS Dedicated Host for macOS (x86_64 or arm64/Apple Silicon)
- [x] Deploy a macOS machine via root-volume replacement (not standard AMI boot)
- [x] Support optional CI integration: GitHub Actions runner, Cirrus persistent worker, GitLab runner
- [x] Optionally fix the dedicated host to a specific region/AZ (`FixedLocation`)
- [x] Enforce the 24-hour minimum tenancy: do not attempt to release a host allocated < 24h ago
- [x] Write output files: `host`, `username`, `id_rsa`
- [x] `destroy` handles the 24h wait or errors clearly if host is not yet releasable

## Out of Scope
- Mac Pool service (managed pool of mac hosts — see `004-aws-mac-pool-service.md`)
- Windows or Linux hosts

## Affected Areas
- `pkg/provider/aws/action/mac/` — orchestration
- `pkg/provider/aws/modules/mac/host/` — dedicated host allocation
- `pkg/provider/aws/modules/mac/machine/` — machine setup via volume replacement
- `cmd/mapt/cmd/aws/hosts/mac.go`
- `tkn/template/infra-aws-mac.yaml`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt aws mac create ...` exits 0 and writes `host`, `username`, `id_rsa`
- SSH access to the macOS host works
- `mapt aws mac destroy ...` either releases the host (if >= 24h old) or fails with a clear error

---

## Command

```
mapt aws mac create   [flags]
mapt aws mac destroy  [flags]
mapt aws mac request  [flags]   # borrow a machine from the pool
mapt aws mac release  [flags]   # return a machine to the pool
```

Note: `request` and `release` operate on the mac-pool (see `specs/features/aws/mac-pool-service.md`).
A standalone `create` provisions a dedicated host directly without a pool.

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |

Note: no compute-request, no spot, no timeout, no integration flags.
Mac hardware is allocated as a dedicated host — instance type is fixed by arch+version.

### Target-specific flags (create)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--arch` | string | `m1` | MAC architecture: `x86`, `m1`, `m2` |
| `--version` | string | *(per arch)* | macOS version: 11/12 on x86; 13/14/15 on all |
| `--fixed-location` | bool | false | Force creation in `AWS_DEFAULT_REGION` only |
| `--airgap` | bool | false | Provision as airgap machine |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy / request / release flags

`--dedicated-host-id` — required for `request`, `release`, and `destroy` to identify the host

`--force-destroy`, `--keep-state` on destroy.

### Action args struct populated

`mac.MacArgs` → `pkg/provider/aws/action/mac/mac.go`
