# Overview

This actions will handle provision for EKS cluster instances on AWS.

## EKS

It creates and destroys an EKS cluster with a self-managed Auto Scaling worker group, IAM roles, an OIDC provider for IRSA (IAM Roles for Service Accounts), and optional add-ons such as the AWS Load Balancer Controller and EBS CSI driver.

### Prerequisites

* AWS credentials with permissions to create EKS clusters, VPCs, IAM roles, and related resources
* Credentials exported as environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`

### Architecture

The following AWS resources are created as part of the EKS stack:

* **VPC** with CIDR `10.0.0.0/16` and public subnets across multiple availability zones
* Optional **VPC Endpoints**, selected through the network flags
* No NAT gateway; workers use public subnets
* **EKS Cluster** with public API endpoint access
* **Self-managed worker group** using an AL2023 launch template and Auto Scaling, with configurable instance types and scaling
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

The chart is pinned to version 3.6.0 and installed using Pulumi Helm v4. Custom resources depend on their CRDs becoming established, avoiding discovery races during a fresh install. Chart upgrades require a review of the controller IAM policy.

##### Excluded Zone IDs

Some AWS availability zones are incompatible with EKS (see [AWS knowledge center](https://repost.aws/knowledge-center/eks-cluster-creation-errors)). By default, known incompatible zones are excluded. Use `--excluded-zone-ids` to specify additional zones to exclude.

### Outputs

The create operation produces the following output at the path defined by `--conn-details-output`:

* **kubeconfig**: kubeconfig file for connecting to the EKS cluster

The kubeconfig uses AWS CLI exec-based authentication. To use it, you need the AWS CLI installed and credentials available through the AWS credential chain (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`, a shared profile, or workload credentials).

The kubeconfig records the cluster region explicitly. The internal Pulumi provider uses an SDK-generated token minted after workers and core add-ons are ready, so mapt itself does not require the AWS CLI binary.

The state for the created resources is stored at the path defined by `--backed-url`. This state is required (with the same `--project-name`) to destroy the resources later.

The `eksCleanup` stack output records the schema version, AWS account ID, region, cluster name and VPC ID for teardown. Tags supplied through `--tags` propagate to the cluster security group, load balancer controller resources, CSI volumes and CNI-created ENIs. These also receive `mapt-cluster-name=<cluster name>`, reserved for exact cluster ownership; supplying `build-id=<BUILD_ID>` links resources to a particular run. EKS control-plane ENIs retain AWS-native ownership tags. CSI volume and CNI ENI tags require `aws-ebs-csi-driver` and `vpc-cni` in `--addons`.

#### Destroy

Controller-created resources are identified by exact cluster ownership, not by the shared `projectName=eks` tag. Load balancers, target groups and security groups are matched through the load balancer controller tags or the in-tree `kubernetes.io/cluster/<name>` tag, which also covers classic ELBs and NLBs created from Service annotations. CSI volumes require both `mapt-cluster-name` and `ebs.csi.aws.com/cluster=true`; deleting them includes volumes left behind by `Retain` persistent volumes. Attached volumes and interfaces are never forcibly detached. Other interfaces in the VPC, including requester-managed ones, are never deleted; cleanup waits for AWS to release them. Interface VPC endpoint and NAT gateway interfaces belong to the stack and are removed in the final network phase.

Destroy first removes the Kubernetes resources mapt manages and their provider, then the cluster and its dependents, then external controller resources, and finally the network and remaining IAM resources. Kubernetes resources go first because Pulumi cannot target the cluster while a provider whose configuration depends on it is still in the stack ([pulumi/pulumi#14650](https://github.com/pulumi/pulumi/issues/14650)). Each phase retains state on failure. Retrying the same command resumes cleanup, including after the cluster has gone. Independent resource deletions continue after errors.

The `eksCleanup` output written with the cluster provides the account, region, cluster name and VPC used by the cleanup phases. If create was interrupted before the output was saved, the same values are read from the cluster resource. Kubernetes resources are removed before the cluster; when the provider can no longer reach it, for example after its token expired, they are dropped from state because deleting the cluster removes them anyway. `--force-destroy` only removes stale backend locks; it does not force-detach AWS resources.

Cleanup requires `sts:GetCallerIdentity`, `eks:DescribeCluster`, EC2 describe/delete permissions for ENIs, volumes and security groups, EC2 security-group rule revocation permissions, ELBv2 describe/tag-read/delete permissions for load balancers and target groups, and Elastic Load Balancing describe/tag-read/delete permissions for classic load balancers. It verifies the AWS account and cluster absence before external-resource deletion. Each cleanup phase waits up to ten minutes, then reports resource IDs and retains the checkpoint for retry.

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
