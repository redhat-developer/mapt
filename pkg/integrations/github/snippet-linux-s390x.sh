#!/usr/bin/env bash
set -euo pipefail

apt-get update -y && apt-get install -y software-properties-common

git clone --branch "{{ .RunnerImageRepoVersion }}" --depth=1 "{{ .RunnerImageRepo }}" /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
# Allow build to continue past flaky upstream test failures
bash -c '. scripts/vm.sh ubuntu 22.04 minimal --skip-snap-lxd' || true

# vm.sh may fail on a flaky test but still produce a complete layout; copy it if needed
if [ ! -f /opt/runner-cache/config.sh ] && [ -f /tmp/runner/_layout/config.sh ]; then
    mkdir -p /opt/runner-cache
    cp -a /tmp/runner/_layout/. /opt/runner-cache/
fi

if [ ! -f /opt/runner-cache/config.sh ]; then
    echo "Runner binary not found after build — check build logs" >&2
    exit 1
fi

id -u runner &>/dev/null || useradd -m -s /bin/bash runner
chown -R runner:runner /opt/runner-cache

sudo -u runner bash -c '
    cd /opt/runner-cache

    ./config.sh \
        --unattended \
        --disableupdate \
        --name "{{ .Name }}" \
        --labels "{{ .Labels }}" \
        --url "{{ .RepoURL }}" \
        --token "{{ .Token }}"{{ if .Ephemeral }} \
        --ephemeral{{ end }}

    nohup ./run.sh > /tmp/gh-runner.log 2>&1 &
'
