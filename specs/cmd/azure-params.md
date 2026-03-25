# CLI Params: Azure Shared

**Package:** `github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/params`
**File:** `cmd/mapt/cmd/azure/params/params.go`

Azure-provider-specific shared params, used alongside the cross-provider params in
`specs/cmd/params.md`. Every Azure `create` command that accepts a location registers
this flag.

---

## Location

```go
const (
    Location        = "location"
    LocationDesc    = "If spot is passed location will be calculated based on spot results. Otherwise location will be used to create resources."
    LocationDefault = "westeurope"
)
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--location` | string | `westeurope` | Azure region; ignored when `--spot` is set (spot selects the location) |

No `Add*Flags` helper — each cmd registers it directly:

```go
flagSet.StringP(azureParams.Location, "", azureParams.LocationDefault, azureParams.LocationDesc)
```

Mapped to `AllocationArgs.Location` inside the action. When spot is active the allocation
module ignores this value and picks the best-priced region automatically.

See `specs/api/azure/allocation.md`.

---

## When to Extend

Update this file when adding new Azure-wide shared params (e.g. resource group prefix,
subscription override).
