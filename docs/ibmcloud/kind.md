# Overview

This action provisions an Ubuntu 24.04 x86_64 VM on IBM Cloud VPC and runs a single-node [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) cluster on it. The instance is assigned a floating IP for direct SSH and Kubernetes API access.

The Kind cluster is configured so that the API server certificate includes the public floating IP as a SAN, enabling direct `kubectl` access via the exported kubeconfig.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `IBMCLOUD_API_KEY` | yes | IBM Cloud API key |
| `IC_REGION` | yes | IBM Cloud region (e.g. `us-south`, `us-east`) |
| `IC_ZONE` | yes | Availability zone (e.g. `us-south-2`) |
| `IBMCLOUD_COS_ACCESS_KEY_ID` | only with S3 `--backed-url` | HMAC access key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_SECRET_ACCESS_KEY` | only with S3 `--backed-url` | HMAC secret key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_ENDPOINT` | no | COS S3 endpoint (defaults to `s3.<region>.cloud-object-storage.appdomain.cloud`) |

## Create

```bash
mapt ibmcloud kind create -h
create

Usage:
  mapt ibmcloud kind create [flags]

Flags:
      --conn-details-output string           path to export connection information (host, username, privateKey, kubeconfig)
      --extra-port-mappings string           additional port mappings for the Kind cluster (JSON array of objects with containerPort, hostPort, and protocol)
  -h, --help                                 help for create
      --arch string                          architecture for the machine (default "x86_64")
      --tags stringToString                  tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                       Kubernetes version for the Kind cluster (default "v1.34")

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Outputs

Files written to the path defined by `--conn-details-output`:

| File | Description |
|---|---|
| `host` | Floating IP of the instance |
| `username` | SSH username (`ubuntu`) |
| `id_rsa` | Private key for the instance |
| `kubeconfig` | Kubeconfig for the Kind cluster (API server points to the floating IP) |

A state folder is also created at `--backed-url`. It is required (together with `--project-name`) to destroy the resources later.

### SSH access

```bash
OUTPUT=/path/to/conn-details-output

ssh -i ${OUTPUT}/id_rsa \
    -o StrictHostKeyChecking=no \
    ubuntu@$(cat ${OUTPUT}/host)
```

### Kubernetes access

```bash
OUTPUT=/path/to/conn-details-output

export KUBECONFIG=${OUTPUT}/kubeconfig
kubectl get nodes
```

### Container

```bash
podman run -d --name ibmcloud-kind \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        -e IC_ZONE=us-south-2 \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud kind create \
            --project-name ibmcloud-kind \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --version v1.34
```

### Extra port mappings

The `--extra-port-mappings` flag accepts a JSON array of port mapping objects. Each object must have `containerPort`, `hostPort`, and `protocol` fields. The security group is updated automatically to allow inbound traffic on the extra host ports.

```bash
mapt ibmcloud kind create \
    --project-name ibmcloud-kind \
    --backed-url file:///workspace \
    --conn-details-output /workspace \
    --extra-port-mappings '[{"containerPort":8080,"hostPort":8080,"protocol":"TCP"}]'
```

## Using IBM Cloud Object Storage as S3 backend

To store Pulumi state in IBM COS instead of a local file, create [HMAC credentials](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-uhc-hmac-credentials-main) for your COS instance and pass an `s3://` backed URL:

```bash
podman run -d --name ibmcloud-kind \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        -e IC_ZONE=us-south-2 \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud kind create \
            --project-name ibmcloud-kind \
            --backed-url s3://my-cos-bucket \
            --conn-details-output /workspace
```

An HTTPS endpoint URL is also supported as `--backed-url`, with the bucket name in the path:

```
--backed-url https://s3.us-south.cloud-object-storage.appdomain.cloud/my-cos-bucket
```

## Destroy

```bash
podman run -d --name ibmcloud-kind \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud kind destroy \
            --project-name ibmcloud-kind \
            --backed-url file:///workspace
```

By default, destroy removes the Pulumi state files from the backend after a successful destroy. Use `--keep-state` to preserve them:

```bash
podman run -d --name ibmcloud-kind \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud kind destroy \
            --project-name ibmcloud-kind \
            --backed-url s3://my-cos-bucket \
            --keep-state
```
