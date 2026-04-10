# Spec: AWS Mac Pool Service

## Status
Implemented

## Context
A managed pool of macOS dedicated hosts providing request/release semantics to CI pipelines.
Entry point: `pkg/provider/aws/action/mac-pool/mac-pool.go`.
CLI: `cmd/mapt/cmd/aws/services/mac-pool.go`.

The pool runs a serverless HouseKeeper on a recurring schedule (ECS Fargate) that maintains
the desired offered capacity by adding or removing machines while respecting AWS's 24h minimum
host tenancy. State is stored per-machine in separate Pulumi stacks under a shared S3 prefix.

## Problem
This feature is implemented. This spec documents the architecture and known gaps.

## Requirements
- [x] `create`: provision N machines (OfferedCapacity) up to MaxSize; start the HouseKeeper scheduler
- [x] `create`: generate a least-privilege IAM user/key pair for request/release operations (`requestReleaserAccount`)
- [x] `housekeeper`: add machines if current offered capacity < desired and pool size < max
- [x] `housekeeper`: remove machines if current offered capacity > desired and machines are > 24h old (destroyable)
- [x] `request`: lock the next available (non-locked) machine and write its connection details
- [x] `release`: unlock a machine identified by host ID, resetting it for the next user
- [x] Reject local `file://` BackedURL — pool requires remote S3 state
- [x] `destroy`: remove IAM resources, serverless scheduler, and S3 state

## Out of Scope
- Single mac host (see `003-aws-mac-host.md`)
- Integration-mode selection on `request` (currently hardcoded; TODO in code)

## Affected Areas
- `pkg/provider/aws/action/mac-pool/` — orchestration
- `pkg/provider/aws/modules/mac/` — host, machine, util sub-packages
- `pkg/provider/aws/modules/serverless/` — HouseKeeper recurring task
- `pkg/provider/aws/modules/iam/` — request/releaser IAM account
- `cmd/mapt/cmd/aws/services/mac-pool.go`

## Known Gaps / Improvement Ideas
- `Request` integration-mode is hardcoded (TODO comment at `mac-pool.go:138`)
- `destroyCapacity` has a TODO about allocation time ordering
- `getNextMachineForRequest` picks the newest machine; could be optimized (e.g. LRU)
- No explicit handling when all machines in the pool are locked and none available

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- Pool creates N dedicated hosts and writes IAM credentials
- `housekeeper` invocation adds a machine when pool is below capacity
- `request` writes `host`, `username`, `id_rsa` for a locked machine
- `release` makes the machine available again for the next request

---

## Command

```
mapt aws mac-pool create   [flags]   # create the pool of dedicated hosts
mapt aws mac-pool destroy  [flags]
mapt aws mac-pool request  [flags]   # borrow a machine from the pool
mapt aws mac-pool release  [flags]   # return a machine to the pool
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |

No compute-request, spot, timeout, or integration flags.

### Target-specific flags (create)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--name` | string | — | Pool name (used to identify the resource group) |
| `--arch` | string | `m1` | MAC architecture: `x86`, `m1`, `m2` |
| `--version` | string | *(per arch)* | macOS version |
| `--offered-capacity` | int | *(default in action)* | Number of machines kept available in the pool |
| `--max-size` | int | *(default in action)* | Maximum number of dedicated hosts in the pool |
| `--fixed-location` | bool | false | Force creation in `AWS_DEFAULT_REGION` only |
| `--conn-details-output` | string | — | Path to write IAM credentials |
| `--tags` | map | — | Resource tags |

### Request / release flags

`--project-name`, `--backed-url` (from common)

### Destroy flags

`--force-destroy`, `--keep-state`

### Action args struct populated

`mac.MacPoolArgs` → `pkg/provider/aws/action/mac/mac-pool.go`
