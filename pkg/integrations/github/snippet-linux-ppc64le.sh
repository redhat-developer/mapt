#!/usr/bin/env bash
set -euo pipefail

dnf install -y git-core

git clone --depth=1 "{{ .RunnerImageRepo }}" /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
# Snapshot the runner binary as soon as it is built; a later installer
# failure (e.g. Docker GPG key) can trigger cleanup that deletes it.
(while [ ! -f /opt/runner-cache/config.sh ]; do sleep 10; done
 cp -a /opt/runner-cache /opt/runner-backup) &
WATCHER_PID=$!

bash -c '. scripts/vm.sh rhel 9 minimal --skip-snap-lxd' || true
kill $WATCHER_PID 2>/dev/null || true

if [ ! -f /opt/runner-cache/config.sh ] && [ -d /opt/runner-backup ]; then
    mv /opt/runner-backup /opt/runner-cache
fi

if [ ! -f /opt/runner-cache/config.sh ]; then
    echo "Runner binary not found after build — check build logs" >&2
    exit 1
fi

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
