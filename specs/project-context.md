# mapt — Project Context

## What This Project Is

mapt (Multi Architecture Provisioning Tool) is a Go CLI that provisions ephemeral compute environments
across AWS and Azure using the Pulumi Automation API. It is used primarily by CI/CD pipelines that
need on-demand remote machines of specific OS/arch combinations.

Key design goals:
- **Cost savings**: prefer spot instances with cross-region best-bid selection
- **Speed**: use AMI fast-launch, root volume replacement, pre-baked images
- **Safety**: self-destruct via serverless scheduled tasks (timeout mode)
- **Integration**: emit connection details (host, username, key/password) as output files consumed by CI systems

## Repository Layout

```
cmd/mapt/cmd/          CLI commands (Cobra), one file per target
  params/              Shared flag definitions, Add*Flags helpers, *Args() readers
                       — see specs/cmd/params.md
  aws/hosts/           AWS host subcommands (rhel, windows, fedora, mac, rhelai)
  aws/services/        AWS service subcommands (eks, kind, mac-pool, snc)
  azure/hosts/         Azure host subcommands (rhel, windows, linux, rhelai)
  azure/services/      Azure service subcommands (aks, kind)

pkg/manager/           Pulumi Automation API wrapper
  context/             Context type — carries project/run metadata, integrations
  credentials/         Provider credential helpers

pkg/provider/
  api/                 Shared API types and interfaces (ComputeRequest, SpotArgs, SpotSelector, ComputeSelector, CloudConfig)
                       — see specs/api/provider-interfaces.md
  aws/
    action/            Entry points per target: Create(), Destroy() orchestrate stacks
    modules/           Reusable Pulumi stack components
      allocation/      Spot vs on-demand region/AZ selection
      ami/             AMI copy + fast-launch
      bastion/         Bastion host for airgap scenarios
      ec2/compute/     EC2 instance resource
      iam/             IAM roles/policies
      mac/             Mac dedicated host + machine lifecycle
      network/         Standard and airgap VPC/subnet/LB
      serverless/      ECS Fargate scheduled self-destruct
      spot/            Best-spot-option Pulumi stack
    data/              AWS SDK read-only queries (AMI, AZ, spot price, etc.)
    services/          Low-level Pulumi resource wrappers (keypair, SG, S3, SSM, VPC)
  azure/
    action/            Entry points per target
    modules/           Azure network, VM, allocation
    data/              Azure SDK queries
    services/          Azure Pulumi resource wrappers
  util/                Shared: command readiness, output writing, security, windows helpers

pkg/integrations/      CI system integration snippets
  github/              GitHub Actions self-hosted runner
  cirrus/              Cirrus CI persistent worker
  gitlab/              GitLab runner

pkg/target/            Cloud-init / userdata builders per OS target
  host/rhel/           RHEL cloud-config (base + SNC variant)
  host/fedora/         Fedora cloud-config
  host/rhelai/         RHEL AI API wrapper
  host/windows-server/ Windows PowerShell userdata
  service/kind/        Kind cloud-config
  service/snc/         OpenShift SNC cloud-config + profile deployment
    profile/           SNC profiles: virtualization, serverless, servicemesh

pkg/util/              Generic utilities (cache, cloud-init, file, logging, maps, network, slices)

tkn/                   Tekton Task YAML files (generated from tkn/template/ by make tkn-update)
docs/                  User-facing documentation per target
specs/                 Developer/contributor artifacts
  project-context.md  Project knowledge base (this file)
  features/           Feature specifications
```

## Key Types

```go
// manager/context.ContextArgs — input to every action Create()/Destroy()
type ContextArgs struct {
    ProjectName   string
    BackedURL     string          // "s3://bucket/path" or "file:///local/path"
    ResultsOutput string          // directory where output files are written
    Serverless    bool            // use role-based credentials (ECS task context)
    ForceDestroy  bool
    KeepState     bool
    Tags          map[string]string
    GHRunnerArgs  *github.GithubRunnerArgs   // optional integration
    CirrusPWArgs  *cirrus.PersistentWorkerArgs
    GLRunnerArgs  *gitlab.GitLabRunnerArgs
}

// manager.Stack — describes a Pulumi stack to run
type Stack struct {
    ProjectName         string
    StackName           string
    BackedURL           string
    DeployFunc          pulumi.RunFunc
    ProviderCredentials credentials.ProviderCredentials
}

// provider/aws/modules/allocation.AllocationResult — result of spot/on-demand selection
type AllocationResult struct {
    Region        *string
    AZ            *string
    SpotPrice     *float64   // nil if on-demand
    InstanceTypes []string
}
```

