#!/bin/bash
set -x

# Get the suscription ID
subscriptionId="$(az account list --query "[?isDefault].id" --output tsv)"
echo $subscriptionId

# Create SP 
az ad sp create-for-rbac --name ${1}-sp \
                         --role Contributor \
                         --scopes /subscriptions/${subscriptionId}

appId="$(az ad sp list --display-name ${1}-sp --query "[].id" --output tsv)"

# We need to be able to asign 
cat <<EOF > mapt-aks-role.json
{
  "Name": "Mapt AKS Operator",
  "IsCustom": true,
  "Description": "Can create aks clusters with mapt features.",
  "Actions": [
    "Microsoft.Authorization/roleAssignments/*"
  ],
  "NotActions": [],
  "DataActions": [],
  "NotDataActions": [],
  "AssignableScopes": [
    "/subscriptions/${subscriptionId}"
  ]
}
EOF

az role definition create --role-definition mapt-aks-role.json

az role assignment create --assignee  ${appId} \
--role "Mapt AKS Operator" \
--scope "/subscriptions/${subscriptionId}"

# Create rg for blob container for pulumi state

san=$(echo "${1}maptsa" | tr -cd '[:alnum:]')

az group create \
    --name ${1}-mapt-rg \
    --location westeurope

az storage account create \
    --name ${san} \
    --resource-group ${1}-mapt-rg \
    --location westeurope \
    --sku Standard_ZRS \
    --encryption-services blob \
    --allow-blob-public-access false

az storage container create \
    --account-name ${san} \
    --name ${1}-mapt-state \
    --auth-mode login

# Get az storage account key to set on AZURE_STORAGE_KEY
# https://www.pulumi.com/docs/concepts/state/#azure-blob-storage
az storage account keys list --account-name ${san}