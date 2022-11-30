# Overview

This folder contains severeal scripts to allow building images to improve spin up times, specially when the setups requires rebooting the machine.

## Windows

THis will setup a windows image with:

* create user with admin privileges
* setup autologin for the user
* sshd enabled  
* setup auth based on private key
* enable hyper-v  
* setup specific UAC levels to allow running privileged without prompt

Image will be created on one region, it is required to run `qenvs ami replicate` to copy on each region, to allow use best bid spot module with it.  

**IMPORTANT** On booting the image it is required to add userdata to copy paste content for .ssh/authorized_keys with valid openssh public key to match the desired private key. THis setup only creates a fake file to emulate the behavior.  

### Create images from container

* Windows_Server-2019-English-Full-HyperV-RHQE

This command will create an AMI with name (Windows_Server-2019-English-Full-HyperV-RHQE) on the target region, also it will output a file
`ami-id` holding the ami id for the custom image

```bash
podman run -d --rm \
    -v $PWD:/output:Z \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_REGION="${TARGET_REGION}" \
    quay.io/ariobolo/qenvs-packer-windows:latest
```

* Windows_Server-2019-Spanish-Full-HyperV-RHQE

This command will create an AMI with name (Windows_Server-2019-Spanish-Full-HyperV-RHQE) on the target region, also it will output a file
`ami-id` holding the ami id for the custom image

```bash
podman run -d --rm \
    -v $PWD:/output:Z \
    -e LOCALIZE=spanish \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_REGION="${TARGET_REGION}" \
    quay.io/ariobolo/qenvs-packer-windows:latest
```

* Windows_Server-2019-English-Full-OCPL-${CRC-VERSION}-RHQE

This command will create an AMI with name (Windows_Server-2019-English-Full-OCPL-${CRC-VERSION}-RHQE) on the target region, also it will output a file
`ami-id` holding the ami id for the custom image

```bash
podman run -d --rm \
    -v $PWD:/output:Z \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_REGION="${TARGET_REGION}" \
    -e CRC_VERSION="${CRC_VERSION}" \
    -e CRC_DISTRIBUTABLE_URL="${CRC_DISTRIBUTABLE_URL}" \
    quay.io/ariobolo/qenvs-packer-windows:latest
```

* Windows_Server-2019-Spanish-Full-OCPL-${CRC-VERSION}-RHQE

This command will create an AMI with name (Windows_Server-2019-Spanish-Full-OCPL-${CRC-VERSION}-RHQE) on the target region, also it will output a file
`ami-id` holding the ami id for the custom image

```bash
podman run -d --rm \
    -v $PWD:/output:Z \
    -e LOCALIZE=spanish \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_REGION="${TARGET_REGION}" \
    -e CRC_VERSION="${CRC_VERSION}" \
    -e CRC_DISTRIBUTABLE_URL="${CRC_DISTRIBUTABLE_URL}" \
    quay.io/ariobolo/qenvs-packer-windows:latest
```
