<div align="center">

# ![mapt](./docs/logo/mapt.svg)

### Multi Architecture Provisioning Tool

**Spin up cloud machines in seconds. Tear them down just as fast.**
Spot pricing. Airgap topologies. CI/CD native. Built for operators.

[![Build](https://github.com/redhat-developer/mapt/actions/workflows/build-go.yaml/badge.svg)](https://github.com/redhat-developer/mapt/actions/workflows/build-go.yaml)
[![OCI](https://github.com/redhat-developer/mapt/actions/workflows/build-oci.yaml/badge.svg)](https://github.com/redhat-developer/mapt/actions/workflows/build-oci.yaml)
[![License](https://img.shields.io/github/license/redhat-developer/mapt)](LICENSE)

</div>

---

## What is mapt?

`mapt` is a command-line tool for provisioning and destroying cloud environments across **AWS**, **Azure**, and **IBM Cloud**. It wraps multi-cloud infrastructure into a single, consistent interface — optimized for cost, speed, and CI/CD integration.

```
mapt <provider> <target> <create|destroy> [flags]
```

One pattern. Every cloud. Every OS.

---

## Quickstart

Pull the container and provision a Fedora machine on AWS spot:

```bash
podman run -d --name mapt-fedora \
    -v ${PWD}:/workspace:z \
    -e AWS_ACCESS_KEY_ID=<key> \
    -e AWS_SECRET_ACCESS_KEY=<secret> \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:latest aws fedora create \
        --project-name my-fedora \
        --backed-url file:///workspace \
        --conn-details-output /workspace \
        --spot
```

Connection details land at `${PWD}/host`, `${PWD}/username`, and `${PWD}/id_rsa`. Destroy with the same flags, swapping `create` for `destroy`.

---

## What can you provision?

### Instances

| Target | AWS | Azure | IBM Cloud |
|--------|:---:|:-----:|:---------:|
| **macOS** (x86, M1, M2) | [docs](docs/aws/mac.md) | — | — |
| **Windows Server** | [docs](docs/aws/windows.md) | — | — |
| **Windows Desktop** | — | [docs](docs/azure/windows.md) | — |
| **RHEL** | [docs](docs/aws/rhel.md) | [docs](docs/azure/rhel.md) | — |
| **RHEL AI** | [docs](docs/aws/rhelai.md) | [docs](docs/azure/rhelai.md) | — |
| **Fedora** | [docs](docs/aws/fedora.md) | [docs](docs/azure/fedora.md) | — |
| **Ubuntu** | — | [docs](docs/azure/ubuntu.md) | — |
| **IBM Z** (s390x) | — | — | [docs](docs/ibmcloud/ibm-z.md) |
| **IBM Power** (ppc64le) | — | — | [docs](docs/ibmcloud/ibm-power.md) |

### Services

| Service | AWS | Azure | IBM Cloud | Description |
|---------|:---:|:-----:|:---------:|-------------|
| **Kind** | [docs](docs/aws/kind.md) | [docs](docs/azure/kind.md) | [docs](docs/ibmcloud/kind.md) | Lightweight Kubernetes via Kind |
| **EKS** | [docs](docs/aws/eks.md) | — | — | Managed Kubernetes with spot node groups |
| **AKS** | — | [docs](docs/azure/aks.md) | — | Managed Kubernetes |
| **OpenShift SNC** | [docs](docs/aws/openshift-snc.md) | — | — | Single-node OpenShift for testing |
| **Mac-Pool** | [docs](docs/aws/mac-pool.md) | — | — | Shared Mac host pool — amortize the 24h minimum |

### Architectures

| Architecture | Providers |
|---|---|
| x86_64 | AWS, Azure, IBM Cloud |
| arm64 | AWS, Azure |
| s390x | IBM Cloud |
| ppc64le | IBM Cloud |

---

## Key features

### Spot-optimized provisioning

mapt scans placement scores and pricing across all regions to find the best **cost vs. availability** balance — no manual region hunting. If a region doesn't have the instance you need, mapt falls back automatically.

```bash
mapt aws rhel create --spot \
    --project-name my-rhel --backed-url file:///workspace \
    --conn-details-output /workspace
```

### Hardware-spec instance selection

Describe the machine you need; mapt picks the right instance type:

```bash
mapt azure fedora create \
    --arch arm64 --cpus 4 --memory 16 \
    --project-name fedora-arm --backed-url file:///workspace \
    --conn-details-output /workspace
```

Flags: `--arch`, `--cpus`, `--memory`, `--nested-virt`, `--compute-sizes`
Details: [instance selection docs](docs/instance-selection.md)

### Airgap topology

Provision an isolated machine behind a jump bastion. mapt wires up the full network — you get bastion connection details alongside the target host.

```bash
mapt aws rhel create --airgap \
    --project-name rhel-airgap --backed-url file:///workspace \
    --conn-details-output /workspace
```

Outputs: `host`, `username`, `id_rsa`, `bastion_host`, `bastion_username`, `bastion_id_rsa`

### Self-destruct timer (serverless mode)

Set `--timeout` and mapt will tear itself down automatically if the destroy never runs — pipeline crash, lost state, whatever. No orphaned resources, no surprise bills.

Details: [serverless mode docs](docs/serverless-mode.md)

---

## CI/CD integrations

mapt machines register themselves with your CI system at provision time — nothing to configure after the fact.

### GitHub Actions self-hosted runner

```bash
mapt aws fedora create --spot \
    --install-ghactions-runner \
    --ghactions-runner-repo "https://github.com/your-org/your-repo" \
    --ghactions-app-id "123456" \
    --ghactions-app-installation-id "789012" \
    --ghactions-app-private-key "/path/to/private-key.pem" \
    --project-name fedora-runner --backed-url file:///workspace \
    --conn-details-output /workspace
```

Auth methods: GitHub App (recommended), PAT, or pre-generated registration token.
Supported targets: AWS (Windows, RHEL, Fedora, macOS) · Azure (Windows, RHEL) · IBM Cloud (Power, Z)
Details: [self-hosted runner docs](docs/self-hosted-runner.md)

### GitLab Runner

```bash
mapt aws fedora create --spot \
    --glrunner-token <token> \
    --project-name fedora-gitlab --backed-url file:///workspace \
    --conn-details-output /workspace
```

Supported targets: AWS (Windows, RHEL, Fedora, macOS) · Azure (Windows, RHEL) · IBM Cloud (Power, Z)
Details: [GitLab runner docs](docs/gitlab-runner.md)

### Tekton tasks

Tekton tasks for dynamic provisioning inside pipelines are available in the [`tkn/`](tkn) directory.

---

## Running mapt

### Container (recommended)

```bash
podman run -d --name mapt \
    -v ${PWD}:/workspace:z \
    -e AWS_ACCESS_KEY_ID=<key> \
    -e AWS_SECRET_ACCESS_KEY=<secret> \
    -e AWS_DEFAULT_REGION=us-east-1 \
    quay.io/redhat-developer/mapt:latest aws fedora create \
        --project-name my-env \
        --backed-url file:///workspace \
        --conn-details-output /workspace
```

The `--backed-url` volume mount holds your stack state — keep it, you need it to destroy.

### Binary

```bash
go install github.com/redhat-developer/mapt/cmd/mapt@latest
mapt --help
```

---

## State management

mapt uses [Pulumi](https://www.pulumi.com/) under the hood. Stack state is stored at `--backed-url`:

- **Local**: `file:///absolute/path` — simplest, works for local dev
- **S3**: `s3://your-bucket` — required for serverless mode and shared CI
- **Azure Blob**: `azblob://your-container`

The `--project-name` flag namespaces stacks, so you can run multiple environments from the same backend.

---

<div align="center">

**[AWS docs](docs/aws.md)** · **[Azure docs](docs/azure)** · **[IBM Cloud docs](docs/ibmcloud)** · **[Changelog](CHANGELOG.md)**

</div>
