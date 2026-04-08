# Overview

This actions will handle provision for EKS cluster instances on AWS.

## EKS

It creates / destroy a basic EKS Cluster. The cluster is provisioned with a managed node group, IAM roles, an OIDC provider for IRSA (IAM Roles for Service Accounts), and optional add-ons like the AWS Load Balancer Controller and EBS CSI driver.

### Prerequisites

* AWS credentials with permissions to create EKS clusters, VPCs, IAM roles, and related resources
* Credentials exported as environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`

### Architecture

The following AWS resources are created as part of the EKS stack:

* **VPC** with CIDR `10.0.0.0/16` and public subnets across multiple availability zones
* **VPC Endpoints** for S3, ECR, and SSM (to support private networking)
* **NAT Gateway** in single mode
* **EKS Cluster** with public API endpoint access
* **Managed Node Group** with configurable instance types and scaling
* **IAM Roles** for the EKS service and node groups, including required AWS managed policies
* **OIDC Provider** for enabling pod-level IAM roles via service accounts (IRSA)

### Operations

#### Create

This will create a K8s cluster using EKS according to params specified:

```bash
mapt aws eks create -h
create

Usage:
  mapt aws eks create [flags]

Flags:
      --addons strings                   List of EKS addons to be installed, separated by commas.
      --arch string                      architecture for the machine. Allowed x86_64 or arm64 (default "x86_64")
      --compute-sizes strings            Comma seperated list of sizes for the machines to be requested. If set this takes precedence over compute by args
      --conn-details-output string       path to export host connection information (host, username and privateKey)
      --cpus int32                       Number of CPUs for the cloud instance (default 8)
      --excluded-zone-ids strings        Comma-separated list of zone IDs to exclude from availability zone selection
      --gpu-manufacturer string          Manufacturer company name for GPU. (i.e. NVIDIA)
      --gpus int32                       Number of GPUs for the cloud instance
  -h, --help                             help for create
      --load-balancer-controller         Install AWS Load Balancer Controller
      --memory int32                     Amount of RAM for the cloud instance in GiB (default 64)
      --nested-virt                      Use cloud instance that has nested virtualization support
      --spot                             if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)
      --spot-eviction-tolerance string   if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest (default "lowest")
      --spot-excluded-regions strings    Comma-separated list of zone IDs to exclude from spot selection
      --spot-increase-rate int           Percentage to be added on top of the current calculated spot price to increase chances to get the machine (default 30)
      --tags stringToString              tags to add on each resource (--tags name1=value1,name2=value2) (default [])
      --version string                   EKS K8s cluster version (default "1.31")
      --workers-desired string           Worker nodes scaling desired size (default "1")
      --workers-max string               Worker nodes scaling maximum size (default "3")
      --workers-min string               Worker nodes scaling minimum size (default "1")

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

##### Add-ons

The `--addons` flag accepts any standard EKS add-on name. The `aws-ebs-csi-driver` add-on receives special handling: when installed, mapt automatically creates a dedicated IAM role via OIDC and configures it as the default storage class.

##### AWS Load Balancer Controller

When `--load-balancer-controller` is set, mapt deploys the AWS Load Balancer Controller via Helm. This includes automatic IRSA configuration with the full IAM policy required for ALB/NLB management.

##### Excluded Zone IDs

Some AWS availability zones are incompatible with EKS (see [AWS knowledge center](https://repost.aws/knowledge-center/eks-cluster-creation-errors)). By default, known incompatible zones are excluded. Use `--excluded-zone-ids` to specify additional zones to exclude.

### Outputs

The create operation produces the following output at the path defined by `--conn-details-output`:

* **kubeconfig**: kubeconfig file for connecting to the EKS cluster

The kubeconfig uses AWS CLI exec-based authentication. To use it, you need the AWS CLI installed and the following environment variables exported: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`.

The state for the created resources is stored at the path defined by `--backed-url`. This state is required (with the same `--project-name`) to destroy the resources later.

#### Destroy

```bash
mapt aws eks destroy -h
destroy

Usage:
  mapt aws eks destroy [flags]

Flags:
      --force-destroy   if force-destroy is set the command will destroy even if there is a lock.
  -h, --help            help for destroy
      --keep-state      keep Pulumi state files in S3 backend after successful destroy (by default, state files are removed)

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Container

When running the container image it is required to pass the authentication information as variables, following a sample snippet on how to create
a cluster with default values:

```bash
podman run -d --rm \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-2 \
        quay.io/redhat-developer/mapt:v1.0.0-dev aws eks create \
            --project-name "mapt-eks" \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```

Creating an EKS cluster with spot instances, add-ons, and the Load Balancer Controller:

```bash
podman run -d --rm \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-2 \
        quay.io/redhat-developer/mapt:1.0.0-dev aws eks create \
            --project-name "mapt-eks" \
            --backed-url file:///workspace \
            --conn-details-output /workspace \
            --spot \
            --addons aws-ebs-csi-driver \
            --load-balancer-controller
```

It is important to notice that backed-url contains the state of the resources on AWS linked to the EKS creation, to destroy we need to reference them as so we should pass same folder as we set for creation:

```bash
podman run -d --rm \
        -v ${PWD}:/workspace:z \
        -e AWS_ACCESS_KEY_ID=XXX \
        -e AWS_SECRET_ACCESS_KEY=XXX \
        -e AWS_DEFAULT_REGION=us-east-2 \
    quay.io/redhat-developer/mapt:v1.0.0-dev aws eks destroy \
            --project-name "mapt-eks" \
            --backed-url file:///workspace
```
