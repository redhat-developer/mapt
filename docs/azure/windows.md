# Overview

mapt offers several operations to manage environments within azure:

## Windows

It creates / destroy a windows dekstop edition ready to be included within the CI/CD system. Features included within the offering:

* Creates an admin user with a self generated passwd (only accessible within rdp)
* Creates an worker user and add it to the admins group
* Setup ssh for the worker user with a self generated private key
* Setup autologin for the worker user  
* Disable UAC  
* Running ssh server as a startup process run by the worker user

## Profiles

It is possible to customize the initial setup for the target host based on profiles, a target host can be created with several profiles:

### crc profile

This wil create the crc-users group and add the user to it, to avoid reboot during crc installation 

Side note: the other requirements for reboot are done by default; hyper-v installation and add user to Admin group

### Operations

#### Create

This will create a Windows desktop according to params specified:

```bash
podman run -it --rm quay.io/redhat-developer/mapt:0.7.0-dev azure windows create -h
create

Usage:
  mapt azure windows create [flags]
Flags:
      --admin-username string             username for admin user. Only rdp accessible within generated password (default "rhqpadmin")
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
      --profile strings                   comma seperated list of profiles to apply on the target machine. Profiles available: crc
      --spot                              if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string    if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --tags stringToString               tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --username string                   username for general user. SSH accessible + rdp with generated password (default "rhqp")
      --vmsize strings                    set specific size for the VM and ignore any CPUs, Memory and Arch parameters set. Type requires to allow nested virtualization
      --windows-featurepack string        windows feature pack (default "23h2-pro")
      --windows-version string            Major version for windows desktop 10 or 11 (default "11")

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete a Windows desktop instance and will give as result several files located at path defined by `--conn-details-output`:

* adminusername: file containing the username for admin user
* adminuserpassword: file containing the passwd for admin user
* username: file containing the username for worker user
* userpassword: file containing the passwd for worker user
* id_rsa: file containing the private key for worker user
* host: file containing the public ip for the instance  

Also, it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to destroy the resources.

When running the container image it is required to pass the authentication information as variables, following a sample snipped on how to create
an instance with default values:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/redhat-developer/mapt:0.7.0-dev azure \
        windows create \
        --project-name "win-desk-11" \
        --backed-url "file:///workspace" \
        --conn-details-output "/workspace" \
        --profile crc \
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
        windows destroy \
        --project-name "win-desk-11" \
        --backed-url "file:///workspace"
```
