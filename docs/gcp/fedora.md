# Overview

This action provisions Fedora Cloud virtual machines on Google Cloud Platform.

It creates a Compute Engine instance with a public IP, SSH key pair, and firewall rules allowing SSH access. The machine type is selected automatically based on requested CPU/memory or can be set explicitly.

## Prerequisites

The following environment variables are required:

| Variable | Description |
|---|---|
| `GOOGLE_PROJECT` | GCP project ID |
| `GOOGLE_REGION` | Default GCP region (e.g. `us-central1`) |
| `GOOGLE_APPLICATION_CREDENTIALS` or `GOOGLE_CREDENTIALS` | Path to service account key JSON |

## Create

```bash
mapt gcp fedora create -h
create

Usage:
  mapt gcp fedora create [flags]

Flags:
      --arch string                    architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --conn-details-output string     path to export host connection information (host, username and privateKey)
      --cpus int32                     Number of CPUs for the cloud instance (default 8)
      --disk-size int                  Disk size in GB (default 200)
      --compute-sizes strings          set specific machine type(s) and ignore CPUs/Memory params (e.g. n2-standard-8)
      --compute-families strings       restrict machine type selection to these families (e.g. n2,c3)
      --memory int32                   Amount of RAM for the cloud instance in GiB (default 64)
      --fedora-major-version string    major version of Fedora (default "43")
      --spot                           if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string if spot is enabled we can define the minimum tolerance level of eviction. Allowed values are: lowest, low, medium, high or highest (default "lowest")
      --spot-increase-rate int         Percentage to be added on top of the current calculated spot price to increase chances to get the machine (default 30)
      --spot-excluded-regions strings  Comma-separated list of region IDs to exclude from spot selection
      --tags stringToString            tags to add on each resource (--tags name1=value1,name2=value2) (default [])
  -h, --help                           help for create

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

### Outputs

It will create an instance and produce the following files at the path defined by `--conn-details-output`:

* **host**: public IP of the instance
* **username**: username to connect (`fedora`)
* **id_rsa**: private key for SSH access

It will also create a state folder at the path defined by `--backed-url`. This state is required (with the same `--project-name`) to destroy the resources.

### Container

```bash
podman run -d --name mapt-gcp-fedora \
    -v ${PWD}:/workspace:z \
    -e GOOGLE_PROJECT=my-project \
    -e GOOGLE_REGION=us-central1 \
    -e GOOGLE_CREDENTIALS=/workspace/sa-key.json \
    quay.io/redhat-developer/mapt:0.7.0-dev gcp fedora create \
        --project-name mapt-gcp-fedora \
        --backed-url file:///workspace \
        --conn-details-output /workspace
```

## Destroy

```bash
podman run -d --name mapt-gcp-fedora-destroy \
    -v ${PWD}:/workspace:z \
    -e GOOGLE_PROJECT=my-project \
    -e GOOGLE_REGION=us-central1 \
    -e GOOGLE_CREDENTIALS=/workspace/sa-key.json \
    quay.io/redhat-developer/mapt:0.7.0-dev gcp fedora destroy \
        --project-name mapt-gcp-fedora \
        --backed-url file:///workspace
```
