# Spec: Airgap Network Topology

## Context
An optional network topology that isolates the target instance from the public internet while
still allowing SSH access via a bastion host. Implemented as a two-phase Pulumi stack update.

Key files:
- `pkg/provider/aws/modules/network/airgap/airgap.go` — VPC/subnet creation
- `pkg/provider/aws/modules/network/network.go` — dispatcher (standard vs airgap)
- `pkg/provider/aws/modules/bastion/bastion.go` — bastion host resource

The same Pulumi stack is applied twice:
1. Phase 1 (`connectivity = ON`): NAT gateway present → machine can reach internet for bootstrapping
2. Phase 2 (`connectivity = OFF`): NAT gateway removed → machine loses egress, bastion still accessible

## Problem
This feature is implemented for AWS RHEL and Windows. This spec documents the design and gaps.

## Requirements
- [ ] Create a VPC with a public subnet (has internet gateway + NAT gateway in phase 1) and a private subnet (target)
- [ ] Phase 1: private subnet has route to NAT gateway; cloud-init runs and machine is bootstrapped
- [ ] Phase 2: NAT gateway is removed; private subnet loses egress; machine is isolated
- [ ] Bastion host in the public subnet provides SSH proxy access throughout both phases
- [ ] Write bastion output files alongside target files (`bastion-host`, `bastion-username`, `bastion-id_rsa`)
- [ ] Targets using airgap: RHEL, Windows Server (AWS); extensible to other targets

## Out of Scope
- Azure airgap (not currently implemented)
- Egress filtering via security groups or NACLs (only NAT removal is used)

## Affected Areas
- `pkg/provider/aws/modules/network/` — standard and airgap network implementations
- `pkg/provider/aws/modules/bastion/` — bastion host and output writing
- `pkg/provider/aws/action/rhel/rhel.go` — `createAirgapMachine()` orchestration
- `pkg/provider/aws/action/windows/windows.go` — same

## Known Gaps / Improvement Ideas
- The error from phase 1 of `createAirgapMachine()` is swallowed in both rhel and windows actions
  (`return nil` instead of `return err`) — this is a bug; phase 2 should not run if phase 1 fails
- No validation that `Airgap=true` requires a remote BackedURL (unlike serverless timeout which does validate)

## Acceptance Criteria
- `mapt aws rhel create --airgap ...` provisions an instance accessible only through the bastion
- Direct SSH to the target host's public IP fails; SSH via bastion succeeds
- Phase 2 is confirmed complete by checking the target cannot reach an external host

---

## Command

This is a cross-cutting feature, not a standalone command. It is activated via the
`--airgap` flag on individual target create commands:

```
mapt aws rhel    create --airgap ...
mapt aws windows create --airgap ...
```

The `--airgap` flag is defined locally in each host cmd file (not in shared params).
No additional flags are specific to the airgap feature itself — the two-phase
connectivity behaviour is controlled internally by the action.
