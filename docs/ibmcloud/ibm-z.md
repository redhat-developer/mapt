# Overview

This actions will handle provision Ubuntu s390x machines on ibm cloud VPC. 
 

## Create

```bash
mapt ibmcloud ibm-z create -h
create

Usage:
  mapt ibmcloud ibm-z create [flags]

Flags:
      --conn-details-output string           path to export host connection information (host, username and privateKey)
      --ghactions-runner-labels strings      List of labels separated by comma to be added to the self-hosted runner
      --ghactions-runner-repo string         Full URL of the repository where the Github Actions Runner should be registered
      --ghactions-runner-token string        Token needed for registering the Github Actions Runner token
  -h, --help                                 help for create
      --it-cirrus-pw-labels stringToString   additional labels to use on the persistent worker (--it-cirrus-pw-labels key1=value1,key2=value2) (default [])
      --it-cirrus-pw-token string            Add mapt target as a cirrus persistent worker. The value will hold a valid token to be used by cirrus cli to join the project.
      --tags stringToString                  tags to add on each resource (--tags name1=value1,name2=value2) (default [])

Global Flags:
      --backed-url string     backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend
      --debug                 Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue.
      --debug-level uint      Set the level of verbosity on debug. You can set from minimum 1 to max 9. (default 3)
      --project-name string   project name to identify the instance of the stack
```

### Outputs

* It will crete an instance and will give as result several files located at path defined by `--conn-details-output`:

  * **host**: host for the Windows machine (lb if spot)
  * **username**: username to connect to the machine
  * **id_rsa**: private key to connect to machine

* Also, it will create a state folder holding the state for the created resources at azure, the path for this folder is defined within `--backed-url`, the content from that folder it is required with the same project name (`--project-name`) in order to destroy the resources.

### Container

```bash
podman run -d --name mapt-rhel \
        -v ${PWD}:/workspace:z \
        -e IBMCLOUD_ACCOUNT=XXX \
        -e IBMCLOUD_API_KEY=XXX \
        -e IC_REGION=us-south \
        -e IC_ZONE=us-south-2 \
        quay.io/redhat-developer/mapt:1.0.0 ibm-z create \
            --project-name ibm-z \
            --backed-url file:///workspace \
            --conn-details-output /workspace
```
