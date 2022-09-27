#!/bin/bash

VALID_CONFIG=true
# Check required ENVs
if [ -z "${INSTANCE_TYPES}" ]; then 
  echo "INSTANCE_TYPES is required"
  VALID_CONFIG=false
fi

if [ -z "${PRODUCT_DESCRIPTION}" ]; then 
  echo "PRODUCT_DESCRIPTION ENV is required"
  VALID_CONFIG=false  
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

# Run qenvs
exec qenvs spot \
    --instance-types "${INSTANCE_TYPES}" \
    --product-description "${PRODUCT_DESCRIPTION}"
