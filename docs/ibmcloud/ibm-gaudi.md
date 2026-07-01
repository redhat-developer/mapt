# Overview

This action provisions an Intel Gaudi 3 accelerated instance on IBM Cloud VPC using the RHEL AI image. The instance uses the `gx3d-160x1792x8gaudi3` profile (160 vCPU, 1792 GB RAM, 8x Gaudi 3 accelerators) and is assigned a floating IP for direct SSH access.

Two networking modes are supported:

- **Existing subnet** (`--subnet-id`): the instance is placed in a pre-existing VPC subnet. VPC, subnet, and gateway are not created. Only `IC_REGION` is required.
- **Auto-provision** (no `--subnet-id`): a new VPC, subnet, and public gateway are created from scratch. Both `IC_REGION` and `IC_ZONE` are required.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `IBMCLOUD_ACCOUNT` | yes | IBM Cloud account ID |
| `IBMCLOUD_API_KEY` | yes | IBM Cloud API key |
| `IC_REGION` | yes | IBM Cloud region (e.g. `us-east`, `us-south`, `eu-de`) |
| `IC_ZONE` | only without `--subnet-id` | Availability zone (e.g. `us-east-1`) |
| `IBMCLOUD_COS_ACCESS_KEY_ID` | only with S3 `--backed-url` | HMAC access key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_SECRET_ACCESS_KEY` | only with S3 `--backed-url` | HMAC secret key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_ENDPOINT` | no | COS S3 endpoint (defaults to `s3.<region>.cloud-object-storage.appdomain.cloud`) |

## Regional availability

Gaudi 3 instances are available in:

- **us-east** (Washington DC)
- **us-south** (Dallas)
- **eu-de** (Frankfurt)

## Create

```bash
mapt ibmcloud ibm-gaudi create -h
create

Usage:
  mapt ibmcloud ibm-gaudi create [flags]

Flags:
      --conn-details-output string           path to export host connection information (host, username and privateKey)
  -h, --help                                 help for create
      --otel-app-code string                 OpenTelemetry appcode identifier (e.g. MAPT-001); when set together with --otel-auth-token, installs the otelcol-contrib filelog collector on the instance
      --otel-auth-token string               OpenTelemetry authentication token (UUID) used to authenticate against the OTLP endpoint
      --otel-endpoint string                 OTLP HTTP endpoint to export logs to (default "https://otel-input.corp.redhat.com")
      --otel-index string                    Splunk index name for log routing (e.g. rh_linux)
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
| `username` | SSH username (`root`) |
| `id_rsa` | Private key for the instance |

A state folder is also created at `--backed-url`. It is required (together with `--project-name`) to destroy the resources later.

### SSH access

```bash
OUTPUT=/path/to/conn-details-output

ssh -i ${OUTPUT}/id_rsa \
    -o StrictHostKeyChecking=no \
    root@$(cat ${OUTPUT}/host)
```

### Container

```bash
# Using an existing VPC subnet
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-east \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi create \
            --project-name ibm-gaudi \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --subnet-id <subnet-id>

# Auto-provisioning VPC, subnet, and gateway
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-east \
        -e IC_ZONE=us-east-1 \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi create \
            --project-name ibm-gaudi \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```

## OpenTelemetry log collection

When both `--otel-app-code` and `--otel-auth-token` are provided, cloud-init installs `otelcol-contrib` on the instance at first boot and configures it to ship `/var/log/messages`, `/var/log/secure`, and `/var/log/audit/audit.log` to the OTLP endpoint.

```bash
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-east \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi create \
            --project-name ibm-gaudi \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --subnet-id <subnet-id> \
            --otel-app-code MAPT-001 \
            --otel-auth-token <uuid-token>
```

## Using IBM Cloud Object Storage as S3 backend

To store Pulumi state in IBM COS instead of a local file, create [HMAC credentials](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-uhc-hmac-credentials-main) for your COS instance and pass an `s3://` backed URL:

```bash
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-east \
        -e IC_ZONE=us-east-1 \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi create \
            --project-name ibm-gaudi \
            --backed-url s3://my-cos-bucket \
            --conn-details-output /workspace
```

## Destroy

```bash
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-east \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi destroy \
            --project-name ibm-gaudi \
            --backed-url file:///workspace
```

By default, destroy removes the Pulumi state files from the backend after a successful destroy. Use `--keep-state` to preserve them:

```bash
podman run -d --name ibm-gaudi \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IC_REGION=us-east \
        -e IBMCLOUD_COS_ACCESS_KEY_ID=XXX \
        -e IBMCLOUD_COS_SECRET_ACCESS_KEY=XXX \
        quay.io/redhat-developer/mapt:latest ibmcloud ibm-gaudi destroy \
            --project-name ibm-gaudi \
            --backed-url s3://my-cos-bucket \
            --keep-state
```
