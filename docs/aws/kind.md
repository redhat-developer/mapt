# Overview

This action will provision an AWS instance running a [Kind][kind] cluster. It handles the full lifecycle of the underlying EC2 instance and exposes the cluster via a kubeconfig file for immediate use.

## Create

```bash
mapt aws kind create -h
create

Usage:
  mapt aws kind create [flags]

Flags:
      --arch string                      architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --compute-families strings         Comma-separated allowlist of compute family prefixes (e.g. m5,m6i,m7i for AWS; D8v3,E16v4 for Azure). Empty means no restriction. Only used when --compute-sizes is not set.
      --compute-sizes strings            Comma seperated list of sizes for the machines to be requested. If set this takes precedence over compute by args
      --conn-details-output string       path to export host connection information (host, username and privateKey)
      --cpus int32                       Number of CPUs for the cloud instance (default 8)
      --disk-size int                    Disk size in GB for the cloud instance (default 200)
      --extra-port-mappings string       Additional port mappings for the Kind cluster. Value should be a JSON array of objects with containerPort, hostPort, and protocol properties. Example: '[{"containerPort": 8080, "hostPort": 8080, "protocol": "TCP"}]'
      --gpu-manufacturer string          Manufacturer company name for GPU. (i.e. NVIDIA)
      --gpus int32                       Number of GPUs for the cloud instance
  -h, --help                             help for create
      --memory int32                     Amount of RAM for the cloud instance in GiB (default 64)
      --nested-virt                      Use cloud instance that has nested virtualization support
      --service-endpoints strings        Comma-separated list of VPC endpoints to create. Accepted values: s3, ecr, ssm. Empty = no endpoints.
      --spot                             if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string   if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --spot-excluded-regions strings    Comma-separated list of zone IDs to exclude from spot selection
      --spot-increase-rate int           Percentage to be added on top of the current calculated spot price to increase chances to get the machine (default 30)
      --tags stringToString              tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --timeout string                   if timeout is set a serverless destroy actions will be set on the time according to the timeout. The Timeout value is a duration conforming to Go ParseDuration format.
      --version string                   version for k8s offered through Kind. (default "v1.35")
      --vpc-id string                    ID of an existing VPC to deploy the instance into. When set, airgap is not supported and spot search is restricted to AZs with public subnets in that VPC.

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer.
      --debug                 Enable debug traces and set verbosity to max.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Outputs

- It will create an instance running a Kind cluster and will give as result several files located at the path defined by `--conn-details-output`:
  - **host**: public host for the instance (load balancer DNS when spot is enabled)
  - **username**: username to connect to the machine
  - **id_rsa**: private key to connect to the machine
  - **kubeconfig**: kubeconfig file pre-configured to access the Kind cluster remotely

- Also, it will create a state folder holding the state for the created resources on AWS; the path for this folder is defined within `--backed-url`. The content from that folder is required with the same project name (`--project-name`) in order to destroy the resources.

### Container

When running the container image it is required to pass the authentication information as environment variables (to setup AWS credentials there is a [helper script](./../../hacks/aws/aws_setup.sh)). Following is a sample snippet on how to create a Kind cluster with default values:

```bash
podman run -d --name mapt-kind \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:v1.0.0 aws kind create \
            --project-name mapt-kind \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```

To bring your own VPC, ensure the selected AZ has a public subnet whose associated route table has a default route to an internet gateway:

```bash
podman run -d --name mapt-kind \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:v1.0.0 aws kind create \
            --project-name mapt-kind \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --vpc-id vpc-XXX
```

## Destroy

```bash
mapt aws kind destroy -h
destroy

Usage:
  mapt aws kind destroy [flags]

Flags:
      --force-destroy   force destroy even if the stack is in a failed state
  -h, --help            help for destroy
      --keep-state      destroy cloud resources but keep the Pulumi state
      --serverless      trigger serverless teardown

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer.
      --debug                 Enable debug traces and set verbosity to max.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Container

```bash
podman run -d --name mapt-kind \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:v1.0.0 aws kind destroy \
            --project-name mapt-kind \
            --backed-url file:///workspace
```

[kind]: https://kind.sigs.k8s.io/
