# Overview

This actions will handle provision for EKS cluster instances on AWS.

## EKS

It creates / destroy a basic EKS Cluster.

### Operations

#### Create

This will create K8s cluster using EKS according to params specified:

```bash
mapt aws eks create -h
create

Usage:
  mapt aws eks create [flags]

Flags:
      --addons strings               List of EKS addons to be installed, separated by commas.
      --arch string                  architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --conn-details-output string   path to export host connection information (host, username and privateKey)
      --cpus int32                   Number of CPUs for the cloud instance (default 8)
  -h, --help                         help for create
      --load-balancer-controller     Install AWS Load Balancer Controller
      --memory int32                 Amount of RAM for the cloud instance in GiB (default 64)
      --nested-virt                  Use cloud instance that has nested virtualization support
      --spot                         if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-increase-rate int       Percentage to be added on top of the current calculated spot price to increase chances to get the machine (default 20)
      --tags stringToString          tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string               EKS K8s cluster version (default "1.31")
      --workers-desired string       Worker nodes scaling desired size (default "1")
      --workers-max string           Worker nodes scaling maximum size (default "3")
      --workers-min string           Worker nodes scaling minimum size (default "1")

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

It will crete an EKS cluster and kubeconfig file to connect will be placed at `--conn-details-output`. To use the kubeconfig file, you need to have the authentication information exported as variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`).

When running the container image it is required to pass the authentication information as variables, following a sample snippet on how to create
a instance with default values:

```bash
podman run -d --rm \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:v1.0.0-dev aws eks create \
            --project-name "mapt-eks" \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```

It is important to notice that backed-url contains the state of the resources on AWS linked to the EKS creation, to destroy we need to reference them as so we should pass same folder as we set for creation:

```bash
podman run -d --rm \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:v1.0.0-dev aws eks destroy \
            --project-name "mapt-eks" \
            --backed-url file:///workspace
```
