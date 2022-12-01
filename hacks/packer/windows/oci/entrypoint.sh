#!/bin/sh

VALID_CONFIG=true

if [ -z "${AWS_ACCESS_KEY_ID+x}" ] || [ -z "${AWS_SECRET_ACCESS_KEY+x}" ]; then 
  echo "AWS ENV for credentials are required"
  VALID_CONFIG=false  
fi

if [ "${VALID_CONFIG}" = false ]; then
  echo "Add the required ENVs"
  exit 1
fi

packer init .

build_vars=""

if [ ! -z "${AWS_REGION+x}" ] ; then
  build_vars="${build_vars} -var region=${AWS_REGION}"
fi

if [ ! -z "${LOCALIZE+x}" ] ; then
  build_vars="${build_vars} -var localize=${LOCALIZE}"
fi

if [ ! -z "${CRC_VERSION+x}" ] ; then
  build_vars="${build_vars} -var crc-version=${CRC_VERSION}"
fi

if [ ! -z "${CRC_DISTRIBUTABLE_URL+x}" ] ; then
  build_vars="${build_vars} -var crc-distributable-url=${CRC_VERSION}"
fi

packer build ${build_vars} .

OUTPUT="${OUTPUT:-"/output"}"

mkdir -p "${OUTPUT}"

jq '(.builds[-1].artifact_id |= split(":")) | .builds[-1].artifact_id[1]' manifest.json > "${OUTPUT}/ami-id"
