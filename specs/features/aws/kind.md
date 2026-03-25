# Spec: AWS Kind Cluster

## Context
Provisions a Kind (Kubernetes-in-Docker) cluster on an EC2 instance.
Entry point: `pkg/provider/aws/action/kind/`. Cloud-config: `pkg/target/service/kind/`.
CLI: `cmd/mapt/cmd/aws/services/kind.go`.

Kind is a lighter-weight alternative to EKS/SNC for CI pipelines that need a disposable
Kubernetes cluster without managed-service overhead.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [ ] Provision an EC2 instance and install Kind + Docker via cloud-init
- [ ] Create a Kind cluster during cloud-init; export kubeconfig
- [ ] Support configurable Kubernetes version (via Kind node image)
- [ ] Support spot instance allocation
- [ ] Write output files: `host`, `username`, `id_rsa`, `kubeconfig`
- [ ] `destroy` cleans up stack and S3 state

## Out of Scope
- Azure Kind (see `014-azure-kind.md`)
- EKS managed clusters (see `006-aws-eks.md`)

## Affected Areas
- `pkg/provider/aws/action/kind/`
- `pkg/target/service/kind/` — cloud-config generation and test
- `cmd/mapt/cmd/aws/services/kind.go`
- `tkn/template/infra-aws-kind.yaml`

## Acceptance Criteria
- `mapt aws kind create ...` produces a working kubeconfig
- `kubectl get nodes` returns a Ready node
- `mapt aws kind destroy ...` removes all resources

---

## Command

```
mapt aws kind create  [flags]
mapt aws kind destroy [flags]
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |
| Compute Request | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |

Note: no integration flags.

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `v1.34` | Kubernetes version for Kind |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--extra-port-mappings` | string | — | JSON array of `{containerPort, hostPort, protocol}` objects |
| `--timeout` | string | — | Self-destruct duration |
| `--conn-details-output` | string | — | Path to write kubeconfig |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`kind.KindArgs` → `pkg/provider/aws/action/kind/kind.go`
