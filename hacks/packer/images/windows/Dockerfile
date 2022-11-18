FROM docker.io/hashicorp/packer:1.8.4

WORKDIR /workspace
# context is packer folder
COPY windows .
COPY images/windows/entrypoint.sh /usr/local/bin

RUN apk add --update curl aws-cli jq

ENTRYPOINT entrypoint.sh


