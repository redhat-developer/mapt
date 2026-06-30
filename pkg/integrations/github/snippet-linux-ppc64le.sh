#!/usr/bin/env bash
set -euo pipefail

dnf install -y git-core

git clone --depth=1 "{{ .RunnerImageRepo }}" /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
bash -c '. scripts/vm.sh rhel 9 minimal --skip-snap-lxd' || true

# The upstream configure-system.sh runs chmod -R 777 /usr/share which breaks
# sshd privilege separation (/usr/share/empty.sshd must be root-owned, not
# world-writable). Also fix PAM duplicates from configure-limits.sh.
chmod 755 /usr/share/empty.sshd 2>/dev/null || true
chown root:root /usr/share/empty.sshd 2>/dev/null || true
for f in /etc/pam.d/system-auth /etc/pam.d/password-auth; do
    if [ -f "$f" ]; then
        awk '!seen[$0]++' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
    fi
done
systemctl restart sshd 2>/dev/null || true

if [ ! -f /opt/runner-cache/config.sh ]; then
    echo "Runner binary not found after build" >&2
    exit 1
fi

id -u runner &>/dev/null || useradd -m -s /bin/bash runner
chown -R runner:runner /opt/runner-cache

sudo -u runner bash -c '
    cd /opt/runner-cache

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
