#!/bin/sh

VALID_CONFIG=true
# Check required ENVs
if [ -z "${OPERATION}" ]; then 
  echo "OPERATION is required"
  VALID_CONFIG=false
fi

if [ -z "${PROJECT_NAME}" ]; then 
  echo "PROJECT_NAME ENV is required"
  VALID_CONFIG=false  
fi

if [ -z "${BACKED_URL}" ]; then 
  echo "${INTERNAL_OUTPUT} will be used as backed url it will be exported as volume"
  BACKED_URL="file://${INTERNAL_OUTPUT}"
fi

if [ -z "${CONNECTION_OUTPUT}" ]; then 
  echo "${INTERNAL_OUTPUT} will be used as output folder for connecion resources"
  CONNECTION_OUTPUT="${INTERNAL_OUTPUT}"
fi

if [ -z "${AWS_ACCESS_KEY_ID}" ] || [ -z "${AWS_SECRET_ACCESS_KEY}" ] || [ -z "${AWS_DEFAULT_REGION}" ]; then 
  echo "AWS ENV for credentials are required"
  VALID_CONFIG=false  
fi

if [ -z "${PULUMI_CONFIG_PASSPHRASE}" ]; then 
  # https://www.pulumi.com/docs/reference/cli/environment-variables/
  PULUMI_CONFIG_PASSPHRASE="passphrase"
fi

if [ "${VALID_CONFIG}" = false ]; then
  echo "Add the required ENVs"
  exit 1
fi

# //https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
AWS_SDK_LOAD_CONFIG=1

if [[ "${OPERATION}" == "create" ]]; then
  if [ -z "${SUPPORTED_HOST_ID}" ]; then 
    echo "SUPPORTED_HOST_ID is required"
    VALID_CONFIG=false
  fi
  exec  env AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
        env AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
        env AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
        env AWS_SDK_LOAD_CONFIG=${AWS_SDK_LOAD_CONFIG} \
        env PULUMI_CONFIG_PASSPHRASE=${PULUMI_CONFIG_PASSPHRASE} \
        qenvs host create \
          --project-name "${PROJECT_NAME}" \
          --backed-url "${BACKED_URL}" \
          --conn-details-output "${CONNECTION_OUTPUT}" \
          --host-id "${SUPPORTED_HOST_ID}"
else 
  exec  env AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
        env AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
        env AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
        env AWS_SDK_LOAD_CONFIG=${AWS_SDK_LOAD_CONFIG} \
        env PULUMI_CONFIG_PASSPHRASE=${PULUMI_CONFIG_PASSPHRASE} \
        qenvs host destroy \
          --project-name "${PROJECT_NAME}" \
          --backed-url "${BACKED_URL}" 
fi
