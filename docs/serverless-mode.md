# Overview

The serverless mode allows to run mapt as a serverless app on the provider the target will be provisioned (i.e. if mapt will create a rhel machine on AWS, it will be executed on Fargate).

The serverless mode can be beneficial on certain circumstances like parallelizing workloads, or release capacity on the system executing mapt. As an example running mapt on serverless-mode will not block a gh runner for the time it takes to provision the infrastructure.

We can also take advantage of the serverless-mode to housekeeping provisioned infrastructures, as an example on the create operation we can set a timeout which will create the matching destroy operation after the timeout. As so in case of any failure preventing actively run the destroy operation it will be executed anyway after the timeout is reached. 

## Requirements

Each cloud provider requires a specific setup for run serverless containers, it will be a requisite to have those setups in place with the specific values for mapt to facilitate this job mapt is providing a set of scripts to create those resources with the naming and values matching its expectations.

One of the main reasons to handle those resources outside of the action executed by mapt is that typically the resources will handle scheduled tasks as so those resources are long time living resources (way beyond the mapt action time)

## AWS - Fargate

On AWS the serverless container service is named Fargate, and it requires an ECS Cluster. 

TBC

## Azure - Azure Containers Instances

TBC