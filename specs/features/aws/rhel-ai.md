# Spec: AWS RHEL AI Host

## Context
Provisions a RHEL AI instance on AWS, designed for AI/ML workloads. Entry point:
`pkg/provider/aws/action/rhel-ai/`. API: `pkg/target/host/rhelai/`.
CLI: `cmd/mapt/cmd/aws/hosts/rhelai.go`.

RHEL AI differs from standard RHEL in that it uses specialised GPU-capable instance types
and a RHEL AI-specific AMI.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [ ] Provision a RHEL AI instance using the RHEL AI AMI
- [ ] Target GPU-capable instance types (e.g. g4dn, p3 families)
- [ ] Support spot allocation
- [ ] Write output files: `host`, `username`, `id_rsa`
- [ ] `destroy` cleans up all resources and state

## Out of Scope
- Standard RHEL (see `001-aws-rhel-host.md`)
- Azure RHEL AI (see `015-azure-rhel-ai.md`)

## Affected Areas
- `pkg/provider/aws/action/rhel-ai/`
- `pkg/target/host/rhelai/`
- `cmd/mapt/cmd/aws/hosts/rhelai.go`
- `tkn/template/infra-aws-rhel-ai.yaml`
- `Pulumi.rhelai.yaml` — stack configuration for the rhelai Pulumi stack

## Acceptance Criteria
- `mapt aws rhel-ai create ...` provisions an accessible RHEL AI instance
- SSH access works
- `mapt aws rhel-ai destroy ...` removes all resources

---

## Command

```
mapt aws rhel-ai create  [flags]
mapt aws rhel-ai destroy [flags]
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
| `--version` | string | `3.0.0` | RHEL AI version |
| `--accelerator` | string | `cuda` | GPU accelerator type: `cuda` or `rocm` |
| `--custom-ami` | string | — | Override with a custom AMI ID |
| `--timeout` | string | — | Self-destruct duration |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`rhelai.RHELAIArgs` → `pkg/provider/aws/action/rhelai/rhelai.go`
