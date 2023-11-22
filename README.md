# qenvs

automation for qe environments using pulumi

![code check](https://github.com/adrianriobo/qenvs/actions/workflows/build-go.yaml/badge.svg)
![oci builds](https://github.com/adrianriobo/qenvs/actions/workflows/build-oci.yaml/badge.svg)

## Overview

This project is intended to easily spin environments and plug them in on any CI/CD system through ssh. 

Qenvs create the target machine and return the information and credentials required to connect within the target marchine (host + username + private key)

Also Qenvs offers a set of features wich are transversal to each type of target machine as so they can be applied to any of them (airgap, proxyed, vpn,...)


## Supported environments

| Platform       | Archs         | Provider      | Type          | Information                | Tekton                                       | Features |
| -------------- | ------------- | ------------- | ------------- | -------------------------- | -------------------------------------------- | -------- |
| Mac            | x86, M1, M2   | AWS           | Baremetal     | [info](docs/aws/mac.md)    | [task](tkn/infra-aws-mac.yaml)               | a        | 
| Windows Server | x86           | AWS           | Baremetal     | [info](docs/aws/windows.md)| [task](tkn/infra-aws-windows-server.yaml)    | a,s      |
| Windows Desktop| x86           | Azure         | Virtualized   | [info](docs/azure.md)      | [task](tkn/infra-azure-windows-desktop.yaml) | s        |
| RHEL           | x86           | AWS           | Baremetal     | [info](docs/aws/rhel.md)   | [task](tkn/infra-aws-rhel.yaml)              | a,s      |
| Fedora         | x86           | AWS           | Baremetal     | [info](docs/aws/fedora.md) | [task](tkn/infra-aws-fedora.yaml)            | a,s      |

Features:

* a airgap
* s spot
* p proxy
* v vpn