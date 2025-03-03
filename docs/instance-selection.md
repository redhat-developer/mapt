# Instance Selection based on hardware specification

## Overview

With `mapt` users can select the type of instance to create on the cloud provider by providing its specification, i.e CPUs count, CPU architecture, Memory size
and support for nested virtualization.

There are flags `--cpus`, `--memory`, `--arch` and `--nested-virt` for most of the `mapt <provider> <profile | os> create` except `mac` and `windows` OS for AWS
and `aks` profile for Azure.

The flags also have sensible default values, if an instance type satisfying the requested hardware specs is not offered by the provider, it'll use the default
values to create a machine.

## Creating a fedora VM on AWS with user provided specs

As an example we can use the `--cpus`, `--arch` and `--memory` flags with the `mapt aws fedora create` command to create an arm machine with 8 cpus and 64GB of RAM:
```
$ mapt aws fedora create --spot \
    --arch arm64 \
    --cpus 8 --memory 64 \
    --project-name aws-mapt-fedora-test \
    --backed-url file:///home/mapt/workspace \
    --conn-details-output /tmp/fedora
```
