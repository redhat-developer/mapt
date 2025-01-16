#!/bin/bash

#Usage ./serverless.sh ACCESS_KEY SECRET_KEY 

CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-"podman"}"
AWS_CLI="${CONTAINER_RUNTIME} run --rm -it -e AWS_ACCESS_KEY_ID=${1} -e AWS_SECRET_ACCESS_KEY=${2} -e AWS_DEFAULT_REGION="us-west-1" docker.io/amazon/aws-cli:latest"

aws_cmd () {
    ${AWS_CLI} ${1}
}

aws_cmd "ecs create-cluster --cluster-name serverless-mapt"
