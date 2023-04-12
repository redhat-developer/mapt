# Overview

qenvs offers several operations to manage environments within azure:

## Windows

It creates / destroy a windows dekstop edition ready to be included within the CI/CD system. Features included within the offering:

* Creates an admin user with a self generated passwd (only accessible within rdp)
* Creates an worker user and add it to the admins group
* Setup ssh for the worker user with a self generated private key
* Setup autologin for the worker user  
* Disable UAC  
* Running ssh server as a startup process run by the worker user

### Operations

#### Create

This will create a windows desktop accordig to params specificed:

```bash
podman run -it --rm quay.io/rhqp/qenvs:0.0.4 azure windows create -h
create

Usage:
  qenvs azure windows create [flags]

Flags:
      --admin-username string        username for admin user. Only rdp accessible within generated password (default "rhqpadmin")
      --conn-details-output string   path to export host connection information (host, username and privateKey)
  -h, --help                         help for create
      --location string              location for created resources within Windows desktop (default "West US")
      --tags stringToString          tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --username string              username for general user. SSH accessible + rdp with generated password (default "rhqp")
      --vmsize string                size for the VM. Type requires to allow nested virtualization (default "Standard_D4_v5")
      --windows-featurepack string   windows feature pack (default "22h2-pro")
      --windows-version string       Major version for windows desktop 10 or 11 (default "11")

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete a windows desktop instance and will give as result several files located at path defined by `--conn-details-output`:

* adminusername: file containing the username for admin user
* adminuserpassword: file containing the passwd for admin user
* username: file containing the username for worker user
* userpassword: file containing the passwd for worker user
* id_rsa: file containing the private key for worker user
* host: file containing the public ip for the instance  

Also it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to detroy the resources.

When running the container image it is required to pass the authetication information as variables, following a sample snipped on how to create
a instance with default values:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/rhqp/qenvs:0.0.4 azure \
        windows create \
        --project-name "win-desk-11" \
        --backed-url "file:///workspace" \
        --conn-details-output "/workspace" 
```

The following is a snipped on how to destroy the resources:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/rhqp/qenvs:0.0.4 azure \
        windows destroy \
        --project-name "win-desk-11" \
        --backed-url "file:///workspace"
```
