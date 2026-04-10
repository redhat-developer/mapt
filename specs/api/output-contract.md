# API: Output Contract

**Package:** `github.com/redhat-developer/mapt/pkg/provider/util/output`

Defines the files written to `ResultsOutput` after a successful `create`. These files are
the interface between mapt and the CI systems that consume it (Tekton tasks, GitHub workflows,
shell scripts). Changing a filename is a breaking change for all consumers.

---

## Function

### `output.Write`

```go
func Write(stackResult auto.UpResult, destinationFolder string, results map[string]string) error
```

- `results` maps a Pulumi stack output key → destination filename
- Writes each value as a plain text file with permissions `0600`
- Silently skips outputs that are not strings (logs a debug message)
- No-op when `destinationFolder` is empty

---

## Standard Output Files

These filenames are stable across all targets that produce them.
CI consumers depend on these exact names.

| Filename | Content | Targets |
|---|---|---|
| `host` | Hostname or IP to SSH/RDP to | All |
| `username` | OS login username | All |
| `id_rsa` | PEM-encoded SSH private key | All Linux targets, Windows (SSH) |
| `userpassword` | Administrator password (plaintext) | Windows targets |
| `kubeconfig` | kubectl-compatible kubeconfig YAML | SNC, EKS, Kind |
| `kubeadmin-password` | OCP kubeadmin password | SNC only |
| `developer-password` | OCP developer password | SNC only |

### Airgap Additional Files (written by `bastion.WriteOutputs`)

| Filename | Content |
|---|---|
| `bastion_host` | Bastion public IP |
| `bastion_username` | Bastion SSH username (`ec2-user`) |
| `bastion_id_rsa` | Bastion SSH private key |

---

## Pulumi Stack Export Keys

Stack output keys follow the pattern `<prefix>-<name>`. The `prefix` defaults to `"main"`
when not explicitly set by the caller.

| Stack output key | → | Filename |
|---|---|---|
| `<prefix>-host` | | `host` |
| `<prefix>-username` | | `username` |
| `<prefix>-id_rsa` | | `id_rsa` |
| `<prefix>-userpassword` | | `userpassword` |
| `<prefix>-kubeconfig` | | `kubeconfig` |
| `<prefix>-kubeadmin-password` | | `kubeadmin-password` |
| `<prefix>-developer-password` | | `developer-password` |
| `<prefix>-bastion_id_rsa` | | `bastion_id_rsa` |
| `<prefix>-bastion_username` | | `bastion_username` |
| `<prefix>-bastion_host` | | `bastion_host` |

---

## Usage Pattern in `manageResults()`

```go
func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix *string, airgap *bool) error {
    results := map[string]string{
        fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
        fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
        fmt.Sprintf("%s-%s", *prefix, outputHost):           "host",
    }
    if *airgap {
        if err := bastion.WriteOutputs(stackResult, *prefix, mCtx.GetResultsOutputPath()); err != nil {
            return err
        }
    }
    return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}
```

Output key constants (`outputHost`, `outputUsername`, etc.) are defined in the action's
`constants.go` and must match the `ctx.Export(...)` calls in `deploy()`.

---

## When to Change This Contract

Any change to filenames is **breaking** — update this spec and notify consumers:
- Tekton task definitions that read the files (`tkn/template/`)
- GitHub workflow files that reference the output directory
- Any external documentation or user guides

New output files can be added without breaking existing consumers.
