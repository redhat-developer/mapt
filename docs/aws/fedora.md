# Overview

This actions will handle provision Fedora Cloud machines on dedicated hosts. This is a requisite to run nested virtualization on AWS.
 

## Create

```bash
mapt aws fedora create -h
create

Usage:
  mapt aws fedora create [flags]
Flags:
      --airgap                            if this flag is set the host will be created as airgap machine. Access will done through a bastion
      --arch string                       architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --conn-details-output string        path to export host connection information (host, username and privateKey)
      --cpus int32                        Number of CPUs for the cloud instance (default 8)
      --ghactions-runner-labels strings   List of labels separated by comma to be added to the self-hosted runner
      --ghactions-runner-name string      Name for the Github Actions Runner
      --ghactions-runner-repo string      Full URL of the repository where the Github Actions Runner should be registered
      --ghactions-runner-token string     Token needed for registering the Github Actions Runner token
  -h, --help                              help for create
      --install-ghactions-runner          Install and setup Github Actions runner in the instance
      --memory int32                      Amount of RAM for the cloud instance in GiB (default 64)
      --nested-virt                       Use cloud instance that has nested virtualization support
      --spot                              if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)
      --tags stringToString               tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                    version for the Fedora Cloud OS (default "40")
      --vm-types strings                  set an specific set of vm-types and ignore any CPUs, Memory, Arch parameters set. Note vm-type should match requested arch. Also if --spot flag is used set at least 3 types.

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

### Outputs

* It will crete an instance and will give as result several files located at path defined by `--conn-details-output`:

  * **host**: host for the Windows machine (lb if spot)
  * **username**: username to connect to the machine
  * **id_rsa**: private key to connect to machine
  * **bastion_host**: host for the bastion (airgap)
  * **bastion_username**: username to connect to the bastion (airgap)
  * **bastion_id_rsa**: private key to connect to the bastion (airgap)

* Also, it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to destroy the resources.

### Container

When running the container image it is required to pass the authentication information as variables(to setup AWS credentials there is a [helper script](./../../hacks/aws_setup.sh)), following a sample snipped on how to create an instance with default values:  

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
