# Overview

This feature of `mapt` allows to setup hosts deployed by it as a GitHub Self Hosted Runner, which can then be directly used for running GitHub Actions jobs.
It benefits from all the existing features that `mapt` already provides, allowing to create self-hosted runners that can be used for different QE scenarios.

## Providers and Platforms

Currently, it allows to create self-hosted runners on:

* AWS: Windows Server, RHEL, Fedora, macOS
* Azure: Windows Desktop, RHEL
* IBM Cloud: IBM Power (ppc64le), IBM Z (s390x)

## Authentication

Registering a Self Hosted Runner requires a short-lived registration token from GitHub. `mapt` supports three ways to supply it. Precedence order: GitHub App credentials → explicit registration token → GITHUB_TOKEN (PAT auto-generate).

### Option A — GitHub App (recommended)

Using a GitHub App avoids long-lived credentials. `mapt` exchanges the App's private key for a short-lived installation access token inside the Pulumi deployment, then uses it to fetch a fresh runner registration token automatically.

**Prerequisites:**

1. Create a GitHub App with the `administration:write` permission on the target repository.
2. Install the App on the target repository and note the **Installation ID**.
3. Generate and download the App's **private key** (`.pem` file).

**Flags:**

| Flag | Description |
|------|-------------|
| `--ghactions-app-id` | GitHub App ID (numeric, shown on the App settings page) |
| `--ghactions-app-installation-id` | Installation ID for the target org or repository |
| `--ghactions-app-private-key` | Path to the App RSA private key PEM file |
| `--ghactions-runner-repo` | Full URL of the repository (mutually exclusive with `--ghactions-runner-org`) |
| `--ghactions-runner-org` | GitHub organization name for an org-level runner (mutually exclusive with `--ghactions-runner-repo`) |
| `--ghactions-runner-labels` | Comma-separated labels to attach to the runner (optional) |

**Example:**

```bash
mapt aws rhel create \
    --ghactions-runner-repo "https://github.com/redhat-developer/mapt" \
    --ghactions-app-id "123456" \
    --ghactions-app-installation-id "789012" \
    --ghactions-app-private-key "/path/to/app-private-key.pem" \
    --project-name mapt-rhel-aws \
    --backed-url file:///workspace/state \
    --conn-details-output /workspace/conn-details
```

### Option B — Personal Access Token (PAT)

Set the `GITHUB_TOKEN` environment variable to a PAT with `repo` admin scope. `mapt` will call the GitHub API to generate a registration token automatically before deployment.

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
mapt aws rhel create \
    --ghactions-runner-repo "https://github.com/redhat-developer/mapt" \
    --project-name mapt-rhel-aws \
    --backed-url file:///workspace/state \
    --conn-details-output /workspace/conn-details
```

### Option C — Pre-generated registration token

Pass a registration token obtained manually from the GitHub API directly via `--ghactions-runner-token`. Tokens expire after 1 hour.

```bash
# Obtain a token via the GitHub API:
curl -L -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer <pat>" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/redhat-developer/mapt/actions/runners/registration-token

mapt azure windows create --spot \
    --ghactions-runner-repo "https://github.com/redhat-developer/mapt" \
    --ghactions-runner-token "ACDZL3QXEIC73UXBDGSEYEI" \
    --project-name mapt-windows-azure \
    --backed-url file:///workspace/state \
    --conn-details-output /workspace/conn-details
```

> **Note:** Additional labels can be added to the runner with `--ghactions-runner-labels`, e.g. `--ghactions-runner-labels="azure,mapt,windows"`.
