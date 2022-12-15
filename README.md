# qenvs

automation for qe environments using pulumi

![code check](https://github.com/adrianriobo/qenvs/actions/workflows/build-go.yaml/badge.svg)
![oci builds](https://github.com/adrianriobo/qenvs/actions/workflows/build-oci.yaml/badge.svg)

## Environment

Create a composable environment with different qe target machines aggregated on different topologies and with specific setups (like vpns, proxys, airgaps,...)

Current available features allow to create supported hosts on AWS using cmd `qenvs host create` current supported hosts:

* RHEL v8 (host id `ol-rhel`)
* MacOS M1 v12 (host id `g-macos-m1`)

![Environment](./docs/diagrams/base.svg)

## Spot price use case

This module allows to check for best bid price on all regions, to request instances at lower price to reduce costs. To calculate the best option, it is also required to:  

* reduce interruptions
* ensure capacity

to check those requisites the module make use of spot placement scores based on machine requirements. Then best scores are crossed with lowers price from spot price history to pick the most valuable option.

Current use case is working on one machine but it will be exteded to analyze any required environment offered by qenvs (checking with all the machines included on a specific environment).

Current information about supported machines can be checked at [support-matrix](pkg/infra/aws/support-matrix/matrix.go)

## Build and usage

qenvs can be build as container

```bash
make container-build
```

run qenvs container, to setup AWS credentials there is a [helper script](hacks/aws_setup.sh)

```bash
# state and connection outputs will be created in this folder
mkdir -p output

# Create rhel host
# Recommended this region us-east-1
# https://github.com/adrianriobo/qenvs/issues/24
podman run -d --name qenvs-rhel \
        -v $PWD/output:/data/qenvs:Z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-1 \
        -e PROJECT_NAME=qenvs-rhel \
        -e BACKED_URL=file:///data/qenvs \
        -e CONNECTION_OUTPUT=/data/qenvs \
        -e OPERATION=create \
        -e SUPPORTED_HOST_ID=ol-rhel \
        quay.io/ariobolo/qenvs:0.0.1

# state should be passed to container for destroy
# project name should be the same
podman run -d --name qenvs-rhel \
        -v $PWD/output:/data/qenvs:Z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        # Recommended this region 
        # https://github.com/adrianriobo/qenvs/issues/24
        -e AWS_DEFAULT_REGION=us-east-1 \
        -e PROJECT_NAME=qenvs-rhel \
        -e BACKED_URL=file:///data/qenvs \
        -e OPERATION=destroy \
        quay.io/ariobolo/qenvs:0.0.1
```

### Tekton

To facilitate the inclusion within a pipeline a [task defintion](hacks/tekton/infra-management.yaml) can be used  as wrapper.

### Packer

Due to hard requirements on nested virtualization, on AWS it is required to run baremetal machines, these type of machines have extra validations (in addtion to the standar ones for virtualize ec2 instances) which increases time on every boot.

To minimize the impact of reboots the project contains [packer templates](hacks/packer/README.md) to create custom AMIs on those scnearios 
where userdata requires reboots.  

Also notice images use to be created per region, and the process itself requires to spinup a machine. As qenvs offers the feature to dynamically adjust the location of the infrastructure based on best bid price, we need to replicate those images per region. To avoid extra cost with packer qenvs offers now a new command to replicate the ami created by packer `qenvs ami replicate`
