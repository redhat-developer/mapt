# Spec: OCI Image Build and Publish

## Status
Implemented

## Context
Builds the `mapt` container image for both `amd64` and `arm64` on every PR and push.
On PR, publishes a multi-arch manifest to `ghcr.io` tagged `:pr-<number>` for downstream
integration testing. On push to `main` or a tag, publishes to `quay.io`.

Relevant files:
- `.github/workflows/build-oci.yaml` — matrix build + push on merge
- `.github/workflows/push-oci-pr.yml` — combines artifacts and publishes PR image to ghcr.io
- `Makefile` — `oci-build-amd64`, `oci-build-arm64`, `oci-save-*`, `oci-load`, `oci-push`
- `oci/Containerfile` — the container image definition

## Problem
This feature is implemented. This spec documents the current behaviour and the two-workflow
design needed to produce a multi-arch manifest from a matrix build.

## Requirements
- [x] Build `amd64` image on `ubuntu-24.04` and `arm64` image on `ubuntu-24.04-arm` in parallel
- [x] Save each image as a `.tar` artifact (`mapt-amd64`, `mapt-arm64`)
- [x] On PR: publish a multi-arch manifest to `ghcr.io/redhat-developer/mapt:pr-<number>`
- [x] On push to `main` or a tag: push both arch images and a multi-arch manifest to `quay.io`
- [x] Install `podman` explicitly on `arm64` runner (not pre-installed)
- [x] PR image publication runs in a separate workflow triggered by `oci-builds` completion
      to work around GitHub Actions artifact cross-workflow access restrictions

## Out of Scope
- Go code build and test (see `code-build.md`)
- Tekton task bundle (see `tkn-bundle.md`)

## Must Reuse
- `make oci-build-amd64` / `make oci-build-arm64` — builds arch-specific image
- `make oci-save-amd64` / `make oci-save-arm64` — saves image to `.tar`
- `make oci-load` — loads both arch tars back into podman
- `make oci-push` — pushes multi-arch manifest to registry
- `redhat-actions/podman-login@v1` — authenticates to quay.io / ghcr.io

## Must Create
- `.github/workflows/build-oci.yaml` — matrix build; push job on `push` events
- `.github/workflows/push-oci-pr.yml` — triggered by `oci-builds` completion; publishes PR image

## API Changes
- none

## Acceptance Criteria

### Unit
<!-- Verifiable without cloud credentials. -->
- Both workflow YAML files are syntactically valid
- `make oci-build-amd64` completes successfully on an amd64 host

### Integration
<!-- Requires real GitHub Actions and registry credentials. -->
- PR to `main` produces `ghcr.io/redhat-developer/mapt:pr-<number>` as a multi-arch manifest
- Push to `main` updates `quay.io/redhat-developer/mapt:main` (amd64 + arm64)
- A semver tag push produces a versioned image on quay.io
