# API: Network (Azure)

> Concept: [specs/api/concepts/network.md](../concepts/network.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/azure/modules/network`

Creates the VNet, subnet, public IP, and network interface for any Azure VM target.
Called after the resource group and security group are created in a `deploy()` function.

---

## Types

### `NetworkArgs`

```go
type NetworkArgs struct {
    Prefix        string
    ComponentID   string
    ResourceGroup *resources.ResourceGroup   // must be created before calling network.Create()
    Location      *string                    // from AllocationResult.Location
    SecurityGroup securityGroup.SecurityGroup // must be created before calling network.Create()
}
```

Note: unlike AWS, the security group is passed **in** to `network.Create()` rather than
being created after. Creation order in `deploy()` is therefore:
**resource group → security group → network → VM**

### `Network`

```go
type Network struct {
    Network          *network.VirtualNetwork
    PublicSubnet     *network.Subnet
    NetworkInterface *network.NetworkInterface  // pass to VirtualMachineArgs.NetworkInterface
    PublicIP         *network.PublicIPAddress    // export as <prefix>-host
}
```

---

## Functions

### `Create`

```go
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*Network, error)
```

Creates in sequence:
1. VNet (`10.0.0.0/16`) with RunID as name
2. Subnet (`10.0.2.0/24`)
3. Static Standard-SKU public IP
4. NIC attached to subnet + public IP + security group

All resources are tagged via `mCtx.ResourceTags()`.

---

## CIDRs (fixed, not configurable)

| Range | Value |
|---|---|
| VNet | `10.0.0.0/16` |
| Subnet | `10.0.2.0/24` |

---

## Usage Pattern

```go
// 1. Create resource group (outside network module)
rg, err := resources.NewResourceGroup(ctx, ..., &resources.ResourceGroupArgs{
    Location: pulumi.String(*r.allocationData.Location),
})

// 2. Create security group (before network)
sg, err := securityGroup.Create(ctx, mCtx, &securityGroup.SecurityGroupArgs{
    Name:         resourcesUtil.GetResourceName(*r.prefix, azureTargetID, "sg"),
    RG:           rg,
    Location:     r.allocationData.Location,
    IngressRules: []securityGroup.IngressRules{securityGroup.SSH_TCP},
})

// 3. Create network (takes sg as input)
n, err := network.Create(ctx, mCtx, &network.NetworkArgs{
    Prefix:        *r.prefix,
    ComponentID:   azureTargetID,
    ResourceGroup: rg,
    Location:      r.allocationData.Location,
    SecurityGroup: sg,
})

// 4. Pass to VM:
// n.NetworkInterface → VirtualMachineArgs.NetworkInteface
// n.PublicIP.IpAddress → export as <prefix>-host
```

---

## When to Extend This API

Open a spec under `specs/features/azure/` and update this file when:
- Adding airgap support for Azure (bastion + private subnet pattern)
- Adding load balancer support for spot VM scenarios
- Making CIDRs configurable
