#!/bin/bash

podman run -it --rm mcr.microsoft.com/azure-cli

az login

# Go to auth page and add the code

# Get the suscription ID
subscriptionId="$(az account list --query "[?isDefault].id" --output tsv)"
echo $subscriptionId

# Create SP 
az ad sp create-for-rbac --name ${name} \
                         --role Contributor \
                         --scopes /subscriptions/${subscriptionId}

# Create rg for blob container for pulumi state

az group create \
    --name crc-mapt \
    --location westeurope

az storage account create \
    --name crcmapt \
    --resource-group crc-mapt \
    --location westeurope \
    --sku Standard_ZRS \
    --encryption-services blob \
    --allow-blob-public-access false

az storage container create \
    --account-name crcmapt \
    --name crc-mapt-state \
    --auth-mode login

# Get az storage account key to set on AZURE_STORAGE_KEY
# https://www.pulumi.com/docs/concepts/state/#azure-blob-storage
az storage account keys list --account-name crcmapt