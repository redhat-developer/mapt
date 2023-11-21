# qenvs

automation for qe environments using pulumi

![code check](https://github.com/adrianriobo/qenvs/actions/workflows/build-go.yaml/badge.svg)
![oci builds](https://github.com/adrianriobo/qenvs/actions/workflows/build-oci.yaml/badge.svg)

## Supported environments

Currently qenvs wil handle offerings on azure and aws the main purpose is offer machines which allows nested virtualization and could be added to a CI/CD system for handling automation on top of them.

| Platform       | Archs         | Provider      | Type          | Information                | Tekton                                       |
| -------------- | ------------- | ------------- | ------------- | -------------------------- | -------------------------------------------- |
| Mac            | x86, M1, M2   | AWS           | Baremetal     | [info](docs/aws/mac.md)    | [task](tkn/infra-aws-mac.yaml)               |
| Windows Server | x86           | AWS           | Baremetal     | [info](docs/aws/windows.md)| [task](tkn/infra-aws-windows-server.yaml)    |
| Windows Desktop| x86           | Azure         | Virtualized   | [info](docs/azure.md)      | [task](tkn/infra-azure-windows-desktop.yaml) |
| RHEL           | x86           | AWS           | Baremetal     | [info](docs/aws/rhel.md)   | [task](tkn/infra-aws-rhel.yaml)              |
| Fedora         | x86           | AWS           | Baremetal     | [info](docs/aws/fedora.md) | [task](tkn/infra-aws-fedora.yaml)            |