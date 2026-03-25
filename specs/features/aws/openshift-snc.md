# Spec: AWS OpenShift Single Node Cluster (SNC)

## Context
Provisions a single-node OpenShift cluster (CRC/SNC) on an EC2 instance using a pre-baked AMI.
Entry point: `pkg/provider/aws/action/snc/`. Profile system: `pkg/target/service/snc/profile/`.
CLI: `cmd/mapt/cmd/aws/services/snc.go`.

The cluster setup runs inside cloud-init on boot. Sensitive values (pull secret, kubeadmin
password, developer password) are managed via AWS SSM Parameter Store. Readiness is verified
by SSH-checking the kubeconfig availability and CA rotation completion.

## Problem
This feature is implemented. This spec documents behaviour, the profile system, and gaps.

## Requirements
- [ ] Provision an EC2 instance using the SNC pre-baked AMI (looked up by version + arch)
- [ ] Fail early with a clear error if the AMI does not exist in the target region
- [ ] Store pull secret, kubeadmin password, and developer password in SSM; inject via cloud-init
- [ ] Verify cluster readiness: SSH up → kubeconfig exists → CA rotation complete
- [ ] Export kubeconfig (with public IP replacing internal API endpoint) as a secret output
- [ ] Support optional profiles deployed post-cluster-ready via the Kubernetes Pulumi provider:
  - `virtualization` — enables nested virtualisation on the compute instance
  - `serverless-serving` — installs Knative Serving
  - `serverless-eventing` — installs Knative Eventing
  - `serverless` — installs both Knative Serving and Eventing
  - `servicemesh` — installs OpenShift Service Mesh 3
- [ ] Validate profile names before provisioning begins
- [ ] Support spot allocation and serverless self-destruct timeout
- [ ] Write output files: `host`, `username`, `id_rsa`, `kubeconfig`, `kubeadmin-password`, `developer-password`
- [ ] `destroy` cleans up main stack, spot stack, S3 state

## Out of Scope
- Multi-node OCP (full IPI/UPI install)
- EKS (see `006-aws-eks.md`)

## Affected Areas
- `pkg/provider/aws/action/snc/` — orchestration, kubeconfig extraction
- `pkg/target/service/snc/` — cloud-config, SSM management, readiness commands
- `pkg/target/service/snc/profile/` — profile registry and deployment
- `cmd/mapt/cmd/aws/services/snc.go`
- `tkn/template/infra-aws-ocp-snc.yaml`

## Known Gaps / Improvement Ideas
- Profile deployment failures are logged as warnings, not errors (`snc.go:279`)
  — consider making this configurable (fail-fast vs warn-and-continue)
- `disableClusterReadiness` flag skips the readiness wait entirely; useful for debugging
  but not documented in the Tekton task
- The `--version` flag accepts a free-form string; no validation against available AMIs beyond
  the early existence check

## Acceptance Criteria
- Cluster is reachable via the exported kubeconfig
- `oc get nodes` shows one Ready node
- Profiles deploy successfully when specified
- `mapt aws openshift-snc destroy` removes all resources and state

---

## Command

```
mapt aws openshift-snc create  [flags]
mapt aws openshift-snc destroy [flags]
```

### Shared flag groups (`specs/cmd/params.md`)

| Group | Flags added |
|---|---|
| Common | `--project-name`, `--backed-url` |
| Compute Request | `--cpus`, `--memory`, `--arch`, `--nested-virt`, `--compute-sizes` |
| Spot | `--spot`, `--spot-eviction-tolerance`, `--spot-increase-rate`, `--spot-excluded-regions` |

Note: no integration flags.

### Target-specific flags (create only)

| Flag | Type | Default | Description |
|---|---|---|---|
| `--version` | string | `4.21.0` | OpenShift version |
| `--arch` | string | `x86_64` | `x86_64` or `arm64` |
| `--pull-secret-file` | string | — | Path to Red Hat pull secret JSON file (required) |
| `--snc` | []string | — | SNC profiles to apply (comma-separated) |
| `--disable-cluster-readiness` | bool | false | Skip cluster readiness check after provision |
| `--timeout` | string | — | Self-destruct duration |
| `--conn-details-output` | string | — | Path to write kubeconfig |
| `--tags` | map | — | Resource tags |

### Destroy flags

`--serverless`, `--force-destroy`, `--keep-state`

### Action args struct populated

`snc.SNCArgs` → `pkg/provider/aws/action/snc/snc.go`
