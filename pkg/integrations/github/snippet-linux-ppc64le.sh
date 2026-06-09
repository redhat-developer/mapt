#!/usr/bin/env bash
set -euo pipefail

git clone {{ .RunnerImageRepo }} /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
bash -c '. scripts/vm.sh rhel 9 minimal --skip-snap-lxd'

cd /opt/runner-cache
export DOTNET_ROOT=/opt/dotnet
export PATH=$PATH:$DOTNET_ROOT

./config.sh \
    --unattended \
    --disableupdate \
    --ephemeral \
    --name "{{ .Name }}" \
    --labels "{{ .Labels }}" \
    --url "{{ .RepoURL }}" \
    --token "{{ .Token }}"

nohup ./run.sh > /var/log/gh-runner.log 2>&1 &
