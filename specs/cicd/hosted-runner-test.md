# Spec: Windows Integration Test via Self-Hosted Runner

## Status
Implemented

## Context
Runs the full Go test suite on a real Windows host provisioned by mapt itself. This is
mapt's only integration-level CI gate: the tool provisions its own test environment and
then runs tests inside it. Triggered only when the GitHub Actions runner integration code
changes (`pkg/integrations/github/`).

Relevant files:
- `.github/workflows/build-img-ghrunner-test.yaml` — builds a dedicated OCI image for the test
- `.github/workflows/build-on-hosted-runner.yaml` — orchestrates provision → test → destroy
- `.github/workflows/provision-hosted-runner.yaml` — reusable: fetches runner token, runs mapt create
- `.github/workflows/destroy-hosted-runner.yaml` — reusable: runs mapt destroy

## Problem
This feature is implemented. This spec documents the current behaviour, the four-workflow
design, and the always-destroy guarantee.

## Requirements
- [x] Trigger only on PRs to `main` that change `pkg/integrations/github/*.go` or
      `.github/workflows/build-img-ghrunner-test.yaml` (path filter)
- [x] Build a dedicated OCI image tagged `:pr-<number>` for the test run (separate from
      the standard PR image)
- [x] Fetch a GitHub Actions runner registration token via the GitHub API before provisioning
- [x] Provision an Azure Windows VM with the GitHub runner pre-installed using mapt,
      authenticated via ARM_* secrets and Azure Blob Storage for Pulumi state
- [x] Wait 120 seconds after provisioning for the runner to register with GitHub
- [x] Run `go test -v ./...` on the self-hosted Windows runner
- [x] Destroy the Azure VM after the test regardless of outcome (`if: always()`)
- [x] Pulumi state backend: `azblob://mapt-gh-runner-mapt-state/<repo>-<run-id>`

## Out of Scope
- Integration tests for AWS targets (not currently automated)
- Integration tests on Linux self-hosted runners
- Testing non-GitHub runner integrations (Cirrus, GitLab) in CI

## Design
Four workflows are needed because GitHub Actions does not allow a single job to both
provision a self-hosted runner and then use it (the runner must be registered before
the job that runs on it is dispatched). The split is:

```
build-img-ghrunner-test   builds OCI image → artifact
        │ (workflow_run trigger)
build-on-hosted-runner    orchestrates:
  ├── provision-hosted-runner (reusable)
  │     fetch token → download artifact → mapt create → sleep 120s
  ├── test_run_selfhosted_runner   [self-hosted, x64, Windows]
  │     go test -v ./...
  └── destroy-hosted-runner (reusable)     if: always()
        download artifact → mapt destroy
```

The `destroy-hosted-runner` job runs with `if: always()` and depends on both
`hosted_runner_provision` and `test_run_selfhosted_runner`, ensuring the VM is
destroyed even when tests fail or the provision job partially succeeds.

## Must Reuse
- `mapt azure windows create` — provisions the Azure Windows VM with `--install-ghactions-runner`
- `mapt azure windows destroy` — tears down the VM and cleans up Pulumi state
- `make oci-build-amd64` / `make oci-save-amd64` — builds and saves the test image
- GitHub runner registration token API: `POST /repos/{owner}/{repo}/actions/runners/registration-token`

## Must Create
- `.github/workflows/build-img-ghrunner-test.yaml` — path-gated build; uploads artifact
- `.github/workflows/build-on-hosted-runner.yaml` — orchestration workflow
- `.github/workflows/provision-hosted-runner.yaml` — reusable provision workflow
- `.github/workflows/destroy-hosted-runner.yaml` — reusable destroy workflow

## API Changes
- none

## Known Gaps
- The 120s sleep is a fixed wait; there is no poll-until-ready mechanism
- Only `amd64` is tested; `arm64` Windows is not covered
- Path filter means changes to non-integration Go code do not trigger Windows tests
- `destroy-hosted-runner` downloads artifact by name `mapt` (singular) but
  `build-img-ghrunner-test` uploads as `mapt-<arch>` — verify names are consistent

## Acceptance Criteria

### Unit
<!-- Verifiable without cloud credentials. -->
- All four workflow YAML files are syntactically valid
- The `if: always()` condition on the destroy job is present

### Integration
<!-- Requires real GitHub Actions, Azure credentials, and a GitHub PAT. -->
- A PR changing `pkg/integrations/github/*.go` triggers the full pipeline
- The self-hosted Windows runner appears in the repository runner list during the test
- `go test -v ./...` passes on the Windows runner
- The Azure VM is destroyed after the run (both on success and failure)
- A PR not touching `pkg/integrations/github/` does not trigger this pipeline
