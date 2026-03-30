# Spec: Azure Linux Host (Fedora / Ubuntu)

## Status
Implemented

## Context
Provisions a generic Linux VM on Azure (Fedora or Ubuntu). Entry point:
`pkg/provider/azure/action/linux/`. CLI: `cmd/mapt/cmd/azure/hosts/linux.go`.
Also referenced as separate Fedora/Ubuntu targets in docs (`docs/azure/fedora.md`, `docs/azure/ubuntu.md`).

This is a general-purpose Linux provisioner for Azure that accepts a configurable image reference.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision a Linux VM on Azure with a configurable Marketplace image (Fedora, Ubuntu, etc.)
- [x] Support spot (low-priority) VMs
- [x] Support optional CI integrations (GitHub runner, Cirrus worker, GitLab runner)
- [x] Write output files: `host`, `username`, `id_rsa`
- [x] `destroy` cleans up all resources and state

## Out of Scope
- Azure RHEL (subscription-managed — see `010-azure-rhel-host.md`)
- AWS Fedora (see `008-aws-fedora-host.md`)

## Affected Areas
- `pkg/provider/azure/action/linux/`
- `pkg/provider/azure/data/` — image reference lookup
- `cmd/mapt/cmd/azure/hosts/linux.go`
- `tkn/template/infra-azure-fedora.yaml`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt azure linux create ...` provisions an accessible Linux VM
- SSH access works
- `mapt azure linux destroy ...` removes all resources

---

## Command

```
mapt azure linux create  [flags]   # Ubuntu default; reused for Fedora with different version
mapt azure linux destroy [flags]
```

### Shared flag groups

| Group | Source | Flags added |
|---|---|---|
| Common | `specs/cmd/params.md` | `--project-name`, `--backed-url` |
| Compute Request | `specs/cmd/params.md` | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `specs/cmd/params.md` | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |
| Integrations | `specs/cmd/params.md` | `--ghactions-runner-*`, `--it-cirrus-pw-*`, `--glrunner-*` |
| Location | `specs/cmd/azure-params.md` | `--location` (default: `westeurope`) |

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `24.04` | OS version (Ubuntu format; `42` for Fedora) |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--username` | string | `rhqp` | OS username for SSH access |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

*(none beyond common)*

### Action args struct populated

`linux.LinuxArgs` → `pkg/provider/azure/action/linux/linux.go`
