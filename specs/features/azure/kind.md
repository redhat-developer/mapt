# Spec: Azure Kind Cluster

## Status
Implemented

## Context
Provisions a Kind (Kubernetes-in-Docker) cluster on an Azure VM.
Entry point: `pkg/provider/azure/action/kind/`. CLI: `cmd/mapt/cmd/azure/services/kind.go`.

Mirrors the AWS Kind target but runs on Azure infrastructure.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision an Azure VM and install Kind + Docker via cloud-init
- [x] Create a Kind cluster; export kubeconfig
- [x] Support configurable Kubernetes version
- [x] Support spot (low-priority) VMs
- [x] Write output files: `host`, `username`, `id_rsa`, `kubeconfig`
- [x] `destroy` cleans up all resources and state

## Out of Scope
- AWS Kind (see `007-aws-kind.md`)
- Azure AKS managed clusters (see `012-azure-aks.md`)

## Affected Areas
- `pkg/provider/azure/action/kind/`
- `cmd/mapt/cmd/azure/services/kind.go`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt azure kind create ...` produces a working kubeconfig
- `kubectl get nodes` returns a Ready node
- `mapt azure kind destroy ...` removes all resources

---

## Command

```
mapt azure kind create  [flags]
mapt azure kind destroy [flags]
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
| `--version` | string | `v1.34` | Kubernetes version for Kind |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--extra-port-mappings` | string | — | JSON array of `{containerPort, hostPort, protocol}` |
| `--conn-details-output` | string | — | Path to write kubeconfig |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`

### Action args struct populated

`kind.KindArgs` → `pkg/provider/azure/action/kind/kind.go`
