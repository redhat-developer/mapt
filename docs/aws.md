# Overview

Create a composable environment with different qe target machines aggregated on different topologies and with specific setups (like vpns, proxys, airgaps,...)

## Hosts

Current available features allow to create supported hosts on AWS using cmd `mapt host create` current supported hosts:

* RHEL (host id `ol-rhel`)
* Fedora (host id `ol-fedora`)
* MacOS M1 v12 (host id `g-macos-m1`)
* Windows_Server-2019-English-Full (host id `ol-windows`)
* Windows_Server-2019-English-Full (host id `ol-windows-non-eng`)
* [SNC runner](https://github.com/crc-org/snc)  (host id `s-snc`)

![Environment](./diagrams/base.svg)

### Spot price use case

This module allows to check for best bid price on all regions, to request instances at lower price to reduce costs. To calculate the best option, it is also required to:  

* reduce interruptions
* ensure capacity

to check those requisites the module make use of spot placement scores based on machine requirements. Then best scores are crossed with lowers price from spot price history to pick the most valuable option.

Current use case is working on one machine but it will be exteded to analyze any required environment offered by mapt (checking with all the machines included on a specific environment).

Current information about supported machines can be checked at [support-matrix](./../pkg/infra/aws/support-matrix/matrix.go)

### Operations

It creates / destroy supported hosts ready to be included within the CI/CD system. Features included within the offering:

```bash
podman run -it --rm quay.io/redhat-developer/mapt:0.7.0-dev aws host create -h
create

Usage:
  mapt aws host create [flags]

Flags:
      --conn-details-output string        path to export host connection information (host, username and privateKey)
      --fedora-major-version string       major version for fedora image 36, 37 (default "37")
  -h, --help                              help for create
      --host-id string                    host id from supported hosts list
      --rh-major-version string           major version for rhel image 7, 8 or 9 (default "8")
      --rh-subscription-password string   password for rhel subcription
      --rh-subscription-username string   username for rhel subcription

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete an instance and will give as result several files located at path defined by `--conn-details-output`:

* username: file containing the username for worker user
* id_rsa: file containing the private key for worker user
* host: file containing the public ip for the instance  

Also it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to detroy the resources.

When running the container image it is required to pass the authetication information as variables(to setup AWS credentials there is a [helper script](./../hacks/aws_setup.sh)), following a sample snipped on how to create an instance with default values:  

```bash
# Create rhel host
# Recommended this region us-east-1
# https://github.com/redhat-developer/mapt/issues/24
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws host create \
            --project-name mapt-rhel \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --host-id ol-rhel \
            --rh-subscription-username ${name} \
            --rh-subscription-password ${passwd} \
```

The following is a snipped on how to destroy the resources:

```bash
# Create rhel host
# Recommended this region us-east-1
# https://github.com/redhat-developer/mapt/issues/24
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws host destroy \
            --project-name mapt-rhel \
            --backed-url file:///workspace 
```
