# API: Security Group (Azure)

> Concept: [specs/api/concepts/security-group.md](../concepts/security-group.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group`

Creates an Azure Network Security Group (NSG). The NSG is created **before** the network
module is called, because `network.Create()` takes the NSG as an input argument.
See `specs/api/azure/network.md`.

---

## Types

### `SecurityGroupArgs`

```go
type SecurityGroupArgs struct {
    Name         string                     // resourcesUtil.GetResourceName(prefix, id, "sg")
    RG           *resources.ResourceGroup   // resource group the NSG belongs to
    Location     *string                    // from AllocationResult.Location
    IngressRules []IngressRules
}
```

### `IngressRules`

```go
type IngressRules struct {
    Description string
    FromPort    int
    ToPort      int
    Protocol    string   // "tcp", "udp", "*" (all)
    CidrBlocks  string   // source CIDR; empty = allow any source ("*")
}
```

### `SecurityGroup`

```go
type SecurityGroup = *network.NetworkSecurityGroup
```

A type alias — the raw Pulumi Azure NSG resource. Passed directly into `NetworkArgs.SecurityGroup`.

---

## Functions

### `Create`

```go
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *SecurityGroupArgs) (SecurityGroup, error)
```

Creates the NSG with inbound allow rules. Priorities are auto-assigned starting at 1001,
incrementing by 1 per rule. Egress is unrestricted (Azure default).

---

## Pre-defined Rules

```go
// Defined in security-group/defaults.go — safe to use directly (not value copies like AWS)
var SSH_TCP = IngressRules{Description: "SSH", FromPort: 22,   ToPort: 22,   Protocol: "tcp"}
var RDP_TCP = IngressRules{Description: "RDP", FromPort: 3389, ToPort: 3389, Protocol: "tcp"}

var SSH_PORT int = 22
var RDP_PORT int = 3389
```

Unlike the AWS equivalent, Azure `IngressRules` do not have a source SG field — only CIDR.
Empty `CidrBlocks` allows from any source (`*`), which is the default for SSH and RDP rules.

---

## Usage Pattern

```go
sg, err := securityGroup.Create(ctx, mCtx, &securityGroup.SecurityGroupArgs{
    Name:     resourcesUtil.GetResourceName(*r.prefix, azureTargetID, "sg"),
    RG:       rg,
    Location: r.allocationData.Location,
    IngressRules: []securityGroup.IngressRules{
        securityGroup.SSH_TCP,
        // securityGroup.RDP_TCP,  // add for Windows targets
    },
})

// Pass directly into network:
n, err := network.Create(ctx, mCtx, &network.NetworkArgs{
    SecurityGroup: sg,
    ...
})
```

---

## When to Extend This API

Open a spec under `specs/features/azure/` and update this file when:
- Adding source NSG reference support (intra-VNet rules)
- Adding egress rule customisation
- Adding new pre-defined rule constants
