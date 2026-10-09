# Overview

This action provisions a single-node OpenShift cluster on IBM Cloud VPC, using bundles from the [SNC](https://github.com/crc-org/snc) project. The VM is created from a catalog image and assigned a floating IP for direct SSH and Kubernetes API access.

The catalog image is automatically discovered by querying the IBM Cloud Catalog Management API. When `--version` is omitted, the latest available version is selected by semver.

## Networking

By default a new VPC, subnet, and public gateway are created. When `--vpc-id` is provided, mapt reuses the existing VPC and only provisions a new subnet inside it — useful when the account is near the VPC quota limit.

## Prerequisite

A catalog image must be generated from the [SNC](https://github.com/crc-org/snc) bundle and published to an IBM Cloud private catalog. This can be done with [`cloud-importer`](https://github.com/devtools-qe-incubator/cloud-importer):

```bash
cloud-importer openshift-local ibmcloud \
    --arch x86_64 \
    --bundle-url https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/bundles/openshift/4.22.14/crc_libvirt_4.22.14_amd64.crcbundle \
    --shasum-url https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/bundles/openshift/4.22.14/sha256sum.txt \
    --backed-url file:///Users/tester/workspace \
    --output /tmp/snc
```

The resulting catalog image must be named following the pattern `openshift-local-<version>-<arch>` (e.g. `openshift-local-4-22-14-x86-64`) for auto-discovery to work.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `IBMCLOUD_API_KEY` | yes | IBM Cloud API key |
| `IC_REGION` | yes | IBM Cloud region (e.g. `us-south`, `eu-de`) |
| `IC_ZONE` | yes | Availability zone (e.g. `eu-de-2`) |
| `IBMCLOUD_COS_ACCESS_KEY_ID` | only with S3 `--backed-url` | HMAC access key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_SECRET_ACCESS_KEY` | only with S3 `--backed-url` | HMAC secret key for IBM Cloud Object Storage |
| `IBMCLOUD_COS_ENDPOINT` | no | COS S3 endpoint (defaults to `s3.<region>.cloud-object-storage.appdomain.cloud`) |

## Create

```bash
mapt ibmcloud openshift-snc create \
    --project-name my-snc \
    --backed-url file:///workspace \
    --conn-details-output /tmp/snc \
    --version 4.22.14 \
    --pull-secret-file /path/to/pull-secret
```

Reusing an existing VPC:

```bash
mapt ibmcloud openshift-snc create \
    --project-name my-snc \
    --backed-url file:///workspace \
    --conn-details-output /tmp/snc \
    --version 4.22.14 \
    --pull-secret-file /path/to/pull-secret \
    --vpc-id <vpc-id>
```

With spot instances:

```bash
mapt ibmcloud openshift-snc create \
    --spot \
    --version 4.22.14 \
    --project-name my-snc \
    --backed-url file:///workspace \
    --conn-details-output /tmp/snc \
    --pull-secret-file /path/to/pull-secret
```

### Outputs

Files written to the path defined by `--conn-details-output`:

| File | Description |
|---|---|
| `host` | Floating IP of the instance |
| `username` | SSH username (`core`) |
| `id_rsa` | Private key for the instance |
| `kubeconfig` | Kubeconfig for the cluster (API server points to the floating IP via nip.io) |

### Container

```bash
podman run -d --name ibmcloud-snc \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=eu-de \
        -e IC_ZONE=eu-de-2 \
        quay.io/redhat-developer/mapt:v1.0.0-dev mapt ibmcloud openshift-snc create \
            --project-name ibmcloud-snc \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --version 4.22.14 \
            --pull-secret-file /workspace/pull-secret
```

### SSH access

```bash
OUTPUT=/path/to/conn-details-output

ssh -i ${OUTPUT}/id_rsa \
    -o StrictHostKeyChecking=no \
    core@$(cat ${OUTPUT}/host)
```

### Kubernetes access

```bash
OUTPUT=/path/to/conn-details-output

export KUBECONFIG=${OUTPUT}/kubeconfig
oc get nodes
```

## Profiles

Profiles are optional addons installed on the cluster after it is ready. They work identically to the [AWS SNC profiles](../aws/openshift-snc.md#profiles):

```bash
mapt ibmcloud openshift-snc create \
    --version 4.22.14 \
    --project-name my-snc \
    --backed-url file:///workspace \
    --conn-details-output /tmp/snc \
    --pull-secret-file /path/to/pull-secret \
    --profile virtualization
```

See the [AWS SNC docs](../aws/openshift-snc.md#profiles) for the full list of available profiles and operator override flags (`--operator-channel`, `--catalog-source`).

## Destroy

```bash
podman run -d --name ibmcloud-snc \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=eu-de \
        quay.io/redhat-developer/mapt:v1.0.0-dev mapt ibmcloud openshift-snc destroy \
            --project-name ibmcloud-snc \
            --backed-url file:///workspace
```
