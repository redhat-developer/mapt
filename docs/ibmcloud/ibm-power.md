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

- **DNS**: the resolvers configured on the target PowerVS subnet are used, with the first two followed by `8.8.8.8` — glibc reads at most three nameservers from `/etc/resolv.conf`, so the public fallback must stay within the first three entries. Nothing is hardcoded per region: the list is read from the subnet at deploy time, so it follows the workspace's own configuration and works for any user or region. `attempts` is reduced from 5 to 3. A subnet that reports no resolvers falls back to runtime host detection, capped the same way. The list is passed to the runner at registration (`--docker-dns`), so job containers resolve through it directly.
- **MTU**: job containers run on a dedicated `mapt-jobs` podman network created at install time with the subnet MTU (`podman network create --opt mtu=...`). The runner's `network_mtu` key is not applied in network mode, so without this a 1500-byte bridge on a smaller-MTU underlay black-holes large transfers while DNS and small requests still work. When the subnet reports no MTU, the default-route interface MTU is used. Setup inspects the applied MTU and recreates the network when it differs. The implicit default `podman` network is not used: podman synthesises it on a fresh store, so it appears to exist before configuration, and it cannot be removed or recreated to change its MTU.
- **Nested subnet**: `executor-containers.conf` uses `default_subnet = 172.20.0.0/24` for bridges created by nested podman, avoiding conflicts with the host's default podman network, the PowerVS subnet, and the routed `10.0.0.0/8` internal range. Setup aborts if any host route or address overlaps that range (CIDR-aware comparison) or if the host network state cannot be inspected. An abort stops the runner install, so the instance comes up without a registered runner; check the cloud-init log if a runner does not appear. Only the main routing table is checked, so routes in policy-routing tables are not considered.
- **Shared job network**: the runner registers with `--docker-network-mode mapt-jobs`, pinning job containers to the dedicated network created above. GitLab Runner only creates per-job networks when the `FF_NETWORK_PER_BUILD` feature flag is enabled, and an explicit network mode takes precedence over it, so jobs stay on the shared network even if that flag is turned on later. The nested subnet and network-mode approach match the live-runner investigation in [issue #894](https://github.com/redhat-developer/mapt/issues/894); the dedicated network name and the MTU handling still need live validation. Because every job now draws its address from one pool, the netavark lease leak tracked in [#909](https://github.com/redhat-developer/mapt/issues/909) reaches the network's address ceiling sooner than it did with per-job networks.

### Validating the networking on a runner

After provisioning a runner (or on an existing one), verify:

```bash
# Networks are as configured: job network (subnet and MTU) and nested subnet
sudo podman network inspect mapt-jobs | grep -iE 'subnet|mtu'
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

## IPAM lease reclamation

The reported failure on RHEL 9.6 with podman 5.4.0 and netavark 1.14.1 is IPAM pool
exhaustion: jobs fail with `IPAM error: failed to find free IP in range`, and the rootful
runroot store `/run/containers/networks/ipam.db` grows across jobs instead of shrinking.
The cause is unverified. Address strings survive successful deallocation because boltdb
keeps freed pages, so neither a growing file nor a stable string count proves a lease
leak, and this timer has not been shown to resolve
[#909](https://github.com/redhat-developer/mapt/issues/909). Treat it as a mitigation for
the exhaustion symptom, not a fix for a proven allocator defect.

Instances created with a GitLab or GitHub runner also get
`/usr/local/sbin/reclaim-ipam.sh`, `reclaim-ipam.service` and `reclaim-ipam.timer`. The
timer runs five minutes after boot and every 10 minutes, and clears the db only when:

- the db holds at least 85% of the addresses in the smallest podman pool, honouring a
  configured lease range, so 215 on a `/24` and 55703 on a `/16`, and
- `podman ps` shows no container in any state other than `exited`, and
- `podman events --stream=false` reports no container created in the last two minutes,
  twice, five seconds apart.

The delete happens while holding podman's rootful networking lock,
`/run/lock/netavark.lock`. podman takes that fcntl lock before it opens the database, so
an allocator either commits its lease before the reclaimer gets the lock, in which case
the fingerprint comparison aborts the run, or waits and opens the recreated file
afterwards. The reclaimer compares a fingerprint of the db before the idle checks and
again while holding the lock.

The number it compares is a pressure heuristic, not a live lease count. It counts distinct
IPv4-looking strings in the file, and boltdb keeps freed pages, so released addresses
still appear. It can therefore reclaim earlier than needed, but it cannot hide a full
pool.

If podman errors, a journald event read fails, the db cannot be read or fingerprinted, the
pool cannot be determined, or the lock cannot be taken within ten seconds, the timer does
nothing. Above the threshold it writes one line to the journal per run under the
`reclaim-ipam` tag; below it, it stays quiet.

Limits worth knowing:

- It only sees the rootful store at `/run/containers/networks/ipam.db`. Rootful podman
  stores on the same host share that database but are queried separately, so a container
  belonging to another store and older than the event window is invisible. Deployment
  assumes a single rootful store: the template checks that a runner is configured, not
  how many stores the host has.
- It needs `python3` for the fcntl lock and the fingerprint. RHEL 9 has it because dnf
  depends on it.
- Allocators that cooperate with podman's networking lock cannot run inside the critical
  section, so they cannot lose a lease to this delete. The lock serializes new
  allocations only: every lease that already exists when the run starts is what the
  delete removes, including one held by a container in another rootful store that
  `podman ps` cannot see. A process that opens the database outside that lock is not
  covered, and is not a supported configuration.

Check the timer and the history:

```bash
sudo systemctl status reclaim-ipam.timer
sudo journalctl -t reclaim-ipam --no-pager -n 20
```

Count the address strings in the db (a pressure indicator, not a lease count):

```bash
sudo bash -c '
db=/run/containers/networks/ipam.db
[ -r "$db" ] || { echo "cannot read $db" >&2; exit 1; }
leases=$(grep -aoE "([0-9]{1,3}\.){3}[0-9]{1,3}" "$db")
rc=$?
[ "$rc" -eq 0 ] || [ "$rc" -eq 1 ] || { echo "grep failed with status $rc" >&2; exit 1; }
printf "%s" "$leases" | sort -u | wc -l
'
```

Exit status 1 from `grep` means the readable db holds no addresses, so the command prints
`0`; any other failure aborts without printing a count.

This is a temporary mitigation. Once a RHEL 9.6 errata ships the netavark fix, delete the
block from the `ibm-power` cloud-config, then remove `reclaim-ipam.sh`, its `//go:embed`
declaration and the `ReclaimIPAMScript` field with its use in `piUserData` in
`ibm-power.go`, and the test references. Finally, on existing instances:

```bash
sudo systemctl disable --now reclaim-ipam.timer
sudo systemctl stop reclaim-ipam.service
```

Disabling stops future activations and stopping ends a running oneshot. Confirm that
leases are released on `podman rm` before trusting the errata.

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