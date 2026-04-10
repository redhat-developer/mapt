# Concept: Allocation

Allocation is the pre-stack step that resolves **where** a target will run and **on what hardware**,
before any Pulumi resource is created. Every provider action `Create()` calls its allocation
function first and stores the result on the action struct.

---

## Provider-Agnostic Contract

1. Accept hardware constraints (`ComputeRequestArgs`) and an optional spot preference (`SpotArgs`).
2. On the **spot path**: query cloud pricing across eligible regions/locations; select best price.
3. On the **on-demand path**: use the provider default region/location; filter to available sizes.
4. Return a result struct that downstream modules consume directly — no re-querying.

---

## Cross-Provider Types

These types are defined in the shared provider API and used by both AWS and Azure allocation.

### `ComputeRequestArgs`
**Package:** `github.com/redhat-developer/mapt/pkg/provider/api/compute-request`

```go
type ComputeRequestArgs struct {
    CPUs            int32
    GPUs            int32
    GPUManufacturer string
    GPUModel        string
    MemoryGib       int32
    Arch            Arch          // Amd64 | Arm64
    NestedVirt      bool          // true when a profile requires nested virtualisation
    ComputeSizes    []string      // skip selector — use these exact instance types/sizes
}
```

When `ComputeSizes` is set, the instance selector is skipped entirely.

### `SpotArgs`
**Package:** `github.com/redhat-developer/mapt/pkg/provider/api/spot`

```go
type SpotArgs struct {
    Spot                  bool
    Tolerance             Tolerance          // Lowest | Low | Medium | High | Highest
    IncreaseRate          int                // % above current price for bid (default 30)
    ExcludedHostingPlaces []string           // regions/locations to skip
}
```

---

## Provider Comparison

| | AWS (`specs/api/aws/allocation.md`) | Azure (`specs/api/azure/allocation.md`) |
|---|---|---|
| Location key | Region + AZ (two fields in result) | Location (one field in result) |
| Spot persistence | Separate `spotOption` Pulumi stack — idempotent across runs | No stack — re-evaluated each run |
| Instance selector | `aws/data.NewComputeSelector()` | `azure/data.NewComputeSelector()` |
| Extra input | `AMIName`, `AMIProductDescription` | `OSType`, `ImageRef` |
| Extra output | `AZ *string` | `ImageRef *data.ImageReference` |

---

## Implementation References

- AWS: `specs/api/aws/allocation.md`
- Azure: `specs/api/azure/allocation.md`
- Shared types: `specs/api/provider-interfaces.md`
