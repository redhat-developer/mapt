# API: Allocation (Azure)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/azure/modules/allocation`

Single entry point for resolving which Azure location, VM size, and image to use.
All Azure action `Create()` functions call this before any Pulumi stack is touched.

> Concept: [specs/api/concepts/allocation.md](../concepts/allocation.md)

---

## Types

### `AllocationArgs`

> `ComputeRequestArgs` and `SpotArgs` are cross-provider types — see `specs/api/provider-interfaces.md`.

```go
type AllocationArgs struct {
    ComputeRequest *cr.ComputeRequestArgs  // required: hardware constraints
    OSType         string                   // e.g. "Linux", "Windows" — used for spot queries
    ImageRef       *data.ImageReference    // optional: scopes spot search to image availability
    Location       *string                 // required for on-demand; ignored when spot selects location
    Spot           *spotTypes.SpotArgs     // nil = on-demand; non-nil = spot evaluation
}
```

### `AllocationResult`

```go
type AllocationResult struct {
    Location     *string              // Azure region (e.g. "eastus")
    Price        *float64             // nil when on-demand; set when spot was selected
    ComputeSizes []string             // one or more compatible VM size strings
    ImageRef     *data.ImageReference // passed through from args
}
```

---

## Functions

### `Allocation`

```go
func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error)
```

**Spot path** (`args.Spot != nil && args.Spot.Spot == true`):
- Queries spot prices across eligible Azure locations
- Scores by price × availability; selects best location/VM size
- No separate Pulumi stack (unlike AWS) — result is not persisted between runs
- Returns `AllocationResult` with all fields set

**On-demand path** (`args.Spot == nil` or `args.Spot.Spot == false`):
- Uses `args.Location` as the target location
- Filters `ComputeRequest.ComputeSizes` to those available in the location
- Returns `AllocationResult` with `Price == nil`

---

## Related Types

### `ImageReference`
**Package:** `github.com/redhat-developer/mapt/pkg/provider/azure/data`

```go
type ImageReference struct {
    // Marketplace image
    Publisher string
    Offer     string
    Sku       string
    // Azure Community Gallery
    CommunityImageID string
    // Azure Shared Gallery (private or cross-tenant)
    SharedImageID string
}
```

Exactly one of the three variants should be populated. Use `data.GetImageRef()` to build
a reference from OS type, arch, and version:

```go
func GetImageRef(osTarget OSType, arch string, version string) (*ImageReference, error)
```

Supported `OSType` values: `data.Ubuntu`, `data.RHEL`, `data.Fedora`

### `SpotArgs`
**Package:** `github.com/redhat-developer/mapt/pkg/provider/api/spot`

Cross-provider type — see `specs/api/concepts/allocation.md` for field descriptions.

---

## Usage Pattern

```go
// In every Azure action Create():
r.allocationData, err = allocation.Allocation(mCtx, &allocation.AllocationArgs{
    ComputeRequest: args.ComputeRequest,
    OSType:         "Linux",                    // or "Windows"
    ImageRef:       imageRef,                   // from data.GetImageRef()
    Location:       &defaultLocation,           // provider default, ignored if spot
    Spot:           args.Spot,
})

// Then pass results into the deploy function:
// r.allocationData.Location     → NetworkArgs.Location, VM location
// r.allocationData.ComputeSizes → pick one for VirtualMachineArgs.VMSize
// r.allocationData.Price        → VirtualMachineArgs.SpotPrice (when non-nil)
// r.allocationData.ImageRef     → VirtualMachineArgs.Image
```

---

## When to Extend This API

Open a spec under `specs/features/azure/` and update this file when:
- Persisting Azure spot allocation to a Pulumi stack (for idempotency, matching AWS behaviour)
- Adding new `OSType` values to `data.GetImageRef()`
- Adding `ExcludedLocations` filtering to on-demand path
