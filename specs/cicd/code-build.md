# Spec: Go Code Build and Test

## Status
Implemented

## Context
Runs static analysis, build, and unit tests on every PR and push to `main`.
This is the primary gate for Go code correctness.

Relevant files:
- `.github/workflows/build-go.yaml`
- `Makefile` — `check`, `build`, `test`, `lint`, `fmt` targets

## Problem
This feature is implemented. This spec documents the current behaviour.

## Requirements
- [x] Run on every `pull_request` targeting `main` and every `push` to `main` or a tag
- [x] Run `make check` (build + test + lint + renovate-check) on `ubuntu-24.04`
- [x] Pin Go version (`1.26`)
- [x] Free disk space before build to avoid Docker layer cache exhaustion

## Out of Scope
- OCI image build (see `oci-build.md`)
- Integration tests on non-Linux hosts (see `hosted-runner-test.md`)

## Must Reuse
- `make check` — runs `make build`, `make test`, `make lint`, `make renovate-check` in sequence
- `endersonmenezes/free-disk-space@v3` — frees Android/dotnet/Haskell toolchains before build

## Must Create
- `.github/workflows/build-go.yaml`

## API Changes
- none

## Acceptance Criteria

### Unit
<!-- Verifiable without cloud credentials. -->
- Workflow YAML is syntactically valid
- `make check` passes locally on a clean checkout

### Integration
<!-- Requires a real GitHub Actions run. -->
- PR to `main` triggers the workflow and it passes
- Push to `main` triggers the workflow and it passes
- A PR introducing a lint error causes the workflow to fail
