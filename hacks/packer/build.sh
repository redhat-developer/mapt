#!/bin/bash

#Usage ./build.sh ACCESS_KEY SECRET_KEY REGION IMAGE

CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-"podman"}"

# cat > $PACKER_CLI <<- EOM
# ${CONTAINER_RUNTIME} run --rm -it \

# Line 2.
# EOM

BUILD="${CONTAINER_RUNTIME} run -it --rm \
                -e AWS_ACCESS_KEY_ID=${1} \
                -e AWS_SECRET_ACCESS_KEY=${2} \
                -e PACKER_PLUGIN_PATH=/workspace/.packer.d/plugins \
                -v ${PWD}/${3}:/workspace:Z \
                -w /workspace \
                docker.io/hashicorp/packer:latest"

build_cmd () {
    ${BUILD} ${1}
}
                 
# build_cmd "init . && build ami.pkr.hcl"
build_cmd "init ."
build_cmd "build ."

rm -rf "${PWD}/${3}/.packer.d"
