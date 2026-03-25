# Spec: [Title]

## Context
Brief background. What area of the codebase this touches. Links to related existing files.

## Problem
What is missing, broken, or needs improvement.

## Requirements
- [ ] Concrete, testable requirement
- [ ] Another requirement

## Out of Scope
Explicit list of what this spec does NOT cover.

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

## Acceptance Criteria
- Specific observable outcome (command runs, test passes, output file exists, etc.)
