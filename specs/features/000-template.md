# Spec: [Title]

## Status
<!-- Draft | Accepted | Implemented | Deprecated -->
Draft

## Context
Brief background. What area of the codebase this touches. Links to related existing files.

## Problem
What is missing, broken, or needs improvement.

## Requirements
- [ ] Concrete, testable requirement
- [ ] Another requirement

## Out of Scope
Explicit list of what this spec does NOT cover.

## Design
<!-- Optional. For non-trivial features: data flow, error handling decisions, non-obvious choices.
     Omit when the design is fully captured by Must Reuse / Must Create. -->

## Must Reuse
Existing modules and functions that MUST be called. Do not reimplement this logic.
Reference the API spec for each module's full type signatures.

<!-- For a new AWS EC2 host target, this section is almost always: -->
<!--
**In `Create()`:**
- `allocation.Allocation()` — specs/api/aws/allocation.md

**In `deploy()`, in this order:**
- `amiSVC.GetAMIByName()` — AMI lookup
- `network.Create()` — specs/api/aws/network.md
- `keypair.KeyPairRequest.Create()` — SSH keypair
- `securityGroup.SGRequest.Create()` — specs/api/aws/security-group.md
- `<cloud-config builder>.Generate()` — cloud-init / userdata (Must Create)
- `compute.ComputeRequest.NewCompute()` — specs/api/aws/compute.md
- `serverless.OneTimeDelayedTask()` — only when Timeout is set
- `c.Readiness()` — specs/api/aws/compute.md

**In `Destroy()`:**
- `aws.DestroyStack()` → `spot.Destroy()` if `spot.Exist()` → `aws.CleanupState()`

**In `manageResults()`:**
- `bastion.WriteOutputs()` (airgap only) — specs/api/aws/bastion.md
- `output.Write()` — specs/api/output-contract.md
-->

## Must Create
New files to write. Everything not listed under Must Reuse.

- `pkg/provider/<cloud>/action/<target>/<target>.go`
- `pkg/provider/<cloud>/action/<target>/constants.go`
- `pkg/target/host/<target>/` or `pkg/target/service/<target>/`
- `cmd/mapt/cmd/<cloud>/hosts/<target>.go`
- `tkn/template/infra-<cloud>-<target>.yaml`

## API Changes
List any `specs/api/` files that need updating alongside this feature.

- none

## Tasks
<!-- Ordered implementation checklist. Work top-to-bottom.
     Change Status to Implemented and delete this section when all tasks are done. -->
- [ ] Create `constants.go` — stackName, componentID, AMI regex, ports, disk size
- [ ] Create `<target>.go` — Args struct, `Create()`, `Destroy()`, `deploy()`, `manageResults()`, `securityGroups()`
- [ ] Create cloud-config / userdata builder in `pkg/target/`
- [ ] Create Cobra command in `cmd/`
- [ ] Create Tekton template in `tkn/template/`
- [ ] Verify all Must Reuse calls are present and in the mandatory order
- [ ] Update any `specs/api/` files listed in API Changes
- [ ] `make build && make test` passes

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- Specific observable outcome (command runs, output file exists, SSH works, etc.)
