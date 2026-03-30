# Spec: Spec-Driven PR Workflow

## Status
Draft

## Jira
<!-- link to tracking issue once created -->

## Context
mapt is adopting a spec-anchored development approach (see `specs/project-context.md`).
Today, PRs mix spec and implementation in the same review round, or skip the spec entirely.
There is no CI gate that enforces a spec exists before code is merged, and no structured
signal to trigger an AI agent to implement from a spec.

Current CI workflows for reference: `specs/features/cicd/code-build.md`,
`specs/features/cicd/oci-build.md`.

## Problem
- Spec review and code review happen in the same PR, collapsing the gate that makes
  spec-first valuable. Reviewers must context-switch between architectural intent and
  implementation detail in a single pass.
- An AI agent implementing from a spec has no well-defined trigger or input convention.
- There is no CI check that prevents a malformed or `Draft` spec from being merged as
  if it were accepted.

## Requirements
- [ ] A developer (or agent) opens a **Draft PR** containing only a spec file under
      `specs/features/` with `Status: Accepted`
- [ ] CI runs `spec-lint` on the PR: validates that every changed spec file has all
      required sections and that `Status` is not `Draft`
- [ ] `spec-lint` fails the PR if any required section is missing or `Status == Draft`
- [ ] A reviewer approves the spec by posting a `/implement` comment on the PR
- [ ] The `/implement` comment triggers a GitHub Actions workflow that runs a Claude Code
      agent, which reads the spec and adds implementation commit(s) to the same branch
- [ ] The agent commits with message `feat(<target>): implement from specs/<spec-file>`
- [ ] After the agent commit, `code-build` re-runs automatically; the PR is promoted from
      Draft to Ready for Review
- [ ] A second review round covers only the implementation (spec already approved)
- [ ] On merge, existing workflows (`code-build`, `oci-builds`, `tkn-bundle`) run unchanged

## Out of Scope
- Changes to `code-build.md`, `oci-build.md`, or `tkn-bundle.md` workflows
- Jira auto-creation of spec stubs from issues (follow-on)
- Two-PR flow (spec merged separately before implementation PR) — possible future evolution
- Automated integration tests triggered by merge (separate spec)

## Design

### PR Lifecycle

```
1. Dev opens Draft PR
   branch: feat/aws-xyz-host
   commit: "spec: aws xyz host"  ← specs/features/aws/xyz-host.md (Status: Accepted)

2. CI: spec-lint
   ✓ required sections present
   ✓ Status == Accepted (not Draft)
   ✓ Must Reuse references valid specs/api/ paths
   ✗ fails → PR blocked, dev fixes spec

3. Reviewer reads spec only
   → posts /implement comment

4. CI: implement workflow triggers
   → Claude Code agent runs with:
        constitution:  specs/project-context.md
        api context:   all specs/api/ files referenced in Must Reuse
        task:          implement all files in Must Create, calling Must Reuse modules
   → agent pushes commit(s) to the PR branch

5. CI re-runs: make build && make test
   PR auto-promoted Draft → Ready for Review

6. Reviewer does code review (implementation only)
   → merge
```

### spec-lint Rules

| Rule | Check |
|---|---|
| Required sections | `## Status`, `## Context`, `## Problem`, `## Requirements`, `## Must Reuse`, `## Must Create`, `## Acceptance Criteria` all present |
| Status not Draft | Value is `Accepted`, `Implemented`, or `Deprecated` |
| Must Reuse not empty | Section body has at least one bullet point |
| Must Create not empty | Section body has at least one file path |

### `/implement` Trigger

A `issue_comment` GitHub Actions event listens for comments containing `/implement` on PRs.
Before dispatching the agent, the workflow verifies the commenter has `write` permission on
the repository. Either `/implement` comment or a `spec-approved` label can trigger the agent.

### Agent Context

The agent receives:
1. The changed spec file (from the PR diff)
2. `specs/project-context.md` (mandatory module sequences, naming rules)
3. All `specs/api/` files referenced in Must Reuse
4. Read-only access to the existing codebase

The agent is constrained to create only the files listed in Must Create, call only the
modules listed in Must Reuse in the documented order, and verify `make build` passes
before committing.

## Must Reuse

Existing workflows that must **not** be modified:
- `.github/workflows/build-go.yaml` — `code-build`
- `.github/workflows/build-oci.yaml` — `oci-builds`
- `.github/workflows/tkn-bundle.yaml` — `tkn-bundle`

## Must Create

| File | Purpose |
|---|---|
| `.github/workflows/spec-lint.yaml` | Runs `scripts/spec-lint.sh` on PR; blocks merge if spec is malformed or Draft |
| `.github/workflows/spec-implement.yaml` | Listens for `/implement` comment; verifies write access; dispatches agent |
| `scripts/spec-lint.sh` | Shell script: checks required sections, Status value, non-empty Must Reuse/Must Create |

## API Changes
- none

## Tasks
- [ ] Write `scripts/spec-lint.sh` — section presence, Status != Draft, non-empty sections
- [ ] Test `spec-lint.sh` locally against all existing specs (all should pass)
- [ ] Write `.github/workflows/spec-lint.yaml` — triggers on PR, runs lint against changed `specs/features/**/*.md`
- [ ] Write `.github/workflows/spec-implement.yaml` — `issue_comment` trigger, write-access guard, agent dispatch
- [ ] Define agent invocation: model selection, system prompt assembly, output commit convention
- [ ] Add `spec-approved` label to the GitHub repository
- [ ] Update `specs/features/cicd/` with the two new workflow specs once implemented
- [ ] `make build && make test` passes (no Go changes expected)

## Acceptance Criteria

### Unit
<!-- Verifiable without cloud credentials. -->
- `scripts/spec-lint.sh specs/features/aws/rhel-host.md` exits 0
- `scripts/spec-lint.sh specs/features/000-template.md` exits non-zero (Status is `Draft`)
- `scripts/spec-lint.sh` exits non-zero on a spec missing the `## Must Reuse` section

### Integration
<!-- Requires a real GitHub repository with Actions enabled. -->
- A Draft PR with `Status: Draft` causes `spec-lint` to fail and block merge
- A Draft PR with `Status: Accepted` and all required sections causes `spec-lint` to pass
- Posting `/implement` on a passing-spec PR triggers the agent workflow
- Agent commit appears on the branch; `make build && make test` passes
- PR is promoted from Draft to Ready for Review automatically after the agent commit
