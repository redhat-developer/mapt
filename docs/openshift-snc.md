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
        --backed-url file:///Users/tester/workspace \
        --conn-details-output /tmp/snc \
        --pull-secret-file /Users/tester/Downloads/pull-secret
```

After the above command succeeds the `kubeconfig` to access the deployed cluster will be available in `/tmp/snc/kubeconfig`

