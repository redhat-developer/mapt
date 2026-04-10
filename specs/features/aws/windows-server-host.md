# Spec: AWS Windows Server Host

## Status
Implemented

## Context
Provisions a Windows Server EC2 instance on AWS. Follows the standard AWS EC2 host pattern
(see `001-aws-rhel-host.md`) with two additions: AMI cross-region copy and Fast Launch.

Relevant existing files:
- `pkg/provider/aws/action/windows/` — orchestration
- `pkg/provider/aws/modules/ami/` — AMI copy + fast-launch (reused here, not in other targets)
- `pkg/target/host/windows-server/` — PowerShell userdata builder

## Problem
This feature is fully implemented. This spec documents the standard and Windows-specific
module usage, and known gaps.

## Requirements
- [x] Provision Windows Server 2019 (English or non-English variant) EC2 instance
- [x] Accept a custom AMI name/owner/user; fall back to well-known defaults
- [x] Copy the AMI to the target region when not natively available; optionally keep the copy
- [x] Enable Fast Launch on copied AMI with configurable parallelism
- [x] Support spot instance allocation with cross-region best-bid selection
- [x] Support airgap topology (two-phase: connectivity ON → OFF)
- [x] Generate a random administrator password; export as `userpassword`
- [x] Open security group rules for SSH (22) and RDP (3389)
- [x] Optionally schedule serverless self-destruct after timeout
- [x] Write output files: `host`, `username`, `userpassword`, `id_rsa` (and bastion files when airgap)
- [x] `destroy` cleans up main stack, AMI-copy stack (if exists), spot stack (if exists), S3 state

## Out of Scope
- Azure Windows Desktop (see `011-azure-windows-desktop.md`)
- Non-server Windows editions

## Must Reuse

**In `Create()` — standard:**
- `mc.Init(mCtxArgs, aws.Provider())`
- `allocation.Allocation(mCtx, &AllocationArgs{...})` — spot or on-demand

**In `Create()` — Windows-specific addition before `createMachine()`:**
- `data.IsAMIOffered(ctx, ImageRequest{Name, Region})` — check if AMI exists in the target region
- `amiCopy.CopyAMIRequest{..., FastLaunch: true, MaxParallel: N}.Create()` — copy AMI to region when not offered; this creates its own Pulumi stack

**In `deploy()`, in this order — same as standard pattern:**
- `amiSVC.GetAMIByName(ctx, amiName+"*", []string{amiOwner}, nil)`
- `network.Create(ctx, mCtx, &NetworkArgs{..., CreateLoadBalancer: r.spot})`
- `keypair.KeyPairRequest{Name: resourcesUtil.GetResourceName(...)}.Create(ctx, mCtx)`
- `securityGroup.SGRequest{..., IngressRules: [SSH_TCP, RDP_TCP]}.Create(ctx, mCtx)`
- `security.CreatePassword(ctx, resourcesUtil.GetResourceName(...))` — random admin password
- `cloudConfigWindowsServer.GenerateUserdata(ctx, user, password, keyResources, runID)` — PowerShell userdata
- `compute.ComputeRequest{..., LBTargetGroups: []int{22, 3389}}.NewCompute(ctx)`
- `serverless.OneTimeDelayedTask(...)` — only when `Timeout != ""`
- `c.Readiness(ctx, command.CommandPing, ...)` — ICMP ping readiness (not cloud-init wait)

**In `Destroy()` — Windows-specific additions:**
- `aws.DestroyStack(mCtx, DestroyStackRequest{Stackname: stackName})`
- `amiCopy.Destroy(mCtx)` guarded by `amiCopy.Exist(mCtx)` — additional step vs standard pattern
- `spot.Destroy(mCtx)` guarded by `spot.Exist(mCtx)`
- `aws.CleanupState(mCtx)`

**In `manageResults()` — standard:**
- `bastion.WriteOutputs(...)` when airgap
- `output.Write(stackResult, resultsPath, results)` — writes `host`, `username`, `userpassword`, `id_rsa`

**Naming:**
- All resource names via `resourcesUtil.GetResourceName(prefix, awsWindowsDedicatedID, suffix)`

## Must Create
- `pkg/provider/aws/action/windows/windows.go` — `WindowsServerArgs`, `Create()`, `Destroy()`, `deploy()`, `manageResults()`, `securityGroups()`
- `pkg/provider/aws/action/windows/constants.go` — stack name, component ID, AMI defaults, disk size, fast-launch config
- `pkg/target/host/windows-server/windows-server.go` — `GenerateUserdata()`
- `pkg/target/host/windows-server/bootstrap.ps1` — embedded PowerShell bootstrap script
- `cmd/mapt/cmd/aws/hosts/windows.go` — Cobra `create` and `destroy` subcommands
- `tkn/template/infra-aws-windows-server.yaml` — Tekton task template

## Known Gaps
- `createAirgapMachine()` swallows the phase-1 error: `return nil` instead of `return err` at `windows.go:214`
- RDP through the bastion is unfinished — TODO comment at bottom of `windows.go`
- Readiness uses `CommandPing` (ICMP) not `CommandCloudInitWait`; cloud-init completion is not explicitly verified

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- `mapt aws windows create ...` provisions an accessible Windows instance
- RDP port 3389 and SSH port 22 are reachable
- Output directory contains `host`, `username`, `userpassword`, `id_rsa`
- `mapt aws windows destroy ...` removes all stacks and S3 state

---

## Command

```
mapt aws windows create  [flags]
mapt aws windows destroy [flags]
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |
| Spot | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |

Note: no compute-request flags — Windows uses a fixed AMI-based workflow, not hardware-spec selection. No integration flags.

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--ami-name` | string | `Windows_Server-2019-English-Full-Base*` | AMI name pattern to search |
| `--ami-username` | string | `ec2-user` | Default username on the AMI |
| `--ami-region` | string | — | Source region for cross-region AMI copy |
| `--ami-keep-copy` | bool | false | Retain the copied AMI after destroy |
| `--airgap` | bool | false | Provision as airgap machine |
| `--timeout` | string | — | Self-destruct duration |
| `--conn-details-output` | string | — | Path to write connection files |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`windows.WindowsArgs` → `pkg/provider/aws/action/windows/windows.go`
