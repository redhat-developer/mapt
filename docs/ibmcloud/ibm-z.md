# Overview

This action provisions an Ubuntu 22.04 s390x machine on IBM Cloud VPC. The instance is assigned a floating IP for direct SSH access.

Two networking modes are supported:

- **Existing subnet** (`--subnet-id`): the instance is placed in a pre-existing VPC subnet. VPC, subnet, and gateway are not created. Only `IC_REGION` is required.
- **Auto-provision** (no `--subnet-id`): a new VPC, subnet, and public gateway are created from scratch. Both `IC_REGION` and `IC_ZONE` are required.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `IBMCLOUD_ACCOUNT` | yes | IBM Cloud account ID |
| `IBMCLOUD_API_KEY` | yes | IBM Cloud API key |
| `IC_REGION` | yes | IBM Cloud region (e.g. `us-south`, `us-east`) |
| `IC_ZONE` | only without `--subnet-id` | Availability zone (e.g. `us-south-2`) |
| `IBMCLOUD_COS_ACCESS_KEY_ID` | only with S3 `--backed-url` | HMAC access key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_SECRET_ACCESS_KEY` | only with S3 `--backed-url` | HMAC secret key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_ENDPOINT` | no | COS S3 endpoint (defaults to `s3.<region>.cloud-object-storage.appdomain.cloud`) |

## Create

```bash
mapt ibmcloud ibm-z create -h
create

Usage:
  mapt ibmcloud ibm-z create [flags]

Flags:
      --conn-details-output string           path to export host connection information (host, username and privateKey)
      --ghactions-runner-labels strings      List of labels separated by comma to be added to the self-hosted runner
      --ghactions-runner-repo string         Full URL of the repository where the Github Actions Runner should be registered
      --ghactions-runner-token string        Token needed for registering the Github Actions Runner token
  -h, --help                                 help for create
      --it-cirrus-pw-labels stringToString   additional labels to use on the persistent worker (--it-cirrus-pw-labels key1=value1,key2=value2) (default [])
      --it-cirrus-pw-token string            Add mapt target as a cirrus persistent worker. The value will hold a valid token to be used by cirrus cli to join the project.
      --otel-app-code string                 OpenTelemetry appcode identifier (e.g. MAPT-001); when set together with --otel-auth-token, installs the otelcol-contrib filelog collector on the instance
      --otel-auth-token string               OpenTelemetry authentication token (UUID) used to authenticate against the OTLP endpoint
      --otel-endpoint string                 OTLP HTTP endpoint to export logs to (default "https://otel-input.corp.redhat.com")
      --subnet-id string                     ID of an existing VPC subnet to deploy the instance into (optional)
      --tags stringToString                  tags to add on each resource (--tags name1=value1,name2=value2) (default [])

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Outputs

Files written to the path defined by `--conn-details-output`:

| File | Description |
|---|---|
| `host` | Floating IP of the instance (direct SSH) |
| `username` | SSH username (`ubuntu`) |
| `id_rsa` | Private key for the instance |

A state folder is also created at `--backed-url`. It is required (together with `--project-name`) to destroy the resources later.

### SSH access

```bash
OUTPUT=/path/to/conn-details-output

ssh -i ${OUTPUT}/id_rsa \
    -o StrictHostKeyChecking=no \
    ubuntu@$(cat ${OUTPUT}/host)
```

### Container

```bash
# Using an existing VPC subnet
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z create \
            --project-name ibm-z \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --subnet-id <subnet-id>

# Auto-provisioning VPC, subnet, and gateway
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        -e IC_ZONE=us-south-2 \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z create \
            --project-name ibm-z \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```

## OpenTelemetry log collection

When both `--otel-app-code` and `--otel-auth-token` are provided, cloud-init installs `otelcol-contrib` on the instance at first boot and configures it to ship `/var/log/syslog` and `/var/log/auth.log` to the OTLP endpoint.

```bash
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z create \
            --project-name ibm-z \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --subnet-id <subnet-id> \
            --otel-app-code MAPT-001 \
            --otel-auth-token <uuid-token>
```

## Using IBM Cloud Object Storage as S3 backend

To store Pulumi state in IBM COS instead of a local file, create [HMAC credentials](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-uhc-hmac-credentials-main) for your COS instance and pass an `s3://` backed URL:

```bash
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-south \
        -e IC_ZONE=us-south-2 \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z create \
            --project-name ibm-z \
            --backed-url s3://my-cos-bucket \
            --conn-details-output /workspace
```

An HTTPS endpoint URL is also supported as `--backed-url`, with the bucket name in the path:

```
--backed-url https://s3.us-south.cloud-object-storage.appdomain.cloud/my-cos-bucket
```

The COS endpoint and `PULUMI_BACKEND_URL` are constructed automatically from the region and bucket name.

## Destroy

```bash
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z destroy \
            --project-name ibm-z \
            --backed-url file:///workspace
```

By default, destroy removes the Pulumi state files from the backend after a successful destroy. Use `--keep-state` to preserve them:

```bash
podman run -d --name ibm-z \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-south \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:v0.8.0 ibmcloud ibm-z destroy \
            --project-name ibm-z \
            --backed-url s3://my-cos-bucket \
            --keep-state
```
