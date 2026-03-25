# API: Compute (AWS EC2)

> Concept: [specs/api/concepts/compute.md](../concepts/compute.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute`

Creates the EC2 instance (on-demand) or Auto Scaling Group (spot). Always the last Pulumi
resource created in a `deploy()` function, after networking, keypair, and security groups.

---

## Types

### `ComputeRequest`

```go
type ComputeRequest struct {
    MCtx             *mc.Context
    Prefix           string
    ID               string              // component ID — used in resource naming
    VPC              *ec2.Vpc            // from network.NetworkResult.Vpc
    Subnet           *ec2.Subnet         // from network.NetworkResult.Subnet
    Eip              *ec2.Eip            // from network.NetworkResult.Eip
    LB               *lb.LoadBalancer    // from network.NetworkResult.LoadBalancer; nil = on-demand
    LBTargetGroups   []int               // TCP ports to register as LB target groups (e.g. []int{22, 3389})
    AMI              *ec2.LookupAmiResult
    KeyResources     *keypair.KeyPairResources
    SecurityGroups   pulumi.StringArray
    InstaceTypes     []string            // from AllocationResult.InstanceTypes
    InstanceProfile  *iam.InstanceProfile // optional — required by SNC for SSM access
    DiskSize         *int                // nil uses the module default (200 GiB)
    Airgap           bool
    Spot             bool                // true when AllocationResult.SpotPrice != nil
    SpotPrice        float64             // only read when Spot=true
    UserDataAsBase64 pulumi.StringPtrInput // cloud-init or PowerShell userdata
    DependsOn        []pulumi.Resource   // explicit Pulumi dependencies
}
```

### `Compute`

```go
type Compute struct {
    Instance         *ec2.Instance        // set when Spot=false
    AutoscalingGroup *autoscaling.Group   // set when Spot=true
    Eip              *ec2.Eip
    LB               *lb.LoadBalancer
    Dependencies     []pulumi.Resource    // pass to Readiness() and RunCommand()
}
```

---

## Functions

### `NewCompute`

```go
func (r *ComputeRequest) NewCompute(ctx *pulumi.Context) (*Compute, error)
```

- `Spot=false`: creates `ec2.Instance` with direct EIP association
- `Spot=true`: creates `ec2.LaunchTemplate` + `autoscaling.Group` with mixed instances policy, forced spot, capacity-optimized allocation strategy; registers LB target groups

### `Readiness`

```go
func (c *Compute) Readiness(
    ctx *pulumi.Context,
    cmd string,           // command.CommandCloudInitWait or command.CommandPing
    prefix, id string,
    mk *tls.PrivateKey,
    username string,
    b *bastion.BastionResult,  // nil when not airgap
    dependencies []pulumi.Resource,
) error
```

Runs `cmd` over SSH on the instance. Blocks Pulumi until it succeeds (timeout: 40 minutes).
Pass `c.Dependencies` as `dependencies`.

### `RunCommand`

```go
func (c *Compute) RunCommand(
    ctx *pulumi.Context,
    cmd string,
    loggingCmdStd bool,    // compute.LoggingCmdStd or compute.NoLoggingCmdStd
    prefix, id string,
    mk *tls.PrivateKey,
    username string,
    b *bastion.BastionResult,
    dependencies []pulumi.Resource,
) (*remote.Command, error)
```

Like `Readiness` but returns the command resource for use as a dependency in subsequent steps.
Used by SNC to chain SSH → cluster ready → CA rotated → fetch kubeconfig.

### `GetHostDnsName`

```go
func (c *Compute) GetHostDnsName(public bool) pulumi.StringInput
```

Returns `LB.DnsName` when LB is set, otherwise `Eip.PublicDns` (public=true) or `Eip.PrivateDns` (public=false).
Export this as `<prefix>-host`.

### `GetHostIP`

```go
func (c *Compute) GetHostIP(public bool) pulumi.StringOutput
```

Returns `Eip.PublicIp` or `Eip.PrivateIp`. Used by SNC (needs IP not DNS for kubeconfig replacement).

---

## Readiness Commands

| Constant | Value | When to use |
|---|---|---|
| `command.CommandCloudInitWait` | `sudo cloud-init status --long --wait \|\| [[ $? -eq 2 \|\| $? -eq 0 ]]` | Linux targets with cloud-init |
| `command.CommandPing` | `echo ping` | Windows targets (no cloud-init) |

---

## Usage Pattern

```go
cr := compute.ComputeRequest{
    MCtx:             r.mCtx,
    Prefix:           *r.prefix,
    ID:               awsTargetID,
    VPC:              nw.Vpc,
    Subnet:           nw.Subnet,
    Eip:              nw.Eip,
    LB:               nw.LoadBalancer,
    LBTargetGroups:   []int{22},        // add 3389 for Windows
    AMI:              ami,
    KeyResources:     keyResources,
    SecurityGroups:   securityGroups,
    InstaceTypes:     r.allocationData.InstanceTypes,
    DiskSize:         &diskSize,        // constant in constants.go
    Airgap:           *r.airgap,
    UserDataAsBase64: udB64,
}
if r.allocationData.SpotPrice != nil {
    cr.Spot = true
    cr.SpotPrice = *r.allocationData.SpotPrice
}
c, err := cr.NewCompute(ctx)

ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), c.GetHostDnsName(!*r.airgap))

return c.Readiness(ctx, command.CommandCloudInitWait,
    *r.prefix, awsTargetID,
    keyResources.PrivateKey, amiUserDefault,
    nw.Bastion, c.Dependencies)
```

---

## When to Extend This API

Open a spec under `specs/features/aws/` and update this file when:
- Adding support for additional storage volumes
- Adding support for instance store (NVMe) configuration
- Exposing health check grace period as configurable (currently hardcoded at 1200s)
- Adding on-demand with spot fallback (noted as TODO in source)
