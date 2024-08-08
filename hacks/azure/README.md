# Overview 

This folder contais some helper scripts to run on azure:

* azure_setup.sh will create an app as a service principal on Azure with the required permissions to create all resources managed by mapt.

* azure_delete_rg_by_spot_stopped.h help to delete dangling resources in case of an error while destroying operations

# How to run

We can use the azure cli container image and map this folder to run the scripts:

```bash
# run az cli image
podman run -it --rm -v $PWD:/mapt:z --workdir /mapt mcr.microsoft.com/azure-cli 
# login
az login
```

From there we can run the scripts:

```bash
# This will create an app to interact with az to be used for the team openshift local on gh 
./azure_setup.sh openshift-local-gh-runner
```