# Spec: Azure RHEL AI Host

## Status
Implemented

## Context
Provisions a RHEL AI VM on Azure for AI/ML workloads. Entry point:
`pkg/provider/azure/action/rhel-ai/`. CLI: `cmd/mapt/cmd/azure/hosts/rhelai.go`.

Mirrors the AWS RHEL AI target on Azure infrastructure, using GPU-capable VM sizes
and the RHEL AI Marketplace image.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision a RHEL AI VM on Azure using the Marketplace image
- [x] Target GPU-capable Azure VM sizes
- [x] Support spot (low-priority) VMs
- [x] Write output files: `host`, `username`, `id_rsa`
- [x] `destroy` cleans up all Azure resources and state

## Out of Scope
- AWS RHEL AI (see `009-aws-rhel-ai.md`)
- Standard Azure RHEL (see `010-azure-rhel-host.md`)

## Affected Areas
- `pkg/provider/azure/action/rhel-ai/`
- `cmd/mapt/cmd/azure/hosts/rhelai.go`
- `tkn/template/infra-azure-rhel-ai.yaml`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt azure rhel-ai create ...` provisions an accessible RHEL AI VM
- SSH access works
- `mapt azure rhel-ai destroy ...` removes all resources

---

## Command

```
mapt azure rhel-ai create  [flags]
mapt azure rhel-ai destroy [flags]
```

### Shared flag groups

| Group | Source | Flags added |
|---|---|---|
| Common | `specs/cmd/params.md` | `--project-name`, `--backed-url` |
| Compute Request | `specs/cmd/params.md` | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `specs/cmd/params.md` | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |
| Location | `specs/cmd/azure-params.md` | `--location` (default: `westeurope`) |

Note: no integration flags.

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `3.0.0` | RHEL AI version |
| `--accelerator` | string | `cuda` | GPU accelerator: `cuda` or `rocm` |
| `--custom-ami` | string | — | Custom image override |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`rhelai.RHELAIArgs` → `pkg/provider/azure/action/rhelai/rhelai.go`
