# Spec: AWS RHEL Host

## Status
Implemented

## Context
Provisions a RHEL EC2 instance on AWS. This is the reference implementation of the AWS EC2 host
pattern — all other AWS EC2 host targets follow the same structure.

Relevant existing files:
- `pkg/provider/aws/action/rhel/` — orchestration (reference implementation)
- `pkg/target/host/rhel/cloud-config.go` — cloud-config builder
- `cmd/mapt/cmd/aws/hosts/rhel.go` — CLI

## Problem
This feature is fully implemented. This spec documents current behaviour, the mandatory module
sequence, and known gaps. Use it as the template when adding a new AWS EC2 host target.

## Requirements
- [x] Provision a RHEL EC2 instance (versions: 9.x, 8.x) for x86_64 or arm64
- [x] Register with Red Hat Subscription Manager using `SubsUsername` / `SubsPassword` via cloud-init
- [x] Support spot instance allocation with cross-region best-bid selection
- [x] Support on-demand allocation using the default AWS region
- [x] Support airgap topology: two-phase stack update (connectivity ON then OFF)
- [x] Optionally apply the `profileSNC` cloud-config variant to pre-install SNC dependencies
- [x] Optionally schedule serverless self-destruct after a given timeout (requires remote BackedURL)
- [x] Write output files: `host`, `username`, `id_rsa` (and bastion files when airgap)
- [x] `destroy` cleans up main stack, spot stack (if exists), and S3 state

## Out of Scope
- RHEL AI variant (see `009-aws-rhel-ai.md`)
- Azure RHEL (see `010-azure-rhel-host.md`)

## Must Reuse

**In `Create()`:**
- `mc.Init(mCtxArgs, aws.Provider())` — context initialisation
- `allocation.Allocation(mCtx, &AllocationArgs{Prefix, ComputeRequest, AMIProductDescription, Spot})` — resolves region/AZ/instance types for spot or on-demand

**In `deploy()`, in this order:**
- `amiSVC.GetAMIByName(ctx, amiRegex, nil, map[string]string{"architecture": arch})` — finds the RHEL AMI
- `network.Create(ctx, mCtx, &NetworkArgs{Prefix, ID, Region, AZ, CreateLoadBalancer, Airgap, AirgapPhaseConnectivity})` — VPC/subnet/IGW/LB
- `keypair.KeyPairRequest{Name: resourcesUtil.GetResourceName(...)}.Create(ctx, mCtx)` — SSH keypair
- `securityGroup.SGRequest{...}.Create(ctx, mCtx)` — security group (SSH/22 ingress)
- `rhelApi.CloudConfigArgs{...}.GenerateCloudConfig(ctx, mCtx.RunID())` — RHEL cloud-config with subscription and optional SNC profile
- `compute.ComputeRequest{...}.NewCompute(ctx)` — EC2 instance
- `serverless.OneTimeDelayedTask(...)` — only when `Timeout != ""`
- `c.Readiness(ctx, command.CommandCloudInitWait, ...)` — waits for cloud-init to complete

**In `Destroy()`:**
- `aws.DestroyStack(mCtx, DestroyStackRequest{Stackname: stackName})`
- `spot.Destroy(mCtx)` guarded by `spot.Exist(mCtx)`
- `aws.CleanupState(mCtx)`

**In `manageResults()`:**
- `bastion.WriteOutputs(stackResult, prefix, resultsPath)` — only when `airgap=true`
- `output.Write(stackResult, resultsPath, results)` — writes `host`, `username`, `id_rsa`

**Naming:**
- All resource names via `resourcesUtil.GetResourceName(prefix, awsRHELDedicatedID, suffix)`
- Stack name via `mCtx.StackNameByProject(stackName)`

## Must Create
- `pkg/provider/aws/action/rhel/rhel.go` — `RHELArgs`, `Create()`, `Destroy()`, `deploy()`, `manageResults()`, `securityGroups()`
- `pkg/provider/aws/action/rhel/constants.go` — `stackName`, `awsRHELDedicatedID`, `amiRegex`, `diskSize`, `amiProduct`, `amiUserDefault`, output key constants
- `pkg/target/host/rhel/cloud-config.go` — `CloudConfigArgs`, `GenerateCloudConfig()`
- `pkg/target/host/rhel/cloud-config-base` — base cloud-config template file
- `pkg/target/host/rhel/cloud-config-snc` — SNC-variant cloud-config template file
- `cmd/mapt/cmd/aws/hosts/rhel.go` — Cobra `create` and `destroy` subcommands
- `tkn/template/infra-aws-rhel.yaml` — Tekton task template

## Known Gaps
- `createAirgapMachine()` swallows the phase-1 error: returns `nil` instead of `err` at `rhel.go:167`
  — phase 2 must not run if phase 1 fails
- No validation that `SubsUsername`/`SubsPassword` are non-empty when `profileSNC=true`
- `diskSize` is a hardcoded constant; not exposed as a CLI flag

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt aws rhel create --backed-url s3://... --project-name test --version 9 --arch x86_64 --subs-username u --subs-user-pass p` exits 0
- Output directory contains `host`, `username`, `id_rsa`
- SSH access to the provisioned host succeeds
- `mapt aws rhel destroy --backed-url s3://... --project-name test` exits 0 and removes state

---

## Command

```
mapt aws rhel create  [flags]
mapt aws rhel destroy [flags]
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
| `--version` | string | `9.4` | RHEL major.minor version |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--rh-subscription-username` | string | — | Red Hat subscription username |
| `--rh-subscription-password` | string | — | Red Hat subscription password |
| `--snc` | bool | false | Apply SNC profile (sets `nested-virt=true`) |
| `--airgap` | bool | false | Provision as airgap machine (bastion access only) |
| `--timeout` | string | — | Self-destruct duration e.g. `4h` (requires remote `--backed-url`) |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags `name=value,...` |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`rhel.RHELArgs` → `pkg/provider/aws/action/rhel/rhel.go`
