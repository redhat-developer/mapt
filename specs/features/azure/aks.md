# Spec: Azure AKS (Azure Kubernetes Service)

## Context
Provisions a managed AKS cluster on Azure. Entry point: `pkg/provider/azure/action/aks/`.
CLI: `cmd/mapt/cmd/azure/services/aks.go`.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [ ] Provision an AKS cluster with a configurable node pool
- [ ] Support configurable Kubernetes version
- [ ] Support spot node pools (Azure spot VMs)
- [ ] Write kubeconfig output file
- [ ] `destroy` cleans up all resources and state

## Out of Scope
- AWS EKS (see `006-aws-eks.md`)
- Azure Kind (see `014-azure-kind.md`)

## Affected Areas
- `pkg/provider/azure/action/aks/`
- `cmd/mapt/cmd/azure/services/aks.go`
- `tkn/template/infra-azure-aks.yaml`

## Acceptance Criteria
- `mapt azure aks create ...` provisions a functioning AKS cluster
- Exported kubeconfig allows `kubectl get nodes` to return Ready nodes
- `mapt azure aks destroy ...` removes all resources

---

## Command

```
mapt azure aks create  [flags]
mapt azure aks destroy [flags]
```

### Shared flag groups

| Group | Source | Flags added |
|---|---|---|
| Common | `specs/cmd/params.md` | `--project-name`, `--backed-url` |
| Spot | `specs/cmd/params.md` | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |

Note: no compute-request (VM size is explicit), no integrations, no timeout.
AKS uses its own `--location` rather than the shared azure-params one (different default: `West US`).

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--location` | string | `West US` | Azure region (ignored when spot is set) |
| `--vmsize` | string | *(default in action)* | Explicit VM size for node pool |
| `--version` | string | `1.31` | Kubernetes version |
| `--only-system-pool` | bool | false | Create system node pool only (no user pool) |
| `--enable-app-routing` | bool | false | Enable AKS App Routing add-on |
| `--conn-details-output` | string | — | Path to write kubeconfig |
| `--tags` | map | — | Resource tags |

### Destroy flags

*(none beyond common)*

### Action args struct populated

`aks.AKSArgs` → `pkg/provider/azure/action/aks/aks.go`
