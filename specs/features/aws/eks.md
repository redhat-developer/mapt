# Spec: AWS EKS (Elastic Kubernetes Service)

## Status
Implemented

## Context
Provisions a managed EKS cluster on AWS. Entry point: `pkg/provider/aws/action/eks/`.
CLI: `cmd/mapt/cmd/aws/services/eks.go`.

Unlike the SNC target, EKS uses the AWS-managed control plane and worker node groups
rather than a self-managed cluster on a single EC2 instance.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision an EKS cluster with a managed node group
- [x] Support configurable Kubernetes version
- [x] Support spot instances for worker nodes
- [x] Write kubeconfig output file
- [x] `destroy` cleans up all cluster resources and S3 state

## Out of Scope
- OpenShift SNC (see `005-aws-openshift-snc.md`)
- Azure AKS (see `012-azure-aks.md`)
- AWS Kind (see `007-aws-kind.md`)

## Affected Areas
- `pkg/provider/aws/action/eks/` — orchestration
- `cmd/mapt/cmd/aws/services/eks.go`
- `tkn/template/infra-aws-kind.yaml` (verify — may share template)

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt aws eks create ...` provisions a functioning EKS cluster
- Exported kubeconfig allows `kubectl get nodes` to return Ready nodes
- `mapt aws eks destroy ...` removes all resources

---

## Command

```
mapt aws eks create  [flags]
mapt aws eks destroy [flags]
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |
| Compute Request | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |

Note: no integration flags, no timeout (EKS cluster lifecycle is not self-destructed).

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `1.31` | Kubernetes version |
| `--workers-desired` | int | `1` | Worker node group desired size |
| `--workers-max` | int | `3` | Worker node group maximum size |
| `--workers-min` | int | `1` | Worker node group minimum size |
| `--addons` | []string | — | EKS managed addons to install (comma-separated) |
| `--load-balancer-controller` | bool | false | Install AWS Load Balancer Controller |
| `--excluded-zone-ids` | []string | — | AZ IDs to exclude from node placement |
| `--arch` | string | `x86_64` | Worker node architecture |
| `--conn-details-output` | string | — | Path to write kubeconfig |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--force-destroy`, `--keep-state` (no `--serverless`)

### Action args struct populated

`eks.EKSArgs` → `pkg/provider/aws/action/eks/eks.go`
