# Overview

mapt provisions virtual machines on Google Cloud Platform using Compute Engine. All GCP targets support both on-demand and spot (preemptible) instances.

## Supported targets

| Target | Command | Default user | Default version |
|---|---|---|---|
| [RHEL](gcp/rhel.md) | `mapt gcp rhel` | `cloud-user` | 9.4 |
| [Fedora](gcp/fedora.md) | `mapt gcp fedora` | `fedora` | 43 |
| [Windows Server](gcp/windows.md) | `mapt gcp windows` | `mapt-user` | 2022 |

## Authentication

The GCP provider requires a service account with at minimum the following IAM roles:

* `roles/compute.admin` — create/delete VMs, networks, firewall rules, IPs
* `roles/iam.serviceAccountUser` — attach the service account to instances

Set the following environment variables:

| Variable | Description |
|---|---|
| `GOOGLE_PROJECT` | GCP project ID |
| `GOOGLE_REGION` | Default region (e.g. `us-central1`). Alternatives: `GCP_REGION` or `CLOUDSDK_COMPUTE_REGION` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account JSON key file |
| `GOOGLE_CREDENTIALS` | Alternative to `GOOGLE_APPLICATION_CREDENTIALS` (auto-bridged) |

## GPU and TPU instances

GPU and TPU instances are provisioned using the compute request flags available on all GCP targets. Use `--compute-sizes` to specify accelerator-attached machine types directly, or use `--gpus` and `--gpu-manufacturer` for automatic selection.

### GPU example

```bash
mapt gcp rhel create \
    --project-name gpu-test \
    --backed-url file:///workspace \
    --conn-details-output /workspace \
    --compute-sizes a2-highgpu-1g
```

### TPU example

```bash
mapt gcp rhel create \
    --project-name tpu-test \
    --backed-url file:///workspace \
    --conn-details-output /workspace \
    --compute-sizes ct5lp-hightpu-4t
```

### Compute request flags

| Flag | Default | Description |
|---|---|---|
| `--cpus` | 8 | Minimum vCPU count |
| `--memory` | 64 | Minimum memory in GiB |
| `--gpus` | 0 | Number of GPUs |
| `--gpu-manufacturer` | (none) | GPU manufacturer (e.g. `NVIDIA`) |
| `--compute-sizes` | (none) | Explicit machine type(s), bypasses CPU/memory selection |
| `--compute-families` | (none) | Restrict to specific families (e.g. `n2,c3,a2`) |
| `--disk-size` | 200 | Disk size in GB |

When no `--compute-sizes` are provided, the selector lists all machine types in the target zone, filters by CPU/memory requirements, and returns the smallest matching types.

## Spot instances

When `--spot` is passed, mapt uses the GCP Capacity Advice API (beta) to automatically select the best zone and machine type across all available regions. The selection pipeline:

1. **Obtainability** — queries `advice.capacity` for each region/machine-type combination and filters by an obtainability score threshold based on the tolerance level
2. **Exhaustion cache** — skips zone/type combinations that recently failed with capacity errors (cached for 15 minutes)
3. **Preemption rate** — queries `advice.capacityHistory` and filters candidates exceeding the preemption rate threshold
4. **Pricing** — enriches candidates with spot price data from the history API
5. **Ranking** — scores candidates by obtainability, price, region tier, and machine family

### Spot flags

| Flag | Default | Description |
|---|---|---|
| `--spot` | false | Enable spot instance selection |
| `--spot-eviction-tolerance` | `lowest` | Minimum tolerance for eviction risk. Values: `lowest`, `low`, `medium`, `high`, `highest` |
| `--spot-increase-rate` | 30 | Percentage added to calculated spot price to improve success rate |
| `--spot-excluded-regions` | (none) | Comma-separated list of regions to exclude from spot search |

## Known issues

### Spot and accelerator capacity exhaustion

GCP spot instances — especially GPU and TPU machine types — are frequently unavailable due to high demand. The `advice.capacity` API may return low obtainability scores across all regions, causing spot selection to fail with:

```
no zones meet the obtainability threshold for tolerance N
```

**Workaround**: increase the tolerance level with `--spot-eviction-tolerance medium` or `high` to accept zones with lower availability guarantees. Alternatively, provide multiple machine types via `--compute-sizes` to broaden the search, or drop the `--spot` flag to use on-demand instances.

### Capacity Advice API (beta)

The `advice.capacity` and `advice.capacityHistory` APIs are GCP beta features. If these APIs are not enabled or fail, spot selection falls back to a zone-based approach that iterates regions and zones looking for available machine types without obtainability scoring.

To use the beta API, ensure the Compute Engine API is enabled in your project and the service account has `compute.advice.*` permissions.

### Windows instances — no SSH readiness check

GCP Windows instances do not perform an SSH readiness check at creation time. The instance may take several minutes after creation before RDP and SSH are available while the startup script configures the user account.

### Fedora images — x86_64 only

GCP Fedora Cloud images (`fedora-cloud` project) only provide `x86_64` images. The `--arch` flag is accepted but `arm64` images are not available from the upstream image family.
