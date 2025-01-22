# mapt (Multi Architecture Provisionig Tool)

Mapt is a swiss army knife for provisioning environments, it could be used across multiple CI/CD systems:

* Github Actions: It is possible to spin the target machines as self hosted runners on your github repo to make use of them within actions.
* Tekton: Each target environemnt offered has its own tekton task spec which could be used as a extenal spec on tekton (with git resolver or even as a bunde)
* Run from anywhere: mapt funcionallity is offered as an OCI image, as so it allows to create environment from almost everywhere as long as you have a container runtime.

Also it includes out of the box some optimizations around provisioning:

* Spot price option which allows to find the best option for the target machine on any location across the target provider.
* Implement optimization around boot time to reduce the amount of time required to spin the machines (i.e. pre created snapshots or change root volumes)

About the target environments offered it is not limited to a single machine or service but it takes care of the full infra allowing to requrest complex topologies:

* Airgap 
* Proxy (Coming...)
* VPN emulation (Coming... )
* Domain Controller integration (Coming... )

![code check](https://github.com/redhat-developer/mapt/actions/workflows/build-go.yaml/badge.svg)
![oci builds](https://github.com/redhat-developer/mapt/actions/workflows/build-oci.yaml/badge.svg)

## Supported environments

### Virtual Machines

| Platform       | Archs         | Provider      | Type          | Information                  | Tekton                                       
| -------------- | ------------- | ------------- | ------------- | ---------------------------- | -------------------------------------------- 
| Mac            | x86, M1, M2   | AWS           | Baremetal     | [info](docs/aws/mac.md)      | [task](tkn/infra-aws-mac.yaml)               
| Windows Server | x86           | AWS           | Baremetal     | [info](docs/aws/windows.md)  | [task](tkn/infra-aws-windows-server.yaml)    
| Windows Desktop| x86           | Azure         | Virtualized   | [info](docs/azure/windows.md)| [task](tkn/infra-azure-windows-desktop.yaml) 
| RHEL           | x86, arm64    | AWS           | Customizable  | [info](docs/aws/rhel.md)     | [task](tkn/infra-aws-rhel.yaml)              
| RHEL           | x86, arm64    | Azure         | Virtualized   | [info](docs/azure/rhel.md)   | [task](tkn/infra-azure-rhel.yaml)            
| Fedora         | x86, arm64    | AWS           | Customizable  | [info](docs/aws/fedora.md)   | [task](tkn/infra-aws-fedora.yaml)            
| Fedora         | x86, arm64    | Azure         | Customizable  | [info](docs/azure/fedora.md) | [task](tkn/infra-azure-fedora.yaml)          
| Ubuntu         | x86           | Azure         | Virtualized   | [info](docs/azure/ubuntu.md) | -                                            

### Services

| Service        | Provider      | Information                  | Tekton                         
| -------------- | ------------- | -------------                | ---------------------------- | 
| AKS            | Azure         | [info](docs/azure/aks.md)    | [task](tkn/infra-azure-aks.yaml) 
| Mac-pool       | AWS           | [info](docs/aws/mac-pool.md) | - 

## CI/CD integrations

### Github Self hosted runner

`mapt` can setup a deployed machine as a Self Hosted runner on most of the Platform and Provider combinations
it supports.

Use the following flags with `mapt <provider> <platform> create` command:

```
--install-ghactions-runner <bool>   Install and setup Github Actions runner in the instance
--ghactions-runner-name <string>    Name for the Github Actions Runner
--ghactions-runner-repo <string>    Full URL of the repository where the Github Actions Runner should be registered
--ghactions-runner-token <string>   Token needed for registering the Github Actions Runner token
```

