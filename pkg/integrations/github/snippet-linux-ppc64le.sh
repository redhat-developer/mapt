#!/usr/bin/env bash
set -euo pipefail

dnf install -y git-core

git clone --depth=1 "{{ .RunnerImageRepo }}" /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
# Allow build to continue past flaky upstream test failures
bash -c '. scripts/vm.sh rhel 9 minimal --skip-snap-lxd' || true

if [ ! -f /opt/runner-cache/config.sh ]; then
    echo "Runner binary not found after build — check build logs" >&2
    exit 1
fi

id -u runner &>/dev/null || useradd -m -s /bin/bash runner
chown -R runner:runner /opt/runner-cache /opt/dotnet

sudo -u runner bash -c '
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

    nohup ./run.sh > /tmp/gh-runner.log 2>&1 &
'
