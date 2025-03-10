# Overview

mapt offers several operations to manage environments within azure:

## RHEL

It creates / destroy a rhel ready to be included within the CI/CD system. Features included within the offering:

* Creates an admin user with a self generated passwd (only accessible within rdp)
* Creates a user acting as adminuser
* Setup ssh for the user with a self generated private key

### Operations

#### Create

This will create a RHEL according to params specified:

```bash
podman run -it --rm quay.io/redhat-developer/mapt:0.7.0-dev azure rhel create -h
create

Usage:
  mapt azure rhel create [flags]
Flags:
      --arch string                       architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --conn-details-output string        path to export host connection information (host, username and privateKey)
      --cpus int32                        Number of CPUs for the cloud instance (default 8)
      --ghactions-runner-labels strings   List of labels separated by comma to be added to the self-hosted runner
      --ghactions-runner-name string      Name for the Github Actions Runner
      --ghactions-runner-repo string      Full URL of the repository where the Github Actions Runner should be registered
      --ghactions-runner-token string     Token needed for registering the Github Actions Runner token
  -h, --help                              help for create
      --install-ghactions-runner          Install and setup Github Actions runner in the instance
      --location string                   If spot is passed location will be calculated based on spot results. Otherwise localtion will be used to create resources. (default "West US")
      --memory int32                      Amount of RAM for the cloud instance in GiB (default 64)
      --nested-virt                       Use cloud instance that has nested virtualization support
      --rh-subscription-password string   password to register the subscription
      --rh-subscription-username string   username to register the subscription
      --snc                               if this flag is set the RHEL will be setup with SNC profile. Setting up all requirements to run https://github.com/crc-org/snc
      --spot                              if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string    if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --tags stringToString               tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --username string                   username for general user. SSH accessible + rdp with generated password (default "rhqp")
      --version string                    linux version. Version should be formated as X.Y (Major.minor) (default "9.4")
      --vmsize strings                    set specific size for the VM and ignore any CPUs, Memory and Arch parameters set. Type requires to allow nested virtualization

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete a RHEL instance and will give as result several files located at path defined by `--conn-details-output`:


* username: file containing the username for worker user
* id_rsa: file containing the private key for worker user
* host: file containing the public ip for the instance  

Also, it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to destroy the resources.

When running the container image it is required to pass the authentication information as variables, following a sample snipped on how to create
a instance with default values:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/redhat-developer/mapt:0.7.0-dev azure \
        rhel create \
        --project-name "rhel-94" \
        --backed-url "file:///workspace" \
        --conn-details-output "/workspace" \
        --spot
```

The following is a snipped on how to destroy the resources:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/redhat-developer/mapt:0.7.0-dev azure \
        rhel destroy \
        --project-name "rhel-94" \
        --backed-url "file:///workspace"
```
