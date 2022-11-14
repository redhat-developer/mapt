#!/bin/bash

#Usage ./build.sh ACCESS_KEY SECRET_KEY PKR_TEMPLATE_FOLDER

CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-"podman"}"
# -e PACKER_LOG=1 \
BUILD="${CONTAINER_RUNTIME} run -it --rm \
                -e AWS_ACCESS_KEY_ID=${1} \
                -e AWS_SECRET_ACCESS_KEY=${2} \
                -e PACKER_PLUGIN_PATH=/workspace/.packer.d/plugins \
                -v ${PWD}/${3}:/workspace:Z \
                -w /workspace \
                localhost/qenvs-packer:latest"

build_cmd () {
    ${BUILD} ${1}
}

# We will use a custom image to ensure we got the tools used on hcl scripts
${CONTAINER_RUNTIME} build -t qenvs-packer -f images/Dockerfile
                 
# build_cmd "init . && build ami.pkr.hcl"
build_cmd "init ."
build_cmd "build -var crc-distributable-url='https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/2.10.2/crc-windows-installer.zip' ."
# build_cmd "build -machine-readable /workspace/ami.pkr.hcl | tee build.log"

# Extract ami-id
# grep 'artifact,0,id' ${PWD}/${3}/build.log | cut -d, -f6 | cut -d: -f2 > ${PWD}/${3}/ami-id

# Cleanup
rm -rf "${PWD}/${3}/build.log"
rm -rf "${PWD}/${3}/.packer.d"
