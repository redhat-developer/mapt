# API: Virtual Machine (Azure)

> Concept: [specs/api/concepts/compute.md](../concepts/compute.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/azure/modules/virtual-machine`

Creates an Azure VM. The Azure equivalent of `specs/api/aws/compute.md`.
Always the last Pulumi resource created in an Azure `deploy()` function.

---

## Types

### `VirtualMachineArgs`

```go
type VirtualMachineArgs struct {
    Prefix          string
    ComponentID     string
    ResourceGroup   *resources.ResourceGroup
    NetworkInteface *network.NetworkInterface  // note: typo in source — "Inteface" not "Interface"
    VMSize          string                     // pick one from AllocationResult.ComputeSizes

    SpotPrice *float64           // nil = on-demand; non-nil = spot (sets Priority="Spot")

    Image *data.ImageReference   // from AllocationResult.ImageRef

    // Linux: provide PrivateKey (password auth disabled)
    PrivateKey  *tls.PrivateKey
    // Windows: provide AdminPasswd (password auth)
    AdminPasswd *random.RandomPassword

    AdminUsername    string
    UserDataAsBase64 pulumi.StringPtrInput  // cloud-init or custom script (base64)
    Location         string                 // from AllocationResult.Location
}
```

### `VirtualMachine`

```go
type VirtualMachine = *compute.VirtualMachine
```

The returned value is the raw Pulumi Azure VM resource.
Access the public IP via `Network.PublicIP.IpAddress` (not from the VM itself).

---

## Functions

### `Create`

```go
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *VirtualMachineArgs) (VirtualMachine, error)
```

- **Linux VMs**: sets `LinuxConfiguration` with SSH public key; disables password authentication
- **Windows VMs**: sets `AdminPassword`; no SSH configuration
- **Spot**: sets `Priority = "Spot"` and `BillingProfile.MaxPrice = *SpotPrice`
- **On-demand**: no priority or billing profile set
- Disk: 200 GiB Standard_LRS, created from image
- Boot diagnostics disabled (improves provisioning time)
- Image resolution: handles Marketplace, Community Gallery, and Shared Gallery variants automatically

---

## Image Resolution (internal)

`convertImageRef()` resolves the `ImageReference` to a Pulumi `ImageReferenceArgs`:

| ImageReference field set | Azure resource used |
|---|---|
| `CommunityImageID` | Community Gallery (`communityGalleryImageId`) |
| `SharedImageID` (own subscription) | Direct resource ID |
| `SharedImageID` (other subscription) | Shared Gallery (`sharedGalleryImageId`) |
| `Publisher` + `Offer` + `Sku` | Marketplace image; SKU upgraded to Gen2 if available |

Self-owned detection uses `AZURE_SUBSCRIPTION_ID` env var against the image resource path.

---

## Usage Pattern

```go
vm, err := virtualmachine.Create(ctx, mCtx, &virtualmachine.VirtualMachineArgs{
    Prefix:           *r.prefix,
    ComponentID:      azureTargetID,
    ResourceGroup:    rg,
    NetworkInteface:  n.NetworkInterface,
    VMSize:           r.allocationData.ComputeSizes[0],
    SpotPrice:        r.allocationData.Price,       // nil if on-demand
    Image:            r.allocationData.ImageRef,
    AdminUsername:    amiUserDefault,
    PrivateKey:       privateKey,                   // Linux
    // AdminPasswd:   password,                     // Windows instead
    UserDataAsBase64: udB64,
    Location:         *r.allocationData.Location,
})

// Export host from the network public IP (not from the VM):
ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), n.PublicIP.IpAddress)
```

---

## When to Extend This API

Open a spec under `specs/features/azure/` and update this file when:
- Making disk size configurable
- Adding data disk support
- Adding support for VM extensions (currently Windows uses custom script extension directly in some actions)
- Adding `RunCommand` / `Readiness` methods equivalent to `specs/api/aws/compute.md`
