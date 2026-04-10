# Spec: Azure RHEL Host

## Status
Implemented

## Context
Provisions a RHEL VM on Azure. Entry point: `pkg/provider/azure/action/rhel/`.
CLI: `cmd/mapt/cmd/azure/hosts/rhel.go`.

Azure RHEL uses Azure Marketplace images. Root disk expansion is handled via a shell script
(`expand-root-disk.sh`) run during cloud-init since Azure RHEL images often ship with a small root partition.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision a RHEL VM on Azure using the Marketplace image
- [x] Expand the root disk during cloud-init to use the full allocated disk size
- [x] Support spot (Azure low-priority / spot VMs) via `azure/modules/allocation/`
- [x] Support optional CI integrations
- [x] Write output files: `host`, `username`, `id_rsa`
- [x] `destroy` cleans up all Azure resources and state

## Out of Scope
- AWS RHEL (see `001-aws-rhel-host.md`)
- Azure RHEL AI (see `015-azure-rhel-ai.md`)

## Affected Areas
- `pkg/provider/azure/action/rhel/` — including `expand-root-disk.sh`
- `pkg/provider/azure/modules/` — network, virtual-machine, allocation
- `cmd/mapt/cmd/azure/hosts/rhel.go`
- `tkn/template/infra-azure-rhel.yaml`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt azure rhel create ...` provisions an accessible RHEL VM
- Root disk is expanded to the configured size
- SSH access works
- `mapt azure rhel destroy ...` removes all resources

---

## Command

```
mapt azure rhel create  [flags]
mapt azure rhel destroy [flags]
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
| `--version` | string | `9.7` | RHEL major.minor version |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--username` | string | `rhqp` | OS username for SSH access |
| `--rh-subscription-username` | string | — | Red Hat subscription username |
| `--rh-subscription-password` | string | — | Red Hat subscription password |
| `--snc` | bool | false | Apply SNC profile |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

*(none beyond common)*

### Action args struct populated

`rhel.RhelArgs` → `pkg/provider/azure/action/rhel/rhel.go`
