# Concept: Security Group

A security group (or network security group) restricts inbound traffic to the VM/instance.
Both providers create one per target with explicit ingress rules and permissive egress (allow all).

---

## Provider-Agnostic Contract

1. Accept a list of ingress rules (port range, protocol, source CIDR).
2. Deny all inbound traffic not matched by a rule.
3. Allow all outbound traffic (permissive egress — not configurable today).
4. Return a resource reference consumed by the network or compute module.

---

## Creation Order

This is the key structural difference between providers:

| Provider | When created | Passed to |
|---|---|---|
| AWS | After `network.Create()` | `compute.ComputeRequest.SecurityGroups` |
| Azure | Before `network.Create()` | `network.NetworkArgs.SecurityGroup` (required input) |

The Azure network module attaches the NSG to the NIC internally, so the VM does not receive
the security group directly.

---

## Pre-defined Rules

Both providers export `SSH_TCP` and `RDP_TCP` rule constants. Usage differs:

| | AWS | Azure |
|---|---|---|
| Type | Value type — **must copy** before setting `CidrBlocks` | Reference — safe to use directly |
| Source SG | Supported via `IngressRules.SG` | Not supported (CIDR only) |
| Protocol wildcard | `"-1"` (all traffic) | `"*"` |
| Priority | Not applicable | Auto-assigned from 1001 upward |

---

## Provider Comparison

| | AWS (`specs/api/aws/security-group.md`) | Azure (`specs/api/azure/security-group.md`) |
|---|---|---|
| Return type | `*SGResources{SG *ec2.SecurityGroup}` | `SecurityGroup` (alias for `*network.NetworkSecurityGroup`) |
| Source SG in rules | Yes | No |
| VPC/RG binding | Bound to VPC (`SGRequest.VPC`) | Bound to resource group (`SecurityGroupArgs.RG`) |

---

## Implementation References

- AWS: `specs/api/aws/security-group.md`
- Azure: `specs/api/azure/security-group.md`
