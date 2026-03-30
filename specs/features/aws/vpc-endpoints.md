# Feature: Optional VPC Endpoints

## Status
Implemented

## Context

Every public subnet created by mapt unconditionally creates three VPC endpoints inside
`PublicSubnetRequest.Create()` in `pkg/provider/aws/services/vpc/subnet/public.go`:

| Name | Service | Type |
|---|---|---|
| `s3` | `com.amazonaws.{region}.s3` | Gateway |
| `ecr` | `com.amazonaws.{region}.ecr.dkr` | Interface |
| `ssm` | `com.amazonaws.{region}.ssm` | Interface |

Interface endpoints (ECR, SSM) also create a shared security group allowing TCP 443
inbound from the VPC CIDR — this group is also created unconditionally today.

Targets that do not need these endpoints pay for them unnecessarily. Targets that need
other endpoints cannot add them without code changes.

---

## Requirements

- [x] Accept a `Endpoints []string` field on `NetworkArgs` — each entry is a short name
      (`"s3"`, `"ecr"`, `"ssm"`) identifying the endpoint to create
- [x] Empty slice (default) = **no endpoints created** — breaking change from current
      behaviour; callers that need endpoints must opt in explicitly
- [x] Propagate through the full call chain:
      `cmd params` → action `*Args` → `NetworkArgs` → `NetworkRequest` → `PublicSubnetRequest` → `endpoints()`
- [x] `endpoints()` creates only the endpoints present in the list; unknown names return an
      error before any AWS resource is created
- [x] The Interface-endpoint security group is only created when at least one Interface
      endpoint (`ecr`, `ssm`) is in the list
- [x] Targets that currently depend on specific endpoints (verify EKS, SNC) must pass the
      required endpoint names explicitly in their action args

---

## Out of Scope

- Adding new endpoint types beyond the existing three
- Azure (no equivalent mechanism)
- Airgap path — endpoints are only created for public subnets (`standard/`)

---

## Must Reuse

- `network.Create()` — `specs/api/aws/network.md` — extend `NetworkArgs` with `Endpoints []string`
- `standard.NetworkRequest.CreateNetwork()` — pass `Endpoints` down to `PublicSubnetRequest`
- `PublicSubnetRequest.Create()` — pass `Endpoints` down to `endpoints()`

---

## Must Create

No new files. All changes are within existing files:

### 1. Shared CLI params — `cmd/mapt/cmd/params/params.go`

Follow the three-part pattern described in `specs/cmd/params.md`. Add the Network group:

```go
const (
    Endpoints     = "endpoints"
    EndpointsDesc = "Comma-separated list of VPC endpoints to create. " +
                    "Accepted values: s3, ecr, ssm. Empty = no endpoints."
)

func AddNetworkFlags(fs *pflag.FlagSet) {
    fs.StringSliceP(Endpoints, "", []string{}, EndpointsDesc)
}

func NetworkEndpoints() []string {
    return viper.GetStringSlice(Endpoints)
}
```

`StringSliceP` + `viper.GetStringSlice` handle comma-separated input automatically —
the same mechanism used by `--compute-sizes` and `--spot-excluded-regions`.

### 2. Action args structs — one per target that uses network

Add `Endpoints []string` to each action's public args struct and wire it into
`NetworkArgs` inside `deploy()`:

| Action args struct | File |
|---|---|
| `rhel.RHELArgs` | `pkg/provider/aws/action/rhel/rhel.go` |
| `windows.WindowsArgs` | `pkg/provider/aws/action/windows/windows.go` |
| `fedora.FedoraArgs` | `pkg/provider/aws/action/fedora/fedora.go` |
| `kind.KindArgs` | `pkg/provider/aws/action/kind/kind.go` |
| `snc.SNCArgs` | `pkg/provider/aws/action/snc/snc.go` |
| `eks.EKSArgs` | `pkg/provider/aws/action/eks/eks.go` |

In each action's `deploy()`, pass the field to `NetworkArgs`:

```go
nw, err := network.Create(ctx, r.mCtx, &network.NetworkArgs{
    ...
    Endpoints: r.endpoints,   // new field
})
```

### 3. cmd create files — one per target

Call `params.AddNetworkFlags(flagSet)` and pass `params.NetworkEndpoints()` to the
action args. Pattern (shown for RHEL, identical for all others):

```go
// in getRHELCreate() flagSet block:
params.AddNetworkFlags(flagSet)

// in RHELArgs construction:
&rhel.RHELArgs{
    ...
    Endpoints: params.NetworkEndpoints(),
}
```

Affected cmd files:

| File |
|---|
| `cmd/mapt/cmd/aws/hosts/rhel.go` |
| `cmd/mapt/cmd/aws/hosts/windows.go` |
| `cmd/mapt/cmd/aws/hosts/fedora.go` |
| `cmd/mapt/cmd/aws/hosts/rhelai.go` |
| `cmd/mapt/cmd/aws/services/kind.go` |
| `cmd/mapt/cmd/aws/services/snc.go` |
| `cmd/mapt/cmd/aws/services/eks.go` |

### 4. Network module — `pkg/provider/aws/modules/network/network.go`

Add `Endpoints []string` to `NetworkArgs`; pass to `NetworkRequest`.

### 5. Standard network — `pkg/provider/aws/modules/network/standard/standard.go`

Add `Endpoints []string` to `NetworkRequest`; pass to `PublicSubnetRequest`.

### 6. Public subnet — `pkg/provider/aws/services/vpc/subnet/public.go`

Add `Endpoints []string` to `PublicSubnetRequest`.

Refactor `endpoints()`:
- Accept the list; iterate and create only matching entries
- Unknown names: return error immediately
- Create the security group only when at least one Interface endpoint (`ecr`, `ssm`) is present
- Return without creating anything when the list is empty

---

## Endpoint Identifiers

| Name | AWS service name | Type | Needs security group |
|---|---|---|---|
| `s3` | `com.amazonaws.{region}.s3` | Gateway | No |
| `ecr` | `com.amazonaws.{region}.ecr.dkr` | Interface | Yes |
| `ssm` | `com.amazonaws.{region}.ssm` | Interface | Yes |

The security group (TCP 443 ingress from VPC CIDR) is shared by all Interface endpoints
in the subnet. Created once if any Interface endpoint is in the list; omitted otherwise.

---

## API Changes

Update `specs/api/aws/network.md`:
- Add `Endpoints []string` to `NetworkArgs` type block
- Document the accepted names and the security group behaviour

---

## Acceptance Criteria

### Unit
<!-- Verifiable with `make build` and `make test` — no cloud credentials needed. -->
- `make build` succeeds
- `endpoints()` called with an unknown name returns an error without creating any AWS resource

### Integration
<!-- Requires real cloud credentials. Run manually or in nightly CI. -->
- [ ] `mapt aws rhel create` with no `--endpoints` provisions a VPC with zero endpoints
- [ ] `mapt aws rhel create --endpoints s3,ssm` creates only S3 (Gateway) and SSM (Interface);
      ECR is absent; security group is present
- [ ] `mapt aws rhel create --endpoints s3` creates only S3; no security group is created
- [ ] `mapt aws rhel create --endpoints foo` returns an error before any stack is touched
- [ ] Targets that depended on endpoints before this change (verify EKS, SNC) pass their
      required endpoint names explicitly and continue to work
