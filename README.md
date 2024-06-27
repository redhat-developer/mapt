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

| Platform       | Archs         | Provider      | Type          | Information                | Tekton                                       | Features |
| -------------- | ------------- | ------------- | ------------- | -------------------------- | -------------------------------------------- | -------- |
| Mac            | x86, M1, M2   | AWS           | Baremetal     | [info](docs/aws/mac.md)    | [task](tkn/infra-aws-mac.yaml)               | a        | 
| Windows Server | x86           | AWS           | Baremetal     | [info](docs/aws/windows.md)| [task](tkn/infra-aws-windows-server.yaml)    | a,s      |
| Windows Desktop| x86           | Azure         | Virtualized   | [info](docs/azure.md)      | [task](tkn/infra-azure-windows-desktop.yaml) | s        |
| RHEL           | x86, arm64    | AWS           | Baremetal     | [info](docs/aws/rhel.md)   | [task](tkn/infra-aws-rhel.yaml)              | a,s      |
| Fedora         | x86           | AWS           | Baremetal     | [info](docs/aws/fedora.md) | [task](tkn/infra-aws-fedora.yaml)            | a,s      |

Features:

* a airgap
* s spot
* p proxy
* v vpn