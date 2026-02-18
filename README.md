# ![mapt](./docs/logo/mapt.svg) Multi Architecture Provisioning Tool 

![code check](https://github.com/redhat-developer/mapt/actions/workflows/build-go.yaml/badge.svg)![oci builds](https://github.com/redhat-developer/mapt/actions/workflows/build-oci.yaml/badge.svg)

Mapt is a swiss army knife for provisioning environments, the project is focused on cover 3 main purposes:

* Offer a set of target environments / services with different topologies across multiple cloud providers.
* Implement best practices leading to increase cost savings, speed up times and security concerns.
* Easily integrate targets with different CI/CD systems or with local envs to facilitate developers testing experience.

### Instances

Mapt offers a set of instances categorize by the OS, instances can benefit from spot module which will allocate the machine on a region with a good relationship beetween cost / availability. Also and depending on the type of instances it will use specific best practices to boost the provisioning time (i.e Fast Launch, Root Volume Replacement, ...). 

Instances can be wrapped on specific topologies like airgap, in this case mapt will set the target isolated and will create a bastion to allow access to it. 

Instances can also define a timeout to avoid leftovers in case destoy operation is missing. Using this approach mapt will be execute as an unateneded execution using servless technologies. 

[MacOS](docs/aws/mac.md)-[Windows Server](docs/aws/windows.md)-[Windows Desktop](docs/azure/windows.md)-[RHEL](docs/aws/rhel.md)-[Fedora](docs/azure/fedora.md)-[Ubuntu](docs/azure/ubuntu.md)

### Services

Mapt offers some managed services boosted with some of the features from the instances offerings (i.e spot) and also create some ad hoc services on top the instances offerings to improve reutilization of instances when there is no easy way to do it (i.e. Mac-Pool).

[AKS](docs/azure/aks.md)-[EKS](docs/aws/eks.md)-[Mac-Pool](docs/aws/mac-pool.md) - [OpenShift-SNC](docs/aws/openshift-snc.md) - [Kind](docs/aws/openshift-snc.md)


### Integrations

Currently each target offered by Mapt can be added as:

* [Github Self Hosted Runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/about-self-hosted-runners)
* [Cirrus Persistent Worker](https://cirrus-ci.org/guide/persistent-workers/)
* [GitLab Runner](docs/gitlab-runner.md)

And [Tekton taks](tkn) are offered to dynamically provision the remote target to use within tekton pipelines