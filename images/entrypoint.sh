#!/bin/bash

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
  echo "${DEFAULT_BACKED_URL} will be used as backed url it will be exported as volume"
  BACKED_URL="${DEFAULT_BACKED_URL}"
fi

if [ -z "${AWS_ACCESS_KEY_ID}" ] || [ -z "${AWS_SECRET_ACCESS_KEY}" ] || [ -z "${AWS_DEFAULT_REGION}" ]; then 
  echo "AWS ENV for credentials are required"
  VALID_CONFIG=false  
fi

if [ -z "${PULUMI_CONFIG_PASSPHRASE}" ]; then 
  echo "PULUMI_CONFIG_PASSPHRASE ENV is required"
  VALID_CONFIG=false  
fi

if [ "${VALID_CONFIG}" = false ]; then
  echo "Add the required ENVs"
  exit 1
fi

# //https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
AWS_SDK_LOAD_CONFIG=1
# https://www.pulumi.com/docs/reference/cli/environment-variables/
PULUMI_CONFIG_PASSPHRASE="passphrase"

if [[ "${OPERATION}" == "create" ]]; then
    if [ -z "${SUPPORTED_HOST_ID}" ]; then 
      echo "SUPPORTED_HOST_ID is required"
      VALID_CONFIG=false
    fi
    exec qenvs host create \
      --project-name "${PROJECT_NAME}" \
      --backed-url "${BACKED_URL}" \
      --host-id "${SUPPORTED_HOST_ID}"
else 
  exec qenvs host destroy \
      --project-name "${PROJECT_NAME}" \
      --backed-url "${BACKED_URL}" 
fi



