# API: Bastion

**Package:** `github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion`

Creates a bastion host in the public subnet of an airgap network. Called automatically by
`network.Create()` when `Airgap=true` — action code never calls bastion directly during deploy.

Action code calls `bastion.WriteOutputs()` in `manageResults()` when airgap is enabled.

---

## Types

### `BastionArgs`

```go
type BastionArgs struct {
    Prefix string
    VPC    *ec2.Vpc
    Subnet *ec2.Subnet  // must be the PUBLIC subnet, not the target subnet
}
```

### `BastionResult`

```go
type BastionResult struct {
    Instance   *ec2.Instance
    PrivateKey *tls.PrivateKey
    Usarname   string   // note: typo in source — "Usarname" not "Username"
    Port       int      // always 22
}
```

---

## Functions

### `Create`

```go
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *BastionArgs) (*BastionResult, error)
```

Called internally by `network.Create()`. Not called directly from action code.

Creates:
- Amazon Linux 2 `t2.small` instance in the public subnet
- Keypair for SSH access
- Security group allowing SSH ingress from `0.0.0.0/0`

Exports to Pulumi stack:
- `<prefix>-bastion_id_rsa`
- `<prefix>-bastion_username`
- `<prefix>-bastion_host`

### `WriteOutputs`

```go
func WriteOutputs(stackResult auto.UpResult, prefix string, destinationFolder string) error
```

Writes the three bastion stack outputs to files in `destinationFolder`:

| Stack output key | Output filename |
|---|---|
| `<prefix>-bastion_id_rsa` | `bastion_id_rsa` |
| `<prefix>-bastion_username` | `bastion_username` |
| `<prefix>-bastion_host` | `bastion_host` |

---

## Usage Pattern

```go
// In deploy(): bastion is returned as part of NetworkResult — no direct call needed
nw, err := network.Create(ctx, mCtx, &network.NetworkArgs{Airgap: true, ...})
// nw.Bastion is populated automatically

// Pass to Readiness() so SSH goes through the bastion:
c.Readiness(ctx, cmd, prefix, id, privateKey, username, nw.Bastion, deps)

// In manageResults(): write bastion files alongside target files
func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix *string, airgap *bool) error {
    if *airgap {
        if err := bastion.WriteOutputs(stackResult, *prefix, mCtx.GetResultsOutputPath()); err != nil {
            return err
        }
    }
    return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}
```

---

## Bastion Instance Spec (fixed, not configurable)

| Property | Value |
|---|---|
| AMI | Amazon Linux 2 (`amzn2-ami-hvm-*-x86_64-ebs`) |
| Instance type | `t2.small` |
| Disk | 100 GiB |
| SSH user | `ec2-user` |
| SSH port | 22 |

---

## When to Extend This API

Open a spec under `specs/features/aws/` and update this file when:
- Making bastion instance type or disk size configurable
- Adding bastion support to Azure targets
- Adding support for Session Manager as an alternative to bastion SSH
