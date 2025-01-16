#!/bin/bash

az graph query -q "Resources
| where type == 'microsoft.compute/virtualmachines'
| where properties.extended.instanceView.powerState.code == 'PowerState/deallocated'
| project resourceGroup" > groups.json

jq -r 'recurse | scalars' groups.json > groups.list

# Define the file path
file="groups.list"

# Read the file line by line
while IFS= read -r line
do
    echo "Processing resource group: $line"
    # Perform operations, e.g., delete the resource group
    az group delete --name $line --yes --no-wait
done < "$file"