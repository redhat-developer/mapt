# Overview

mapt offers several operations to manage environments within azure:

## AKS

It creates / destroy a basic AKS Cluster. In order to run this operation the user / app requires a custom role, to create or check it see the [azure setup script](./../../hacks/azure/azure_setup.sh)

The cluster has two node pools:

- `systempool` hosts the AKS system pods. It always runs on demand with a single node of the size set by `--system-pool-vmsize`, `Standard_A2_v2` by default.
- `userpool` hosts the workloads. It runs `--workers-desired` nodes, 1 by default, on spot instances when `--spot` is set and on demand otherwise.

With `--spot` the spot search selects both the location and the VM size of the user pool among the `--vmsize` options, and the whole cluster is created in that location, so `--location` is ignored. The spot search only takes the user pool sizes into account, so the system pool size must also be available in the selected location. With `--only-system-pool` no user pool is created, so no spot search runs and `--location` and `--vmsize` are not used.

### Operations

#### Create

This will create K8s cluster using the AKS according to params specified:

```bash
podman run -it --rm quay.io/redhat-developer/mapt:0.7.0-dev azure aks create -h
create

Usage:
  mapt azure aks create [flags]

Flags:
      --conn-details-output string        path to export host connection information (host, username and privateKey)
      --enable-app-routing[=true|false]   enable application routing add-on with NGINX
  -h, --help[=true|false]                 help for create
      --location string                   location for created resources in case spot flag (if available) is not passed (default "West US")
      --only-system-pool[=true|false]     if we do not need bunch of resources we can run only the systempool. More info https://learn.microsoft.com/es-es/azure/aks/use-system-pools?tabs=azure-cli#system-and-user-node-pools
      --spot[=true|false]                 if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string    if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --spot-excluded-regions strings     Comma-separated list of zone IDs to exclude from spot selection
      --spot-increase-rate int            Percentage to be added on top of the current calculated spot price to increase chances to get the machine (default 30)
      --system-pool-vmsize string         VMSize to be used on the system pool. The spot search only takes the user pool size into account, so the system pool size may not be available in the selected location (default "Standard_A2_v2")
      --tags stringToString               tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                    AKS K8s cluster version (default "1.31")
      --vmsize strings                    comma separated list of VMSizes to be used on the user pool. When spot is set the best option across the sizes is selected, otherwise the first one is used (default [Standard_D8as_v5])
      --workers-desired int               number of nodes on the user pool (default 1)

Global Flags:
      --backed-url string     backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket
      --project-name string   project name to identify the instance of the stack
```

It will crete an AKS cluster and kubeconfig file to connect will be placed at `--conn-details-output`:

When running the container image it is required to pass the authentication information as variables, following a sample snippet on how to create
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
