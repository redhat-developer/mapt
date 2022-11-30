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

# /dev/null
jq_cmd () {
    echo '#!/bin/bash' | tee tmp-jq.sh>/dev/null
    echo "jq ${2} /data" | tee -a tmp-jq.sh>/dev/null
    chmod +x tmp-jq.sh
    result=$(${CONTAINER_RUNTIME} run -v "$PWD/${1}":/data:Z -v "$PWD/tmp-jq.sh":/usr/local/bin/tmp-jq.sh:Z -ti quay.io/biocontainers/jq:1.6 tmp-jq.sh)>/dev/null
    rm tmp-jq.sh
    echo $result
}

# We will use a custom image to ensure we got the tools used on hcl scripts
${CONTAINER_RUNTIME} build -t qenvs-packer -f oci/Dockerfile
                 
# # build_cmd "init . && build ami.pkr.hcl"
build_cmd "init ."
# # -var localize=spanish
build_cmd "build -var crc-version=2.10.2 -var crc-distributable-url='https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/2.10.2/crc-windows-installer.zip' ."

jq_cmd "${3}/manifest.json" "'(.builds[-1].artifact_id |= split(\":\")) | .builds[-1].artifact_id[1]'" > ami-id

# # Cleanup
# rm -rf "${PWD}/${3}/build.log"
# rm -rf "${PWD}/${3}/.packer.d"
# rm -rf "${PWD}/${3}/manifest.json"
