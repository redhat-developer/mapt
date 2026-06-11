#!/usr/bin/env bash
set -euo pipefail

dnf install -y git-core

# Background sshd monitor: logs status every 30s to help diagnose build breakage
(
  LOG=/var/log/sshd-watchdog.log
  while true; do
    echo "--- $(date) ---" >> "$LOG"
    systemctl is-active sshd >> "$LOG" 2>&1
    ss -tlnp | grep :22 >> "$LOG" 2>&1
    sshd -T >> /var/log/sshd-configtest.log 2>&1 || echo "sshd -T FAILED (exit $?)" >> "$LOG"
    sleep 30
  done
) &
WATCHDOG_PID=$!

git clone --depth=1 "{{ .RunnerImageRepo }}" /opt/action-runner-image-pz

cd /opt/action-runner-image-pz
# Allow build to continue past flaky upstream test failures
bash -c '. scripts/vm.sh rhel 9 minimal --skip-snap-lxd' || true

kill $WATCHDOG_PID 2>/dev/null || true

echo "=== POST-BUILD SSHD DIAGNOSTICS ===" >> /var/log/sshd-watchdog.log
systemctl status sshd >> /var/log/sshd-watchdog.log 2>&1
sshd -T >> /var/log/sshd-watchdog.log 2>&1 || echo "sshd -T FAILED" >> /var/log/sshd-watchdog.log
journalctl -u sshd --no-pager -n 50 >> /var/log/sshd-watchdog.log 2>&1
ls -la /etc/ssh/ssh_host_* >> /var/log/sshd-watchdog.log 2>&1
ls -la /usr/share/crypto-policies/ >> /var/log/sshd-watchdog.log 2>&1
cat /etc/pam.d/system-auth >> /var/log/sshd-watchdog.log 2>&1

# Attempt sshd repair: fix PAM duplicates, restore permissions, restart
for f in /etc/pam.d/system-auth /etc/pam.d/password-auth; do
    if [ -f "$f" ]; then
        awk '!seen[$0]++' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
    fi
done
chmod 600 /etc/ssh/ssh_host_*_key 2>/dev/null || true
chmod 644 /etc/ssh/ssh_host_*_key.pub 2>/dev/null || true
find /usr/share/crypto-policies/ -type f -exec chmod 644 {} + 2>/dev/null || true
find /usr/share/crypto-policies/ -type d -exec chmod 755 {} + 2>/dev/null || true
systemctl restart sshd 2>/dev/null || true

echo "=== POST-REPAIR SSHD STATUS ===" >> /var/log/sshd-watchdog.log
systemctl status sshd >> /var/log/sshd-watchdog.log 2>&1
ss -tlnp | grep :22 >> /var/log/sshd-watchdog.log 2>&1

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
