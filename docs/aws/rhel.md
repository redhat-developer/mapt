# Overview

This actions will handle provision RHELServer machines on dedicated hosts. This is a requisite to run nested virtualization on AWS.

As a special case it offers the `profileSNC` option if that flag is set the instance will setup all requirements to run 
[single node cluster builder process](https://github.com/crc-org/snc)  

## Create

```bash
mapt aws rhel create -h
create

Usage:
  mapt aws rhel create [flags]

Flags:
      --airgap                            if this flag is set the host will be created as airgap machine. Access will done through a bastion
      --arch string                       architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --conn-details-output string        path to export host connection information (host, username and privateKey)
  -h, --help                              help for create
      --rh-subscription-password string   password to register the subscription
      --rh-subscription-username string   username to register the subscription
      --snc                               if this flag is set the RHEL will be setup with SNC profile. Setting up all requirements to run https://github.com/crc-org/snc
      --spot                              if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)
      --tags stringToString               tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                    version for the RHEL OS (default "9.4")

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

### Outputs

* It will crete an instance and will give as result several files located at path defined by `--conn-details-output`:

  * **host**: host for the windows machine (lb if spot)
  * **username**: username to connect to the machine
  * **id_rsa**: private key to connect to machine
  * **bastion_host**: host for the bastion (airgap)
  * **bastion_username**: username to connect to the bastion (airgap)
  * **bastion_id_rsa**: private key to connect to the bastion (airgap)

* Also it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to detroy the resources.

### Container

When running the container image it is required to pass the authetication information as variables(to setup AWS credentials there is a [helper script](./../../hacks/aws_setup.sh)), following a sample snipped on how to create an instance with default values:  

```bash
# x86_64
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws rhel create \
            --project-name mapt-rhel \
            --backed-url file:///workspace \
            --rh-subscription-password XXXX \
            --rh-subscription-username XXXXX \
            --conn-details-output /workspace

# arm64
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws rhel create \
            --project-name mapt-rhel \
            --arch arm64 \
            --backed-url file:///workspace \
            --rh-subscription-password XXXX \
            --rh-subscription-username XXXXX \
            --conn-details-output /workspace
```