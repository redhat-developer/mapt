# Spec: Azure Windows Desktop Host

## Context
Provisions a Windows Desktop VM on Azure. Entry point: `pkg/provider/azure/action/windows/`.
CLI: `cmd/mapt/cmd/azure/hosts/windows.go`.

This differs from the AWS Windows Server target: it targets Windows Desktop editions on Azure
and includes CI-specific setup scripts (`rhqp-ci-setup.ps1`).

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [ ] Provision a Windows Desktop VM on Azure using the specified Marketplace image
- [ ] Run CI setup PowerShell scripts via custom script extension or userdata
- [ ] Support optional spot (low-priority) VMs
- [ ] Open security group rules for RDP (3389) and WinRM/SSH as needed
- [ ] Write output files: `host`, `username`, `userpassword`
- [ ] `destroy` cleans up all Azure resources and state

## Out of Scope
- AWS Windows Server (see `002-aws-windows-server-host.md`)
- Azure RHEL or Linux (see `010-azure-rhel-host.md`, `013-azure-linux-host.md`)

## Affected Areas
- `pkg/provider/azure/action/windows/` — including `rhqp-ci-setup.ps1`
- `cmd/mapt/cmd/azure/hosts/windows.go`
- `tkn/template/infra-azure-windows-desktop.yaml`

## Acceptance Criteria
- `mapt azure windows create ...` provisions an accessible Windows VM
- RDP connection works with the output credentials
- `mapt azure windows destroy ...` removes all resources

---

## Command

```
mapt azure windows create  [flags]
mapt azure windows destroy [flags]
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
| `--windows-version` | string | `11` | Windows major version |
| `--feature` | string | — | Windows feature/edition variant |
| `--username` | string | `rhqp` | Username for SSH access |
| `--admin-username` | string | `rhqpadmin` | Admin username for RDP access |
| `--profile` | []string | — | Setup profiles to apply (comma-separated) |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

*(none beyond common)*

### Action args struct populated

`windows.WindowsArgs` → `pkg/provider/azure/action/windows/windows.go`
