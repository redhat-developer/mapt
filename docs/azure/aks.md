# Overview

mapt offers several operations to manage environments within azure:

## AKS

It creates / destroy a basic AKS Cluster. In order to run this operation the user / app requires a custom role, to create or check it see the [azure setup script](./../../hacks/azure/azure_setup.sh)

### Operations

#### Create

This will create K8s cluster using the AKS accordig to params specificed:

```bash
podman run -it --rm quay.io/redhat-developer/mapt:0.7.0-dev azure aks create -h
create

Usage:
  mapt azure aks create [flags]

Flags:
      --conn-details-output string       path to export host connection information (host, username and privateKey)
      --enable-app-routing               enable application routing add-on with NGINX
  -h, --help                             help for create
      --location string                  location for created resources in case spot flag (if available) is not passed (default "West US")
      --only-system-pool                 if we do not need bunch of resources we can run only the systempool. More info https://learn.microsoft.com/es-es/azure/aks/use-system-pools?tabs=azure-cli#system-and-user-node-pools
      --spot                             if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string   if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --tags stringToString              tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                   AKS K8s cluster version (default "1.30")
      --vmsize string                    VMSize to be used on the user pool. Typically this is used to provision spot node pools (default "Standard_D8as_v5")

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete an AKS cluster and kubeconfig file to connect will be placed at `--conn-details-output`:

When running the container image it is required to pass the authetication information as variables, following a sample snippet on how to create
a instance with default values:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/redhat-developer/mapt:0.7.0-dev azure \
        aks create \
        --project-name "aks" \
        --backed-url "file:///workspace" \
        --conn-details-output "/workspace" \
        --enable-app-routing \
        --spot
```

It is important to notice that backed-url contains the state of the resources on azure linked to the aks creation, to destroy we need to reference them as so we should pass same folder as we set for creation:

```bash
podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e ARM_TENANT_ID=${ati_value} \
    -e ARM_SUBSCRIPTION_ID=${asi_value} \
    -e ARM_CLIENT_ID=${aci_value} \
    -e ARM_CLIENT_SECRET=${acs_lue} \
    quay.io/redhat-developer/mapt:0.7.0-dev azure \
        aks destroy \
        --project-name "aks" \
        --backed-url "file:///workspace"
```
