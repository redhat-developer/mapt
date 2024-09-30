# mapt

This is a Multi Architecture Provisionig Tool

It allows to spin multiple target environments across multiple cloud providers implementing multiple optimizations like cross data for spot price and eviction rates, or pre create snapshost to improve boot times, ...among others.

![code check](https://github.com/redhat-developer/mapt/actions/workflows/build-go.yaml/badge.svg)
![oci builds](https://github.com/redhat-developer/mapt/actions/workflows/build-oci.yaml/badge.svg)

## Overview

This project is intended to easily spin environments and plug them in on any CI/CD system through ssh. 

mapt create the target machine and return the information and credentials required to connect within the target marchine (host + username + private key)

Also mapt offers a set of features wich are transversal to each type of target machine as so they can be applied to any of them (airgap, proxyed, vpn,...)


## Supported environments

| Platform       | Archs         | Provider      | Type          | Information                  | Tekton                                       | Features |
| -------------- | ------------- | ------------- | ------------- | ---------------------------- | -------------------------------------------- | -------- |
| Mac            | x86, M1, M2   | AWS           | Baremetal     | [info](docs/aws/mac.md)      | [task](tkn/infra-aws-mac.yaml)               | a        | 
| Windows Server | x86           | AWS           | Baremetal     | [info](docs/aws/windows.md)  | [task](tkn/infra-aws-windows-server.yaml)    | a,s      |
| Windows Desktop| x86           | Azure         | Virtualized   | [info](docs/azure/windows.md)| [task](tkn/infra-azure-windows-desktop.yaml) | s        |
| RHEL           | x86, arm64    | AWS           | Customizable  | [info](docs/aws/rhel.md)     | [task](tkn/infra-aws-rhel.yaml)              | a,s      |
| RHEL           | x86, arm64    | Azure         | Virtualized   | [info](docs/azure/rhel.md)   | [task](tkn/infra-azure-rhel.yaml)            | s        |
| Fedora         | x86, arm64    | AWS           | Customizable  | [info](docs/aws/fedora.md)   | [task](tkn/infra-aws-fedora.yaml)            | a,s      |
| Fedora         | x86, arm64    | Azure         | Customizable  | [info](docs/azure/fedora.md) | [task](tkn/infra-azure-fedora.yaml)          | a,s      |
| Ubuntu         | x86           | Azure         | Virtualized   | [info](docs/azure/ubuntu.md) | -                                            | s        |

Features:

* a airgap
* s spot
* p proxy
* v vpn

## Github Self hosted runner

`mapt` can setup a deployed machine as a Self Hosted runner on most of the Platform and Provider combinations
it supports.

Use the following flags with `mapt <provider> <platform> create` command:

```
--install-ghactions-runner <bool>   Install and setup Github Actions runner in the instance
--ghactions-runner-name <string>    Name for the Github Actions Runner
--ghactions-runner-repo <string>    Full URL of the repository where the Github Actions Runner should be registered
--ghactions-runner-token <string>   Token needed for registering the Github Actions Runner token
```

