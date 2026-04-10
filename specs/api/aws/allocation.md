# API: Allocation (AWS)

> Concept: [specs/api/concepts/allocation.md](../concepts/allocation.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation`

Single entry point for resolving where and on what instance type a target will run.
All AWS EC2 action `Create()` functions call this before any Pulumi stack is touched.

---

## Types

### `AllocationArgs`

> `ComputeRequestArgs` and `SpotArgs` are cross-provider types — see `specs/api/provider-interfaces.md`.

```go
type AllocationArgs struct {
    ComputeRequest        *cr.ComputeRequestArgs  // required: hardware constraints
    Prefix                *string                 // required: used to name the spot stack
    AMIProductDescription *string                 // optional: e.g. "Linux/UNIX" — used for spot price queries
    AMIName               *string                 // optional: scopes spot search to AMI availability
    Spot                  *spotTypes.SpotArgs     // nil = on-demand; non-nil = spot evaluation
}
```

### `AllocationResult`

```go
type AllocationResult struct {
    Region        *string   // AWS region to deploy into
    AZ            *string   // availability zone within that region
    SpotPrice     *float64  // nil when on-demand; set when spot was selected
    InstanceTypes []string  // one or more compatible instance type strings
}
```

---

## Functions

### `Allocation`

```go
func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error)
```

**Spot path** (`args.Spot != nil && args.Spot.Spot == true`):
- Creates or reuses a `spotOption-<projectName>` Pulumi stack
- Queries spot prices across eligible regions; selects best region/AZ/price
- Idempotent: if the stack already exists, returns its saved outputs without re-querying
- Returns `AllocationResult` with all four fields set

**On-demand path** (`args.Spot == nil` or `args.Spot.Spot == false`):
- Uses `mCtx.TargetHostingPlace()` as the region (set from provider default)
- Iterates AZs until one supports the required instance types
- Returns `AllocationResult` with `SpotPrice == nil`

**Error:** returns `ErrNoSupportedInstanceTypes` if no AZ in the region supports the requested types.

---

## Usage Pattern

```go
// In every AWS action Create():
r.allocationData, err = allocation.Allocation(mCtx, &allocation.AllocationArgs{
    Prefix:                &args.Prefix,
    ComputeRequest:        args.ComputeRequest,
    AMIProductDescription: &amiProduct,   // constant in the action's constants.go
    Spot:                  args.Spot,
})

// Then pass results into the deploy function:
// r.allocationData.Region  → NetworkArgs.Region, ComputeRequest credential region
// r.allocationData.AZ      → NetworkArgs.AZ
// r.allocationData.InstanceTypes → ComputeRequest.InstaceTypes
// r.allocationData.SpotPrice    → ComputeRequest.SpotPrice (when non-nil)
```

---

## Known Gaps

- `spot.Destroy()` uses `aws.DefaultCredentials` (not region-scoped); verify this is correct
  when the selected spot region differs from the default AWS region
- No re-evaluation of spot selection when the persisted region becomes significantly more expensive
  between runs (by design — idempotency wins; worth documenting in user docs)

---

## When to Extend This API

Open a spec under `specs/features/aws/` and update this file when:
- Adding a new allocation strategy (e.g. reserved instances, on-demand with fallback to spot)
- Adding a new field to `AllocationArgs` that all targets would benefit from
- Changing the idempotency behaviour of the spot stack
