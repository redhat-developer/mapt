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
