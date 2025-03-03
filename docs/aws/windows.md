# Overview

This actions will handle provision Windows Server machines on dedicated hosts. This is a requisite to run nested virtualization on AWS.

Due to how mapt checks the healthy state for the machine and due to some specific characteristics this action is intended for using within a [custom ami](https://github.com/redhat-developer/mapt-builder). 

Some of the customizations this image includes:

* create user with admin privileges
* setup autologin for the user
* sshd enabled
* setup auth based on private key
* enable hyper-v
* setup specific UAC levels to allow running privileged without prompt

## Ami replication

Also, the action is expecting the image exists with the name: `Windows_Server-2022-English-Full-HyperV-RHQE` at least on one region (this AMI can be created using the helper side project [mapt-builder](https://github.com/redhat-developer/mapt-builder)). If `--spot` option is enable and the image is not offered / created on the chosen region it will copy the AMI as part of the stack (As so it will delete it on destroy).

This process (replicate the ami) increase the overall time for spinning the machine, and can be avoided by running the replication cmd on the image to pre replicate the image on all regions.

Also, there is a special flag enabling the AMI to keep beyon the destroy operation `--ami-keep-copy`. In addition, we are using the [fast launch](https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/win-ami-config-fast-launch.html) feature for the AMI to reduce the amount of time for spinning up the machines.

Disclaimer in case of use the `--ami-keep-copy` it will keep the ami and the six snapshots required to enable the fast launch. 

## Create

```bash
mapt aws windows create -h
create

Usage:
  mapt aws windows create [flags]

Flags:
      --airgap                       if this flag is set the host will be created as airgap machine. Access will done through a bastion
      --ami-keep-copy                in case the ami needs to be copied to a target region (i.e due to spot) if ami-keep-copy flag is present the destroy operation will not remove the AMI (this is intended for speed it up on coming provisionings)
      --ami-lang string              language for the ami possible values (eng, non-eng). This param is used when no ami-name is set and the action uses the default custom ami (default "eng")
      --ami-name string              name for the custom ami to be used within windows machine. Check README on how to build it (default "Windows_Server-2022-English-Full-HyperV-RHQE")
      --ami-owner string             alias name for the owner of the custom AMI (default "self")
      --ami-username string          name for de default user on the custom AMI (default "ec2-user")
      --conn-details-output string   path to export host connection information (host, username and privateKey)
  -h, --help                         help for create
      --spot                         if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)
      --tags stringToString          tags to add on each resource (--tags name1=value1,name2=value2) (default [])

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

* Also, it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to destroy the resources.

### Container

When running the container image it is required to pass the authentication information as variables(to setup AWS credentials there is a [helper script](./../../hacks/aws_setup.sh)), following a sample snipped on how to create an instance with default values:  

```bash
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        quay.io/redhat-developer/mapt:0.7.0-dev aws windows create \
            --project-name mapt-windows \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```
