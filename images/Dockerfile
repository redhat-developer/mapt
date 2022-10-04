FROM docker.io/library/golang:1.18 as builder

USER root
WORKDIR /workspace
COPY . .

ENV PULUMI_VERSION 3.40.2
ENV PULUMI_URL https://github.com/pulumi/pulumi/releases/download/v${PULUMI_VERSION}/pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz

RUN make build \
    && curl -LO ${PULUMI_URL} \
    && tar -xzvf pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL MAINTAINER "Adrian Riobo" "<ariobolo@redhat.com>"

COPY --from=builder /workspace/out/qenvs /workspace/pulumi/pulumi /workspace/images/entrypoint.sh /usr/local/bin/

RUN microdnf install -y awscli \
    # this is for dev purposes need to improve with backedURL
    && mkdir -p /tmp/qenvs

ENTRYPOINT entrypoint.sh 
