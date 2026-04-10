# Concept: Network

The network module is always the **first Pulumi resource created** in a `deploy()` function.
It establishes the virtual network, subnet, and public IP that all subsequent resources depend on.

---

## Provider-Agnostic Contract

1. Accept a prefix, component ID, and location/region from `AllocationResult`.
2. Create a virtual network + subnet with fixed CIDRs (`10.0.0.0/16` / `10.0.2.0/24`).
3. Produce a public IP (or EIP) and a subnet reference consumed by the compute module.
4. Return a result struct — downstream modules must not re-query network state.

---

## Creation Order in `deploy()`

```
network.Create()           ← first
securityGroup.Create()     ← depends on network (AWS only; Azure reverses this)
keypair / password
compute.NewCompute()       ← last
```

Azure is the exception: the security group is created **before** `network.Create()` because
`NetworkArgs.SecurityGroup` is a required input. See `specs/api/concepts/security-group.md`.

---

## Provider Comparison

| | AWS (`specs/api/aws/network.md`) | Azure (`specs/api/azure/network.md`) |
|---|---|---|
| Airgap support | Yes — two-phase NAT removal, private subnet, bastion | No |
| Load balancer | Optional, created internally when spot is used | Not managed by this module |
| Security group | Created after network; passed to compute | Created before network; passed in as input |
| Public address output | EIP (`NetworkResult.Eip`) or LB DNS | `Network.PublicIP.IpAddress` |
| Bastion | Automatic when `Airgap=true` | Not available |

---

## Implementation References

- AWS: `specs/api/aws/network.md`
- Azure: `specs/api/azure/network.md`
