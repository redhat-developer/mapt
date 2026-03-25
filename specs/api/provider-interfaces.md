# API: Provider Interfaces (Cross-Cloud)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/api`

Defines the hardware-constraint and spot-selection types that are **shared across all cloud
providers**. Both AWS and Azure allocations are driven by the same input structs; each provider
supplies its own implementation of the selector interfaces.

This layer sits *below* `specs/api/aws/allocation.md` and `specs/api/azure/allocation.md` —
those allocation modules call these selectors internally. Action code interacts with these
types directly (passing `ComputeRequestArgs` and `SpotArgs` into `AllocationArgs`), but
never calls the selector interfaces itself.

---

## Package: `compute-request`

**Full path:** `github.com/redhat-developer/mapt/pkg/provider/api/compute-request`

### Types

```go
type Arch int

const (
    Amd64 Arch = iota + 1
    Arm64
    MaxResults = 20  // max VM types returned per selector call
)

type ComputeRequestArgs struct {
    CPUs            int32
    GPUs            int32
    GPUManufacturer string
    GPUModel        string
    MemoryGib       int32
    Arch            Arch
    NestedVirt      bool
    // Override: skip selector entirely, use these sizes directly
    ComputeSizes    []string
}

type ComputeSelector interface {
    Select(args *ComputeRequestArgs) ([]string, error)
}
```

`ComputeRequestArgs` is embedded in every `AllocationArgs` on both clouds.
If `ComputeSizes` is pre-populated, the selector is skipped — useful when
a specific VM type is required rather than capacity-matched selection.

### Functions

```go
func Validate(cpus, memory int32, arch Arch) error
func (a Arch) String() string  // "x64" | "Arm64"
```

### Provider Implementations

| Provider | Type | Package |
|---|---|---|
| AWS | `data.ComputeSelector` | `pkg/provider/aws/data` |
| Azure | `data.ComputeSelector` | `pkg/provider/azure/data` |

**AWS** uses the `amazon-ec2-instance-selector` library to filter by vCPUs, memory, and arch
across all available instance types.

**Azure** queries the ARM Resource SKUs API, then filters by vCPUs, memory, arch, HyperV Gen2
support, nested virt eligibility, PremiumIO, and `AcceleratedNetworkingEnabled`. Results are
sorted by vCPU count ascending. Azure also exposes `FilterComputeSizesByLocation()` as a
standalone helper used by the on-demand allocation path.

---

## Package: `spot`

**Full path:** `github.com/redhat-developer/mapt/pkg/provider/api/spot`

### Types

```go
type Tolerance int

const (
    Lowest  Tolerance = iota  // eviction rate 0–5% (AWS: placement score ≥ 7)
    Low                       // eviction rate 5–10%
    Medium                    // eviction rate 10–15%
    High                      // eviction rate 15–20%
    Highest                   // eviction rate 20%+  (AWS: placement score ≥ 1)
)

var DefaultTolerance = Lowest

type SpotArgs struct {
    Spot                  bool
    Tolerance             Tolerance
    IncreaseRate          int       // bid price = base × (1 + IncreaseRate/100); default 30%
    ExcludedHostingPlaces []string  // regions/locations to skip
}

type SpotRequestArgs struct {
    ComputeRequest *cr.ComputeRequestArgs
    OS             *string   // "linux", "windows", "RHEL", "fedora" — affects product filter
    ImageName      *string   // AWS: scopes region search to AMI availability
    SpotParams     *SpotArgs
}

type SpotResults struct {
    ComputeType      []string  // AWS: multiple types for ASG; Azure: single type
    Price            float64   // bid price (already inflated by SafePrice)
    HostingPlace     string    // AWS: region; Azure: location
    AvailabilityZone string    // AWS only; empty on Azure
    ChanceLevel      int       // not yet populated (TODO in source)
}

type SpotSelector interface {
    Select(mCtx *mc.Context, args *SpotRequestArgs) (*SpotResults, error)
}
```

### Functions

```go
func ParseTolerance(str string) (Tolerance, bool)
// "lowest"|"low"|"medium"|"high"|"highest" → Tolerance

func SafePrice(basePrice float64, spotPriceIncreaseRate *int) float64
// Returns basePrice × (1 + rate/100). Default rate = 30%.
// Called by both provider SpotInfo() implementations before returning results.
```

### Provider Implementations

| Provider | Type | Selection strategy |
|---|---|---|
| AWS | `data.SpotSelector` | Placement scores × spot price history across all regions |
| Azure | `data.SpotSelector` | Eviction rates × spot price (via Azure Resource Graph) |

**AWS**: Queries placement scores (API requires an opt-in region as API endpoint) and
spot price history in parallel across all regions. Filters regions where the AMI is
available. Returns up to 8 instance types for the winning AZ (used by the ASG mixed-instances
policy).

**Azure**: Queries eviction rates and spot prices via Azure Resource Graph KQL. Crosses eviction
rate buckets against allowed tolerance, then picks the lowest-price / lowest-eviction-rate
location. Falls back to price-only ranking if eviction-rate data is unavailable. Returns a
single compute size.

---

## Package: `config/userdata`

**Full path:** `github.com/redhat-developer/mapt/pkg/provider/api/config/userdata`

```go
type CloudConfig interface {
    CloudConfig() (*string, error)
}
```

Implemented by cloud-init / cloud-config builder packages used to generate the
`UserData` / `UserDataAsBase64` field on compute resources. Every target that
injects software at boot implements this interface.

---

## Architecture Summary

```
pkg/provider/api/            ← provider-agnostic types & interfaces
  compute-request/
    ComputeRequestArgs        used in AllocationArgs (both clouds)
    ComputeSelector           interface
  spot/
    SpotArgs, SpotResults     used in AllocationArgs (both clouds)
    SpotSelector              interface
    SafePrice()               shared bid-price calculation
  config/userdata/
    CloudConfig               interface for cloud-init builders

pkg/provider/aws/data/       ← AWS implementations
  ComputeSelector             ec2-instance-selector
  SpotSelector                placement scores + price history

pkg/provider/azure/data/     ← Azure implementations
  ComputeSelector             ARM Resource SKUs API
  SpotSelector                Azure Resource Graph (eviction + price)
```

---

## When to Extend This API

Open a spec under `specs/features/aws/` or `specs/features/azure/` and update this file when:
- Adding a third cloud provider (implement both interfaces in the new `data` package)
- Adding GPU-based compute selection (currently fields exist but filtering is partial)
- Making `CPUsRange` / `MemoryRange` filters active (currently commented out)
- Populating `SpotResults.ChanceLevel` (currently a TODO in both implementations)
- Adding `ExcludedRegions` to AWS spot path (field exists in `SpotInfoArgs` but not wired into `SpotRequestArgs`)