## Module Reuse Contract

**This is the most important architectural rule in mapt.**

Logic that exists in a module MUST be reused, never reimplemented. The layers are:

- `modules/` — reusable Pulumi stack components. Always call these; never inline their logic into an action.
- `services/` — low-level Pulumi resource wrappers. Always use these; never call Pulumi provider resources directly from an action.
- `data/` — read-only cloud API queries. Always use these; never call AWS/Azure SDKs directly from an action.
- `action/` — the only layer allowed to contain orchestration logic specific to a single target.

When writing a spec or implementing a feature, explicitly list which existing modules are called
(Must Reuse) separately from which new files are created (Must Create). This is the distinction
the spec template enforces.

### AWS EC2 Host — Mandatory Module Sequence

Every AWS EC2-based host target calls these modules in this order. Deviation requires justification.

**`Create()` function:**
```
mc.Init(mCtxArgs, aws.Provider())
allocation.Allocation(mCtx, &AllocationArgs{...})   // spot or on-demand
r.createMachine() | r.createAirgapMachine()
```

**`deploy()` Pulumi RunFunc — always in this order:**
```
amiSVC.GetAMIByName()                    // AMI lookup
network.Create()                         // VPC, subnet, IGW, optional LB, optional airgap
keypair.KeyPairRequest.Create()          // TLS keypair → export <prefix>-id_rsa
securityGroup.SGRequest.Create()         // security group with ingress rules
<target cloud-config builder>.Generate() // cloud-init / userdata
compute.ComputeRequest.NewCompute()      // EC2 instance
serverless.OneTimeDelayedTask()          // only when Timeout != ""
c.Readiness()                            // remote command readiness check
```

**`Destroy()` function — always in this order:**
```
aws.DestroyStack()
spot.Destroy() guarded by spot.Exist()   // only if spot was used
amiCopy.Destroy() guarded by amiCopy.Exist()  // only if AMI copy was needed (Windows)
aws.CleanupState()
```

**`manageResults()` function:**
```
bastion.WriteOutputs()   // only when airgap=true
output.Write()           // always — writes host/username/key files
```

**Naming — non-negotiable:**
```
resourcesUtil.GetResourceName(prefix, componentID, suffix)  // all resource names
mCtx.StackNameByProject(stackName)                          // all Pulumi stack names
```

### AWS EC2 Host — Files to Create (only these)

For each new AWS EC2 target, exactly these files are created — everything else is reused:

```
pkg/provider/aws/action/<target>/<target>.go    // Args struct, Create, Destroy, deploy, manageResults, securityGroups
pkg/provider/aws/action/<target>/constants.go   // stackName, componentID, AMI regex, disk size, ports
pkg/target/host/<target>/                       // cloud-config or userdata builder
cmd/mapt/cmd/aws/hosts/<target>.go             // Cobra create/destroy subcommands
tkn/template/infra-aws-<target>.yaml           // Tekton task template
```

### Azure VM Host — Mandatory Module Sequence

**`Create()` function:**
```
mc.Init(mCtxArgs, azure.Provider())
allocation.Allocation(mCtx, &AllocationArgs{...})   // azure spot or on-demand
```

**`deploy()` Pulumi RunFunc:**
```
azure resource group
azure/modules/network.Create()           // VNet, subnet, NIC, optional public IP
keypair or password generation
azure/services/network/security-group.SGRequest.Create()
virtualmachine.NewVM()                   // Azure VM resource
readiness check via remote command
```

**`Destroy()` function:**
```
azure.DestroyStack()
azure.CleanupState()
```

### Adding a New AWS Host Target

1. **Args struct** in `pkg/provider/aws/action/<target>/<target>.go`
   - Embed `*cr.ComputeRequestArgs`, `*spotTypes.SpotArgs`
   - Include `Prefix`, `Airgap bool`, `Timeout string`

