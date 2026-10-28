# Overview

This action provisions **two instances** when `--vpc-public-subnet-id` is set, or **one instance** otherwise:

1. **PowerVS instance** (always) — RHEL 9 ppc64le on an s1022 (Power10) frame, attached to the existing private subnet inside the workspace. This instance has a private IP only and is the build target.
2. **VPC bastion** (optional, when `--vpc-public-subnet-id` is set) — a small Ubuntu 24.04 x86_64 VPC instance (`cx2-2x4`) with a floating IP, used as an SSH jump host to reach the PowerVS instance over the Transit Gateway private network.

The workspace must use the **Transit Gateway** networking model.

On first boot, cloud-init automatically configures the PowerVS instance for on-prem network access:
- Adds a persistent `10.0.0.0/8` route via the subnet gateway (required for Transit Gateway on-prem routing)
- DNS resolvers are picked up from the subnet configuration via DHCP

## Prerequisites

- An existing PowerVS workspace with Transit Gateway networking
- An existing private subnet within that workspace (`--pi-private-subnet-id`)
- _(Optional)_ An existing VPC subnet with a public gateway, connected to the Transit Gateway (`--vpc-public-subnet-id`) — required for SSH access since PowerVS instances are on a private network only

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `IBMCLOUD_ACCOUNT` | yes | IBM Cloud account ID |
| `IBMCLOUD_API_KEY` | yes | IBM Cloud API key |
| `IC_REGION` | yes | IBM Cloud region (e.g. `us-south`, `us-east`) |
| `IBMCLOUD_COS_ACCESS_KEY_ID` | only with S3 `--backed-url` | HMAC access key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_SECRET_ACCESS_KEY` | only with S3 `--backed-url` | HMAC secret key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_ENDPOINT` | no | COS S3 endpoint (defaults to `s3.<region>.cloud-object-storage.appdomain.cloud`) |

## Create

```bash
mapt ibmcloud ibm-power create -h
create

Usage:
  mapt ibmcloud ibm-power create [flags]

Flags:
      --conn-details-output string           path to export host connection information (host, username and privateKey)
      --ghactions-runner-labels strings      List of labels separated by comma to be added to the self-hosted runner
      --ghactions-runner-repo string         Full URL of the repository where the Github Actions Runner should be registered
      --ghactions-runner-token string        Token needed for registering the Github Actions Runner token
  -h, --help                                 help for create
      --otel-app-code string                 OpenTelemetry appcode identifier (e.g. MAPT-001); when set together with --otel-auth-token, installs the otelcol-contrib filelog collector on the instance
      --otel-auth-token string               OpenTelemetry authentication token (UUID) used to authenticate against the OTLP endpoint
      --otel-endpoint string                 OTLP HTTP endpoint to export logs to (default "https://otel-input.corp.redhat.com")
      --pi-private-subnet-id string          ID of an existing Power VS private subnet to attach the instance to
      --tags stringToString                  tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --vpc-public-subnet-id string          ID of an existing VPC subnet (with public gateway, connected to Transit Gateway) for the SSH bastion
      --workspace-id string                  ID of an existing Power VS workspace (cloud instance)

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Outputs

Files written to the path defined by `--conn-details-output`:

| File | Description |
|---|---|
| `host` | Private IP of the PowerVS instance |
| `username` | SSH username (`root`) |
| `id_rsa` | Private key for the PowerVS instance |
| `bastion_host` | Floating IP of the VPC bastion _(only when `--vpc-public-subnet-id` is set)_ |
| `bastion_username` | SSH username for the bastion (`ubuntu`) _(only when `--vpc-public-subnet-id` is set)_ |
| `bastion_id_rsa` | Private key for the bastion _(only when `--vpc-public-subnet-id` is set)_ |

A state folder is also created at `--backed-url`. It is required (together with `--project-name`) to destroy the resources later.

### SSH access

When `--vpc-public-subnet-id` is set, use the bastion as a jump host:

```bash
OUTPUT=/path/to/conn-details-output

ssh -i ${OUTPUT}/id_rsa \
    -o StrictHostKeyChecking=no \
    -o ProxyCommand="ssh -i ${OUTPUT}/bastion_id_rsa -o StrictHostKeyChecking=no -W %h:%p ubuntu@$(cat ${OUTPUT}/bastion_host)" \
    root@$(cat ${OUTPUT}/host)
```

### Container

```bash
# With VPC bastion for SSH access (recommended for Transit Gateway workspaces)
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power create \
            --project-name ibm-power \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --workspace-id <workspace-id> \
            --pi-private-subnet-id <private-subnet-id> \
            --vpc-public-subnet-id <vpc-subnet-id>

