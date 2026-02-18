# Overview

This feature of `mapt` allows you to set up hosts deployed by it as GitLab Runners, which can then be directly used for running GitLab CI/CD jobs.
It benefits from all the existing features that `mapt` already provides, allowing you to create self-hosted runners that can be used for different QE scenarios.

## Providers and Platforms

Currently, it allows you to create self-hosted GitLab runners on AWS (Fedora, RHEL, Windows Server, macOS) and Azure (RHEL, Windows Desktop).

### Prerequisite

To register a GitLab Runner, you need an access token with `create_runner` and `manage_runner` scopes. There are two options:

#### Option 1: Project Access Token (Recommended)

Project Access Tokens are scoped to a specific project, making them more secure than personal tokens.

**Requirements:**
* GitLab Premium/Ultimate (on GitLab.com)
* OR self-managed GitLab instance (version 13+)

**Steps:**
1. Go to your project → **Settings** → **Access Tokens**
2. Create a new token with:
   - **Token name**: `mapt-runner`
   - **Role**: `Maintainer` or `Owner`
   - **Scopes**: Select `create_runner` and `manage_runner`
   - **Expiration**: Set according to your needs
3. Click **Create project access token**
4. Copy the token immediately

#### Option 2: Personal Access Token (Fallback)

If Project Access Tokens are not available, use a Personal Access Token:

1. Go to **User Settings** → **Access Tokens**
2. Create a new token with:
   - **Token name**: `mapt-runner-token`
   - **Scopes**: Select `create_runner` and `manage_runner`
   - **Expiration**: Set according to your needs
3. Click **Create personal access token**
4. Copy the token immediately

#### Getting Project or Group ID

You also need either a Project ID or Group ID:

* **Project ID**: Go to your project → **Settings** → **General** (displayed at the top)
* **Group ID**: Go to your group → **Settings** → **General** (displayed at the top)

> *NOTE:* For Group Runners with Group Access Tokens, you'll need GitLab Premium/Ultimate on GitLab.com or a self-managed instance.

### Operations

After obtaining the token and the project/group ID, you can deploy a runner using the appropriate flags.

**Required flags:**
* `--glrunner-token` - Your GitLab access token (project or personal)
* `--glrunner-url` - GitLab instance URL (e.g., `https://gitlab.com`)
* `--glrunner-project-id` OR `--glrunner-group-id` - Either project or group ID (not both)

**Optional flags:**
* `--glrunner-tags` - Comma-separated tags (e.g., `aws,mapt,fedora`)

To deploy a Fedora runner on AWS as a project runner:

```bash
mapt aws fedora create --spot \
    --glrunner-token="glpat-xxxxxxxxxxxxxxxxxxxx" \
    --glrunner-url="https://gitlab.com" \
    --glrunner-project-id="12345678" \
    --glrunner-tags="aws,fedora,mapt" \
    --project-name mapt-fedora-aws \
    --backed-url file:///Users/tester/workspace \
    --conn-details-output /Users/tester/workspace/conn-details
```

To deploy a RHEL runner on Azure as a group runner:

```bash
mapt azure rhel create --spot \
    --glrunner-token="glpat-xxxxxxxxxxxxxxxxxxxx" \
    --glrunner-url="https://gitlab.com" \
    --glrunner-group-id="87654321" \
    --glrunner-tags="azure,rhel,mapt" \
    --project-name mapt-rhel-azure \
    --backed-url file:///Users/tester/workspace \
    --conn-details-output /Users/tester/workspace/conn-details
```

> *NOTE:* If no tags are provided with `--glrunner-tags`, the runner can run untagged jobs. If tags are provided, the runner only runs jobs matching those tags.
