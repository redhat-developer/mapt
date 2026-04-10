# Concept: Compute

The compute module is always the **last Pulumi resource created** in a `deploy()` function.
It creates the VM or instance, wires it to the network and security group, and runs a
readiness check before the stack is considered complete.

---

## Provider-Agnostic Contract

1. Accept network outputs (subnet, public IP), credentials (keypair or password), security groups,
   instance types/sizes, and userdata from the action.
2. On the **spot path**: use a spot-aware resource (AWS ASG / Azure VM priority).
3. On the **on-demand path**: use a standard instance/VM with direct IP assignment.
4. Run a **readiness check** — a remote command that blocks Pulumi until the machine is ready.
5. Export the host address as `<prefix>-host`.

---

## Spot Mechanism

| | AWS (`specs/api/aws/compute.md`) | Azure (`specs/api/azure/virtual-machine.md`) |
|---|---|---|
| Spot resource | `ec2.LaunchTemplate` + `autoscaling.Group` (ASG) | Single VM with `Priority="Spot"` + `MaxPrice` |
| Load balancer | Required — ASG registers target groups | Not applicable |
| Selection source | `AllocationResult.SpotPrice != nil` → `Spot=true` | `AllocationResult.Price != nil` → non-nil `SpotPrice` |

---

## Readiness Check

| | AWS | Azure |
|---|---|---|
| Method | `Compute.Readiness()` — built into the module | Remote command run directly in the action |
| Linux command | `sudo cloud-init status --long --wait` | Same command, called differently |
| Windows command | `echo ping` | Equivalent inline |
| Timeout | 40 minutes | Varies by action |

---

## Host Address

| | AWS | Azure |
|---|---|---|
| DNS/IP source | `Compute.GetHostDnsName()` — returns LB DNS or EIP public DNS | `Network.PublicIP.IpAddress` — from the network module, not the VM |
| Export key | `<prefix>-host` | `<prefix>-host` |

---

## Provider Comparison

| | AWS (`specs/api/aws/compute.md`) | Azure (`specs/api/azure/virtual-machine.md`) |
|---|---|---|
| Disk size | Configurable via `DiskSize *int` | Fixed at 200 GiB |
| LB support | Yes (for spot ASG) | No |
| Airgap | Yes — bastion passed to `Readiness()` | No |
| Readiness helper | `Compute.Readiness()` + `RunCommand()` | No equivalent yet |

---

## Implementation References

- AWS: `specs/api/aws/compute.md`
- Azure: `specs/api/azure/virtual-machine.md`
