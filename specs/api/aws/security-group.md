# API: Security Group (AWS)

> Concept: [specs/api/concepts/security-group.md](../concepts/security-group.md)

**Package:** `github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group`

Creates an EC2 security group with ingress rules. Called from every AWS action `deploy()`
and from the bastion module internally.

---

## Types

### `SGRequest`

```go
type SGRequest struct {
    Name         string           // resourcesUtil.GetResourceName(prefix, id, "sg")
    Description  string
    IngressRules []IngressRules
    VPC          *ec2.Vpc         // from network.NetworkResult.Vpc
}
```

### `IngressRules`

```go
type IngressRules struct {
    Description string
    FromPort    int
    ToPort      int
    Protocol    string    // "tcp", "udp", "icmp", "-1" (all)
    CidrBlocks  string    // CIDR string; empty = 0.0.0.0/0; mutually exclusive with SG
    SG          *ec2.SecurityGroup // source SG; mutually exclusive with CidrBlocks
}
```

### `SGResources`

```go
type SGResources struct {
    SG *ec2.SecurityGroup
}
```

---

## Functions

### `Create`

```go
func (r SGRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*SGResources, error)
```

Creates the security group with all ingress rules and a permissive egress (all traffic allowed).

---

## Pre-defined Rules

```go
// Defined in security-group/defaults.go — copy and set CidrBlocks before use
var SSH_TCP  = IngressRules{Description: "SSH",  FromPort: 22,   ToPort: 22,   Protocol: "tcp"}
var RDP_TCP  = IngressRules{Description: "RDP",  FromPort: 3389, ToPort: 3389, Protocol: "tcp"}

// Port constants
const SSH_PORT  = 22
const HTTPS_PORT = 443
```

**Important:** `SSH_TCP` and `RDP_TCP` are value types — copy them before setting `CidrBlocks`:
```go
sshRule := securityGroup.SSH_TCP
sshRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4  // "0.0.0.0/0"
```

---

## Usage Pattern

```go
sg, err := securityGroup.SGRequest{
    Name:        resourcesUtil.GetResourceName(*prefix, awsTargetID, "sg"),
    VPC:         nw.Vpc,
    Description: fmt.Sprintf("sg for %s", awsTargetID),
    IngressRules: []securityGroup.IngressRules{sshRule},
}.Create(ctx, mCtx)

// Convert to StringArray for ComputeRequest:
sgs := util.ArrayConvert([]*ec2.SecurityGroup{sg.SG},
    func(sg *ec2.SecurityGroup) pulumi.StringInput { return sg.ID() })
return pulumi.StringArray(sgs[:]), nil
```

---

## When to Extend This API

Open a spec under `specs/features/aws/` and update this file when:
- Adding new pre-defined rule constants (e.g. WinRM, HTTPS)
- Adding IPv6 CIDR support
- Adding support for egress rule customisation (currently always allow-all egress)
