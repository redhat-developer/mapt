# Overview

TBC

## Create sample CA

CA_SUBJ="/OU=openshift/CN=admin-kubeconfig-signer-custom"
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 356 -out ca.crt -subj "$CA_SUBJ"