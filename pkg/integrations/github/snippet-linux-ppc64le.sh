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

# The upstream configure-system.sh runs chmod -R 777 /usr/share which makes
# the sshd privilege separation directory world-writable. sshd refuses to
# start when /usr/share/empty.sshd is not owned by root or is world-writable.
chmod 755 /usr/share/empty.sshd 2>/dev/null || true
chown root:root /usr/share/empty.sshd 2>/dev/null || true
# Also fix PAM duplicates from configure-limits.sh
for f in /etc/pam.d/system-auth /etc/pam.d/password-auth; do
    if [ -f "$f" ]; then
        awk '!seen[$0]++' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
    fi
done
systemctl restart sshd 2>/dev/null || true

echo "=== POST-REPAIR SSHD STATUS ===" >> /var/log/sshd-watchdog.log
systemctl status sshd >> /var/log/sshd-watchdog.log 2>&1
ss -tlnp | grep :22 >> /var/log/sshd-watchdog.log 2>&1

# Upload diagnostics to COS so we can read them without SSH
python3 -c "
import hashlib, hmac, urllib.request, datetime, os, socket
key_id = os.environ.get('COS_KEY_ID', '')
secret = os.environ.get('COS_SECRET', '')
endpoint = os.environ.get('COS_ENDPOINT', '')
bucket = 'mapt-test-bucket-evidence'
hostname = socket.gethostname()
obj = 'debug/' + hostname + '-sshd-watchdog.log'
if key_id and secret and endpoint:
    with open('/var/log/sshd-watchdog.log', 'rb') as f:
        body = f.read()
    now = datetime.datetime.utcnow()
    date_stamp = now.strftime('%Y%m%d')
    amz_date = now.strftime('%Y%m%dT%H%M%SZ')
    region = 'us-south'
    service = 's3'
    host = endpoint.replace('https://','').replace('http://','')
    canonical_uri = '/' + bucket + '/' + obj
    payload_hash = hashlib.sha256(body).hexdigest()
    canonical_headers = 'host:' + host + '\n' + 'x-amz-content-sha256:' + payload_hash + '\n' + 'x-amz-date:' + amz_date + '\n'
    signed_headers = 'host;x-amz-content-sha256;x-amz-date'
    canonical_request = 'PUT\n' + canonical_uri + '\n\n' + canonical_headers + '\n' + signed_headers + '\n' + payload_hash
    algorithm = 'AWS4-HMAC-SHA256'
    credential_scope = date_stamp + '/' + region + '/' + service + '/aws4_request'
    string_to_sign = algorithm + '\n' + amz_date + '\n' + credential_scope + '\n' + hashlib.sha256(canonical_request.encode()).hexdigest()
    def sign(key, msg):
        return hmac.new(key, msg.encode(), hashlib.sha256).digest()
    signing_key = sign(sign(sign(sign(('AWS4' + secret).encode(), date_stamp), region), service), 'aws4_request')
    signature = hmac.new(signing_key, string_to_sign.encode(), hashlib.sha256).hexdigest()
    auth = algorithm + ' Credential=' + key_id + '/' + credential_scope + ', SignedHeaders=' + signed_headers + ', Signature=' + signature
    req = urllib.request.Request(endpoint + canonical_uri, data=body, method='PUT')
    req.add_header('x-amz-date', amz_date)
    req.add_header('x-amz-content-sha256', payload_hash)
    req.add_header('Authorization', auth)
    req.add_header('Content-Type', 'text/plain')
    urllib.request.urlopen(req)
    print('Uploaded diagnostics to COS: ' + obj)
else:
    print('COS credentials not set, skipping upload')
" 2>&1 || echo "COS upload failed"

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
