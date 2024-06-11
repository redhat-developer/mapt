# Overview

This actions will handle provision Fedora Cloud machines on dedicated hosts. This is a requisite to run nested virtualization on AWS.
 

## Create

```bash
mapt aws fedora create -h
create

Usage:
  mapt aws fedora create [flags]

Flags:
      --airgap                       if this flag is set the host will be created as airgap machine. Access will done through a bastion
      --conn-details-output string   path to export host connection information (host, username and privateKey)
  -h, --help                         help for create
      --spot                         if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)
      --tags stringToString          tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string               version for the Fedora Cloud OS (default "39")

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
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws fedora create \
            --project-name mapt-fedora \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```