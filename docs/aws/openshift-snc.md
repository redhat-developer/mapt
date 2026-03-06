# Overview

This feature of `mapt` allows to deploy a single node OpenShift cluster on the cloud, using bundles from the [SNC](https://github.com/crc-org/snc) project.

## Providers and Platforms

Currently it allows to create the single node cluster on the AWS provider.

## Prerequisite

To be able to create the single node cluster on AWS, an AMI needs to be generated from the [SNC](https://github.com/crc-org/snc) bundle, this can be done with the help of [`cloud-importer`](https://github.com/devtools-qe-incubator/cloud-importer)

To create and publish an AMI for the `4.19.0` OpenShift bundle, use the following command:

```
% cloud-importer openshift-local aws \
    --arch x86_64 \
    --bundle-url https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/bundles/openshift/4.19.0/crc_libvirt_4.19.0_amd64.crcbundle \
    --shasum-url https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/bundles/openshift/4.19.0/sha256sum.txt \
    --backed-url file:///Users/tester/workspace \
    --output /tmp/snc
```

> [!NOTE]
>
> - `cloud-importer` might fail to upload the disk image from the bundle to S3 in case the network is slow, it is often helpful to run `cloud-importer` in an ec2 instance instead of the local dev machine
> - if bundle is already downloaded then use file:///<bundle_path> with `--bundle-url` and file:///<shasum_path> with `--shasum-url`

## Operations

After the AMI is published and accessible by the account, we can use the following `mapt` command to create an OpenShift cluster using spot instances from AWS.

```
% podman run -d --rm \
    -v ${PWD}:/workspace:z \
    -e AWS_ACCESS_KEY_ID=XXX \
    -e AWS_SECRET_ACCESS_KEY=XXX \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:v1.0.0-dev mapt aws openshift-snc create \
        --spot \
        --version 4.19.0 \
        --project-name mapt-snc \
        --backed-url file:///home/tester/workspace \
        --conn-details-output /tmp/snc \
        --pull-secret-file /home/tester/Downloads/pull-secret
```

When `--conn-details-output` is set, the `kubeconfig` is written to disk as soon as the cluster is ready — before any profile deployment begins. This means the kubeconfig is available at `<conn-details-output>/kubeconfig` even if a profile installation fails or times out.

## Profiles

Profiles are optional addons that are installed on the SNC cluster after it is ready. Use the `--profile` flag to enable one or more profiles:

```
mapt aws openshift-snc create \
    --spot \
    --version 4.21.0 \
    --project-name mapt-snc \
    --backed-url file:///home/tester/workspace \
    --conn-details-output /tmp/snc \
    --pull-secret-file /home/tester/Downloads/pull-secret \
    --profile virtualization
```

Multiple profiles can be specified as a comma-separated list (e.g., `--profile virtualization,ai`).

### Available profiles

| Profile | Description |
|---------|-------------|
| `virtualization` | Installs [OpenShift Virtualization](https://docs.openshift.com/container-platform/latest/virt/about_virt/about-virt.html) (CNV) on the cluster, enabling virtual machines to run on the single-node cluster. When this profile is selected, nested virtualization is automatically enabled on the cloud instance. Because standard Nitro-based instances do not expose `/dev/kvm`, a bare metal instance is required.|
| `serverless-serving` | Installs [OpenShift Serverless](https://docs.openshift.com/serverless/latest/about/about-serverless.html) and creates a KnativeServing instance, enabling serverless workloads (Knative Serving) on the cluster.|
| `serverless-eventing` | Installs [OpenShift Serverless](https://docs.openshift.com/serverless/latest/about/about-serverless.html) and creates a KnativeEventing instance, enabling event-driven workloads (Knative Eventing) on the cluster.|
| `serverless` | Installs [OpenShift Serverless](https://docs.openshift.com/serverless/latest/about/about-serverless.html) and creates both KnativeServing and KnativeEventing instances.|
| `servicemesh` | Installs [OpenShift Service Mesh 3](https://docs.openshift.com/service-mesh/latest/about/about-ossm.html) (Sail/Istio) on the cluster, deploying IstioCNI and an Istio control plane.|
| `ai` | Installs [Red Hat OpenShift AI](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed) (RHOAI) on the cluster. Automatically installs Service Mesh v2 (Maistra) and Serverless Serving as prerequisites for Kserve. All three operators install in parallel; the DataScienceCluster CR is only created once Service Mesh and Serverless are fully ready. The minimum instance size is raised to 16 vCPUs (from the default 8) to accommodate the additional operators. **Cannot be combined with the `servicemesh` profile** (which deploys Service Mesh v3/Sail).|


### Adding new profiles

To add a new profile:

1. Create `profile_<name>.go` under `pkg/target/service/snc/` — Go file with a `deploy<Name>()` function that uses the Pulumi Kubernetes provider to create the required resources (Namespace, OperatorGroup, Subscription, CRs, etc.)
2. Register the profile name in `profiles.go` by adding it to `validProfiles` and the `DeployProfiles()` function

