# API: Network (AWS)

> Concept: [specs/api/concepts/network.md](../concepts/network.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/aws/modules/network`

Creates the VPC, subnet, internet gateway, optional load balancer, and optional airgap bastion
for any AWS EC2 target. Always the first Pulumi resource created in a `deploy()` function.

---

## Types

### `NetworkArgs`

```go
type NetworkArgs struct {
    Prefix                  string        // resource name prefix
    ID                      string        // component ID (e.g. "aws-rhel") — used in resource naming
    Region                  string        // from AllocationResult.Region
    AZ                      string        // from AllocationResult.AZ
    CreateLoadBalancer       bool          // true when spot is used (LB fronts the ASG)
    Airgap                  bool          // true for airgap topology
    AirgapPhaseConnectivity Connectivity  // ON (with NAT) or OFF (without NAT)
    // Optional VPC endpoints to create in the public subnet.
    // Empty (default) = no endpoints. Accepted: "s3", "ecr", "ssm".
    // Interface endpoints ("ecr", "ssm") share a security group (TCP 443 from VPC CIDR).
    // See specs/features/aws/vpc-endpoints.md
    Endpoints               []string
}

type Connectivity int
const (
    ON  Connectivity = iota  // NAT gateway present — machine has internet egress
    OFF                      // NAT gateway absent — machine is isolated
)
```

### `NetworkResult`

```go
type NetworkResult struct {
    Vpc                         *ec2.Vpc
    Subnet                      *ec2.Subnet                // target subnet (public or private)
    SubnetRouteTableAssociation *ec2.RouteTableAssociation // only set in airgap
    Eip                         *ec2.Eip                   // always created; used for LB or direct instance
    LoadBalancer                *lb.LoadBalancer            // nil when CreateLoadBalancer=false
    Bastion                     *bastion.BastionResult      // nil when Airgap=false
}
```

---

## Functions

### `Create`

```go
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*NetworkResult, error)
```

**Standard path** (`Airgap=false`):
- VPC (`10.0.0.0/16`) with one public subnet (`10.0.2.0/24`) and internet gateway
- No NAT gateway
- EIP always created
- Load balancer created if `CreateLoadBalancer=true`, attached to EIP

**Airgap path** (`Airgap=true`):
- VPC with public subnet (`10.0.2.0/24`) and private (target) subnet (`10.0.101.0/24`)
- Phase ON: public subnet gets NAT gateway → private subnet has internet egress
- Phase OFF: NAT gateway removed → private subnet is isolated
- Bastion host created in public subnet (see `specs/api/aws/bastion.md`)
- Load balancer when `CreateLoadBalancer=true` is internal-facing (private IP)

---

## CIDRs (fixed, not configurable)

| Range | Value |
|---|---|
| VPC | `10.0.0.0/16` |
| Public subnet | `10.0.2.0/24` |
| Private (airgap target) subnet | `10.0.101.0/24` |

---

## Usage Pattern

```go
nw, err := network.Create(ctx, r.mCtx, &network.NetworkArgs{
    Prefix:                  *r.prefix,
    ID:                      awsTargetID,              // constant from constants.go
    Region:                  *r.allocationData.Region,
    AZ:                      *r.allocationData.AZ,
    CreateLoadBalancer:      r.allocationData.SpotPrice != nil,
    Airgap:                  *r.airgap,
    AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
})

// Pass results to compute:
// nw.Vpc    → ComputeRequest.VPC, securityGroup.SGRequest.VPC
// nw.Subnet → ComputeRequest.Subnet
// nw.Eip    → ComputeRequest.Eip
// nw.LoadBalancer → ComputeRequest.LB
// nw.Bastion → ComputeRequest.Readiness() bastion arg
```

---

## When to Extend This API

Open a spec under `specs/features/aws/` and update this file when:
- Adding support for IPv6
- Making CIDRs configurable
- Adding a new topology (e.g. multi-AZ, private-only without bastion)