2. **`Create()`**: `mc.Init` → `allocation.Allocation` → `createMachine` or `createAirgapMachine`

3. **`deploy()`**: follow the mandatory module sequence above exactly

4. **`Destroy()`**: follow the mandatory destroy sequence above exactly

5. **`manageResults()`**: `bastion.WriteOutputs` (if airgap) then `output.Write`

6. **Cobra command** in `cmd/mapt/cmd/aws/hosts/<target>.go`
   - Subcommands: `create`, `destroy`; bind all flags

7. **Tekton template** in `tkn/template/infra-aws-<target>.yaml`

### Airgap Orchestration

Two-phase stack update on the same stack:
1. `airgapPhaseConnectivity = network.ON` — creates NAT gateway, bootstraps machine
2. `airgapPhaseConnectivity = network.OFF` — removes NAT gateway, machine loses egress

### Spot vs On-Demand (Allocation Module)

`allocation.Allocation()` is the single entry point. It:
- If `Spot.Spot == true`: creates/reuses a `spotOption` Pulumi stack that selects best region + AZ + price
- If on-demand: uses the provider's default region, iterates AZs until instance types are available

The spot stack is idempotent — if it already exists, outputs are reused (region stays stable across re-creates).

### Serverless Self-Destruct

`serverless.OneTimeDelayedTask()` creates an AWS EventBridge Scheduler + Fargate task that runs
`mapt <target> destroy` at `now + timeout`. Requires a remote BackedURL (not `file://`).

### Integration Snippets

Each integration (`github`, `cirrus`, `gitlab`) implements `IntegrationConfig`:
- `GetUserDataValues()` returns token, repo URL, labels, etc.
- `GetSetupScriptTemplate()` returns an embedded shell/PowerShell script template
- Called from cloud-config / userdata builders in `pkg/target/`

### SNC Profiles

Profiles are registered in `pkg/target/service/snc/profile/profile.go`:
- `virtualization` — enables nested virt on the compute instance
- `serverless-serving`, `serverless-eventing`, `serverless` — Knative
- `servicemesh` — OpenShift Service Mesh 3

`profile.RequireNestedVirt()` gates the instance type selection.
`profile.Deploy()` installs operators/CRDs via the Pulumi Kubernetes provider post-cluster-ready.

## Build & Test Commands

```bash
make build          # compile to out/mapt
make install        # go install to $GOPATH/bin
make test           # go test -race ./pkg/... ./cmd/...
make lint           # golangci-lint
make fmt            # gofmt
make check          # build + test + lint + renovate-check
make oci-build      # container image (amd64 + arm64)
make tkn-update     # regenerate tkn/*.yaml from templates
make tkn-push       # push Tekton bundle
```

## Naming Conventions

- Resource names: `resourcesUtil.GetResourceName(prefix, componentID, suffix)`
  e.g. `GetResourceName("main", "aws-rhel", "sg")` → `"main-aws-rhel-sg"`
- Stack names: `mCtx.StackNameByProject(stackName)` → `"<stackName>-<projectName>"`
- Output keys: `"<prefix>-host"`, `"<prefix>-username"`, `"<prefix>-id_rsa"`, `"<prefix>-userpassword"`
- Constants: defined in `constants.go` / `contants.go` next to the action file

## State Backend

Pulumi state is stored at `BackedURL`:
- Remote: `s3://bucket/prefix` (required for serverless timeout and mac pool)
- Local: `file:///path/to/dir` (dev/testing only; incompatible with timeout)

After `Destroy`, `aws.CleanupState()` removes the S3 state files unless `KeepState` is set.

## Dependencies

- **Pulumi Automation API** (`github.com/pulumi/pulumi/sdk/v3/go/auto`) — all infra is managed via inline stacks
- **AWS SDK v2** — read-only queries (spot prices, AMI lookup, AZ enumeration)
- **Azure SDK for Go** — read-only queries (VM sizes, image refs, locations)
- **Cobra + Viper** — CLI parsing
- **go-playground/validator** — struct validation before stack creation
- **logrus** — structured logging
- **freecache** — in-process caching for expensive cloud API calls
