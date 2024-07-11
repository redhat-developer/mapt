# Overview

This actions will handle provision for Mac instances on AWS. Some key considerations:

* AWS supports x86, M1 and M2 archs for mac machines although not all of them are supported on all regions
* There is no spot market for mac machines
* Machine (dedicated host) should be provisioned at least for 24 hours

mapt will take care of:

* Try to honor the AWS_DEFAULT_REGION to spin the mac machine and look for a random AZ on it to launch it. In case the machine is no offered on the region as long as flag for `--fixed-location` is not set mapt will move the Machine to a region were it is offered.

* Previous graph also applies when there is no capacity on an specific region. It could happen that a region offers the machine but there is no capacity under that circumstance it will dinamically move it to other region.

* As the machine will be active for 24 hours, mapt will allow to spin multiple machines (only once at a time) during that period (i.e different OS versions, or setups like airgap or vpn connected machine).

* (Future) Create a scheduled task to remove the machine after 24 hours to avoid missing resources. This will require using s3 as backed URL.

![Mac](./mac.svg)

## Create

```bash
mapt aws mac create -h 
create

Usage:
  mapt aws mac create [flags]

Flags:
      --airgap                       if this flag is set the host will be created as airgap machine. Access will done through a bastion
      --arch string                  mac architecture allowed values x86, m1, m2. Default to m2 (default "m2")
      --conn-details-output string   path to export host connection information (host, username and privateKey)
      --fixed-location               if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)
  -h, --help                         help for create
      --host-id string               host id to create the mac instance. If the param is not pass the dedicated host will be created
      --only-host                    if this flag is set only the host will be created / destroyed
      --only-machine                 if this flag is set only the machine will be destroyed
      --tags stringToString          tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string               macos operating system vestion 11, 12 on x86 and m1; 13 on all archs. Default to 13 (default "14")

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

### Outputs

* It will crete an instance and will give as result several files located at path defined by `--conn-details-output`:
  * Creating dedicated host:
    * **dedicatedHostID**: dedicated host id used to create mac machines on it )
  * Creating mac machine:
    * **host**: host for the mac machine
    * **username**: username to connect to the mac machine
    * **id_rsa**: private key to connect to mac machine
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
        quay.io/redhat-developer/mapt:0.6.8 aws mac create \
            --project-name mapt-mac \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```