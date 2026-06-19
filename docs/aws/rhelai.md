# Overview

mapt offers operations to manage RHEL AI environments on AWS. RHEL AI instances are GPU-enabled machines with pre-installed RHEL AI images, suitable for AI/ML workloads.

## Operations

### List Versions

List available RHEL AI versions for a given accelerator type:

```bash
mapt aws rhel-ai list-versions -h
list-versions

Usage:
  mapt aws rhel-ai list-versions [flags]

Flags:
      --accelerator string   accelerator type. Valid types: cuda and rocm (default "cuda")
  -h, --help                 help for list-versions

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

#### Container

```bash
podman run -it --rm \
    -e AWS_ACCESS_KEY_ID=XXX \
    -e AWS_SECRET_ACCESS_KEY=XXX \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:0.7.0-dev aws \
        rhel-ai list-versions \
        --accelerator cuda
```

### Create

This will create a RHEL AI instance according to params specified:

```bash
mapt aws rhel-ai create -h
create

Usage:
  mapt aws rhel-ai create [flags]

Flags:
      --accelerator string         accelerator type. Valid types: cuda and rocm (default "cuda")
      --conn-details-output string path to export host connection information (host, username and privateKey)
      --cpus int32                 Number of CPUs for the cloud instance (default 8)
      --custom-image string        custom AMI name (overrides version and accelerator)
      --disk-size int              Disk size in GB (default 2000)
      --gpus int32                 Number of GPUs
      --memory int32               Amount of RAM for the cloud instance in GiB (default 64)
      --spot                       if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string  if spot is enabled we can define the minimum tolerance level of eviction. Allowed values are: lowest, low, medium, high or highest (default "lowest")
      --tags stringToString        tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --timeout string             set a timeout for the instance (e.g. 4h)
      --version string             version for the RHELAI OS (default "3.0.0")
  -h, --help                       help for create

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

#### Outputs

It will create a RHEL AI instance and will give as result several files located at path defined by `--conn-details-output`:

* **host**: host for the instance (load balancer DNS if spot)
* **username**: username to connect to the machine
* **id_rsa**: private key to connect to the machine

Also, it will create a state folder holding the state for the created resources at AWS, the path for this folder is defined within `--backed-url`, the content from that folder is required with the same project name (`--project-name`) in order to destroy the resources.

#### Container

When running the container image it is required to pass the authentication information as variables, following a sample snippet on how to create an instance with default values:

```bash
podman run -d --name mapt-rhelai \
    -v ${PWD}:/workspace:z \
    -e AWS_ACCESS_KEY_ID=XXX \
    -e AWS_SECRET_ACCESS_KEY=XXX \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:0.7.0-dev aws \
        rhel-ai create \
        --project-name mapt-rhelai \
        --backed-url file:///workspace \
        --conn-details-output /workspace \
        --spot
```

### Destroy

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e AWS_ACCESS_KEY_ID=XXX \
    -e AWS_SECRET_ACCESS_KEY=XXX \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:0.7.0-dev aws \
        rhel-ai destroy \
        --project-name mapt-rhelai \
        --backed-url file:///workspace
```
