# Spec: AWS Fedora Host

## Status
Implemented

## Context
Provisions a Fedora EC2 instance on AWS. Entry point: `pkg/provider/aws/action/fedora/`.
Cloud-config: `pkg/target/host/fedora/`. CLI: `cmd/mapt/cmd/aws/hosts/fedora.go`.

Fedora on AWS is used for Fedora-specific testing. The instance uses a cloud-init config
with the Fedora cloud image.

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Provision a Fedora EC2 instance (latest or specified version)
- [x] Support spot instance allocation
- [x] Support optional CI integrations (GitHub runner, Cirrus worker, GitLab runner)
- [x] Write output files: `host`, `username`, `id_rsa`
- [x] `destroy` cleans up stack, spot stack, S3 state

## Out of Scope
- Azure Fedora (see docs/azure/fedora.md — currently Azure Linux target)
- RHEL (subscription-managed — see `001-aws-rhel-host.md`)

## Affected Areas
- `pkg/provider/aws/action/fedora/`
- `pkg/target/host/fedora/` — cloud-config
- `cmd/mapt/cmd/aws/hosts/fedora.go`
- `tkn/template/infra-aws-fedora.yaml`

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt aws fedora create ...` provisions an accessible Fedora instance
- SSH access works with the output key
- `mapt aws fedora destroy ...` removes all resources

---

## Command

```
mapt aws fedora create  [flags]
mapt aws fedora destroy [flags]
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |
| Compute Request | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |
| Integrations | `--ghactions-runner-*`, `--it-cirrus-pw-*`, `--glrunner-*` |

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `41` | Fedora Cloud major version |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--airgap` | bool | false | Provision as airgap machine |
| `--timeout` | string | — | Self-destruct duration |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`fedora.FedoraArgs` → `pkg/provider/aws/action/fedora/fedora.go`
