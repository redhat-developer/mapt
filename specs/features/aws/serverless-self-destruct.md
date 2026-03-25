# Spec: Serverless Self-Destruct (Timeout Mode)

## Context
Any provisioned host or service can optionally schedule its own destruction after a given duration.
This prevents cost overruns when a CI pipeline fails to call `destroy` explicitly.

Implementation: `pkg/provider/aws/modules/serverless/`.

Mechanism:
1. An ECS Fargate task definition is created with the `mapt` OCI image
2. An AWS EventBridge Scheduler one-time schedule fires at `now + timeout`
3. The scheduled task runs `mapt <target> destroy --project-name ... --backed-url ... --serverless`
4. A shared ECS cluster and IAM roles are created once per region and retained (`RetainOnDelete(true)`)

## Problem
This feature is implemented. This spec documents the design and constraints.

## Requirements
- [ ] Accept a timeout duration string (Go `time.Duration` format, e.g. `"4h"`, `"30m"`)
- [ ] Reject timeout when BackedURL is `file://` (state must be remotely accessible by Fargate)
- [ ] Create/reuse a named ECS cluster (`mapt-serverless-cluster`) retained on delete
- [ ] Create/reuse task execution and scheduler IAM roles, retained on delete
- [ ] Create a one-time EventBridge Schedule at `now + timeout` in the region's local timezone
- [ ] The Fargate task image is the `mapt` OCI image baked in at compile time via linker flag (`-X ...context.OCI`)
- [ ] Support `--serverless` flag on destroy to use role-based credentials (no static key/secret needed inside ECS)
- [ ] Clean up the EventBridge schedule and task definition on destroy (these are not retained)

## Out of Scope
- Recurring schedules (used internally by mac-pool HouseKeeper via `serverless.Create()` with `Repeat` type)
- Azure self-destruct (not implemented)

## Affected Areas
- `pkg/provider/aws/modules/serverless/serverless.go` — core implementation
- `pkg/provider/aws/modules/serverless/types.go` — schedule types
- `pkg/manager/context/context.go` — `OCI` variable set by linker
- Any action that calls `serverless.OneTimeDelayedTask()` (rhel, windows, snc, fedora, kind, eks)
- `oci/Containerfile` — the container image being scheduled

## Known Gaps / Improvement Ideas
- IAM policy for the task role is very broad (`ec2:*`, `s3:*`, `cloudformation:*`, `ssm:*`, `scheduler:*`)
  — could be scoped down to only what destroy needs
- There is no mechanism to cancel the scheduled self-destruct once set (other than manually deleting
  the EventBridge schedule from the AWS console)
- The OCI image tag used by the Fargate task is baked in at build time; if a newer binary is deployed
  via a different image tag, old scheduled tasks still run the old image

## Acceptance Criteria
- `mapt aws rhel create --timeout 1h ...` creates a visible EventBridge schedule
- After the timeout, the Fargate task fires and the stack is destroyed
- `mapt aws rhel create --timeout 1h --backed-url file:///tmp/...` returns an error immediately

---

## Command

This is a cross-cutting feature, not a standalone command. It is activated via the
`--timeout` flag on individual target create commands, and the `--serverless` flag
on destroy commands:

```
mapt aws rhel create  --timeout 4h ...
mapt aws rhel destroy --serverless ...
```

Both flags are defined in shared params (`specs/cmd/params.md` — Serverless / Destroy group).
No additional flags are specific to the self-destruct feature itself.