# Without bastion (instance is on private network only, requires other means of access)
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power create \
            --project-name ibm-power \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --workspace-id <workspace-id> \
            --pi-private-subnet-id <private-subnet-id>
```

## OpenTelemetry log collection

When both `--otel-app-code` and `--otel-auth-token` are provided, cloud-init installs `otelcol-contrib` on the PowerVS instance at first boot and configures it to ship `/var/log/messages` and `/var/log/secure` to the OTLP endpoint.

```bash
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power create \
            --project-name ibm-power \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --workspace-id <workspace-id> \
            --pi-private-subnet-id <private-subnet-id> \
            --vpc-public-subnet-id <vpc-subnet-id> \
            --otel-app-code MAPT-001 \
            --otel-auth-token <uuid-token>
```

## GitLab runner podman networking

When a GitLab runner is provisioned on the PowerVS instance (`--glrunner-*` flags), cloud-init configures podman networking for nested-container CI jobs — jobs that run `podman` inside the job container. These settings apply to IBM Power targets only; every other target keeps the generic setup.

- **DNS**: the first two of the region's IBM Cloud resolvers (selected from `IC_REGION`) followed by `8.8.8.8` — glibc reads at most three nameservers from `/etc/resolv.conf`, so the public fallback must stay within the first three entries. `attempts` is reduced from 5 to 3. Regions without a resolver table fall back to runtime host detection, capped the same way. The list is passed to the runner at registration (`--docker-dns`), so job containers resolve through it directly.
- **Nested subnet**: `executor-containers.conf` uses `default_subnet = 172.20.0.0/24` for bridges created by nested podman, avoiding conflicts with the host's default podman network, the PowerVS subnet, and the routed `10.0.0.0/8` internal range. Setup aborts if any host route or address overlaps that range (CIDR-aware comparison) or if the host network state cannot be inspected.
- **Shared job network**: the runner registers with `--docker-network-mode podman`, so all job containers share the host's default podman network instead of the runner creating per-job bridges. With a shared network the runner never removes the network or other jobs' containers during cleanup, which is safe with `concurrent > 1`.

Regional resolver table (from the live-runner investigation in [issue #894](https://github.com/redhat-developer/mapt/issues/894)):

| `IC_REGION` | Resolvers |
|---|---|
| `us-south` | 10.130.98.132, 10.130.98.5, 10.130.98.69 |
| `us-east` | 10.130.64.4, 10.130.64.70, 10.130.64.134 |
| `eu-de` | 10.130.128.5, 10.130.128.69, 10.130.128.133 |
| `eu-gb` | 10.130.129.4, 10.130.129.68, 10.130.129.132 |

### Validating the networking on a runner

After provisioning a runner (or on an existing one), verify:

```bash
# Networks are as configured: host default network and nested subnet
sudo podman network inspect podman | grep -i subnet
grep default_subnet /etc/containers/executor-containers.conf

# DNS works in both container layers: host container, then nested container
podman run --rm fedora:latest getent hosts gitlab.com
podman run --rm --privileged \
    -v /etc/containers/executor-containers.conf:/etc/containers/containers.conf:ro \
    fedora:latest podman run --rm fedora:latest getent hosts gitlab.com

# MTU: host interface vs podman bridge — a mismatch black-holes large packets
cat /sys/class/net/$(ip route show default | awk '/default/ {print $5; exit}')/mtu
cat /sys/class/net/podman0/mtu

# Image pulls (large transfers) and repeated pairs of concurrent CI jobs
# (concurrent = 2) must both succeed, including nested podman workloads
sudo podman pull registry.access.redhat.com/ubi9:latest
```

## Using IBM Cloud Object Storage as S3 backend

To store Pulumi state in IBM COS instead of a local file, create [HMAC credentials](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-uhc-hmac-credentials-main) for your COS instance and pass an `s3://` backed URL:

```bash
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-south \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power create \
            --project-name ibm-power \
            --backed-url s3://my-cos-bucket \
            --conn-details-output /workspace \
            --workspace-id <workspace-id> \
            --pi-private-subnet-id <private-subnet-id>
```

An HTTPS endpoint URL is also supported as `--backed-url`, with the bucket name in the path:

```
--backed-url https://s3.us-south.cloud-object-storage.appdomain.cloud/my-cos-bucket
```

The COS endpoint and `PULUMI_BACKEND_URL` are constructed automatically from the region and bucket name.

## Destroy

```bash
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power destroy \
            --project-name ibm-power \
            --backed-url file:///workspace
```

By default, destroy removes the Pulumi state files from the backend after a successful destroy. Use `--keep-state` to preserve them:

```bash
podman run -d --name ibm-power \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-south \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-power destroy \
            --project-name ibm-power \
            --backed-url s3://my-cos-bucket \
            --keep-state
```