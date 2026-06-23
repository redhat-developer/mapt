#!/bin/bash
set -e

# Download GitLab Runner
curl -L -o /tmp/gitlab-runner "{{ .CliURL }}"
chmod +x /tmp/gitlab-runner

# Move to trusted path
sudo mv /tmp/gitlab-runner /usr/bin/gitlab-runner

# Fix SELinux context (no-op on non-SELinux systems)
sudo restorecon -v /usr/bin/gitlab-runner 2>/dev/null || true

# Enable Podman socket so the docker executor can reach it
sudo systemctl enable --now podman.socket

# Detect the host's upstream DNS servers and propagate them into every Podman
# container (including nested build containers created by `podman build`).
# Without this, inner build containers inherit a loopback stub address
# (127.0.0.53 / systemd-resolved) that is unreachable from inside a container,
# causing DNS resolution failures like "Could not resolve host: github.com".
_dns_servers=""
if command -v resolvectl &>/dev/null; then
    _dns_servers=$(resolvectl dns 2>/dev/null \
        | awk '{for(i=2;i<=NF;i++) print $i}' \
        | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' \
        | sort -u | tr '\n' ' ' | xargs)
fi
if [ -z "$_dns_servers" ] && command -v nmcli &>/dev/null; then
    _dns_servers=$(nmcli dev show 2>/dev/null \
        | awk '/IP4\.DNS/ {print $2}' \
        | tr '\n' ' ' | xargs)
fi
# On systemd-resolved systems (Ubuntu), /run/systemd/resolve/resolv.conf holds
# the real upstream DNS servers (not the 127.0.0.53 stub in /etc/resolv.conf).
if [ -z "$_dns_servers" ]; then
    _dns_servers=$(awk '/^nameserver/ && $2 !~ /^127\./ && $2 != "::1" {print $2}' \
        /run/systemd/resolve/resolv.conf 2>/dev/null \
        | tr '\n' ' ' | xargs)
fi
if [ -z "$_dns_servers" ]; then
    _dns_servers=$(awk '/^nameserver/ && $2 !~ /^127\./ && $2 != "::1" {print $2}' /etc/resolv.conf \
        | tr '\n' ' ' | xargs)
fi
# Last-resort fallback: if no local DNS could be detected, use public resolvers.
# The machine must have internet access (it talks to GitLab), so these will work.
if [ -z "$_dns_servers" ]; then
    _dns_servers="8.8.8.8 8.8.4.4"
fi
# Build --docker-dns flags for runner registration so every job container gets
# working DNS servers even when Podman's Docker socket API does not honour
# containers.conf dns_servers (which affects executor-container resolution).
_docker_dns_args=()
for _ip in $_dns_servers; do
    _docker_dns_args+=(--docker-dns "$_ip")
done

if [ -n "$_dns_servers" ]; then
    _toml_list=""
    for _ip in $_dns_servers; do
        [ -n "$_toml_list" ] && _toml_list="${_toml_list}, "
        _toml_list="${_toml_list}\"${_ip}\""
    done
    sudo mkdir -p /etc/containers
    if [ ! -f /etc/containers/containers.conf ]; then
        printf '[containers]\ndns_servers = [%s]\ndns_options = ["timeout:2", "attempts:5", "single-request"]\n' \
            "$_toml_list" | sudo tee /etc/containers/containers.conf > /dev/null
    elif grep -q '^\[containers\]' /etc/containers/containers.conf; then
        # Scope the dns_servers check to the [containers] section only
        if awk '/^\[containers\]/{f=1;next} /^\[/{f=0} f && /^dns_servers/{found=1} END{exit !found}' \
                /etc/containers/containers.conf; then
            # Replace dns_servers only within [containers]
            awk -v "val=dns_servers = [${_toml_list}]" \
                '/^\[containers\]/{s=1} /^\[/ && !/^\[containers\]/{s=0}
                 s && /^dns_servers/{$0=val} 1' \
                /etc/containers/containers.conf \
                | sudo tee /etc/containers/containers.conf.tmp > /dev/null \
                && sudo mv /etc/containers/containers.conf.tmp /etc/containers/containers.conf
        else
            sudo sed -i "/^\[containers\]/a dns_servers = [${_toml_list}]" \
                /etc/containers/containers.conf
        fi
        # Add or update dns_options within [containers]
        if grep -q '^dns_options' /etc/containers/containers.conf; then
            sudo sed -i 's|^dns_options.*|dns_options = ["timeout:2", "attempts:5", "single-request"]|' \
                /etc/containers/containers.conf
        else
            sudo sed -i '/^\[containers\]/a dns_options = ["timeout:2", "attempts:5", "single-request"]' \
                /etc/containers/containers.conf
        fi
    else
        printf '\n[containers]\ndns_servers = [%s]\ndns_options = ["timeout:2", "attempts:5", "single-request"]\n' \
            "$_toml_list" | sudo tee -a /etc/containers/containers.conf > /dev/null
    fi
    # Ensure the file is world-readable so rootless Podman can also load it
    sudo chmod 644 /etc/containers/containers.conf
fi

# Guarantee the file exists even when DNS detection found nothing, so that the
# volume mount added to the runner below always has a real file to bind.
sudo mkdir -p /etc/containers
if [ ! -f /etc/containers/containers.conf ]; then
    printf '[containers]\n' | sudo tee /etc/containers/containers.conf > /dev/null
    sudo chmod 644 /etc/containers/containers.conf
fi

{{- if .LogToJournald}}
# Set journald as the container log driver so CI job output is captured by the
# systemd journal and can be correlated with runner daemon logs via job_id.
sudo mkdir -p /etc/containers
if [ ! -f /etc/containers/containers.conf ]; then
    printf '[containers]\nlog_driver = "journald"\n' \
        | sudo tee /etc/containers/containers.conf > /dev/null
elif grep -q '^\[containers\]' /etc/containers/containers.conf; then
    if awk '/^\[containers\]/{f=1;next} /^\[/{f=0} f && /^log_driver/{found=1} END{exit !found}' \
            /etc/containers/containers.conf; then
        # Replace existing log_driver within [containers]
        awk '/^\[containers\]/{s=1} /^\[/ && !/^\[containers\]/{s=0}
             s && /^log_driver/{$0="log_driver = \"journald\""} 1' \
            /etc/containers/containers.conf \
            | sudo tee /etc/containers/containers.conf.tmp > /dev/null \
            && sudo mv /etc/containers/containers.conf.tmp /etc/containers/containers.conf
    else
        sudo sed -i '/^\[containers\]/a log_driver = "journald"' \
            /etc/containers/containers.conf
    fi
else
    printf '\n[containers]\nlog_driver = "journald"\n' \
        | sudo tee -a /etc/containers/containers.conf > /dev/null
fi
sudo chmod 644 /etc/containers/containers.conf
{{- end}}

# Create an executor-specific containers.conf that adds a non-conflicting inner
# subnet for nested Netavark networks.  The host containers.conf intentionally
# omits [network] so the host Podman bridge keeps its default 10.88.0.0/16.
# The executor copy adds default_subnet = 192.168.100.0/24 so that Netavark
# inside a privileged executor container creates a bridge in a different subnet,
# eliminating the duplicate-route conflict that breaks DNS in nested containers
# on Netavark-based hosts (RHEL 9 / ppc64le).
sudo cp /etc/containers/containers.conf /etc/containers/executor-containers.conf
printf '\n[network]\ndefault_subnet = "192.168.100.0/24"\n' \
    | sudo tee -a /etc/containers/executor-containers.conf > /dev/null
sudo chmod 644 /etc/containers/executor-containers.conf

# Enable IP forwarding so Netavark can NAT containers through the host's
# network interface. Persist via sysctl.d so the setting survives reboots.
printf 'net.ipv4.ip_forward = 1\nnet.ipv4.conf.all.forwarding = 1\n' \
    | sudo tee /etc/sysctl.d/99-podman-ip-forward.conf > /dev/null
sudo sysctl -w net.ipv4.ip_forward=1
sudo sysctl -w net.ipv4.conf.all.forwarding=1

# Ensure NAT masquerade is active for the Podman bridge subnet.
# On RHEL/firewalld systems, Netavark normally configures this, but
# 'podman system reset' can leave firewalld without the masquerade rule
# until the first container is actually created — too late for the runner
# to resolve DNS at job startup. We add the rule explicitly so it is in
# place before any job container tries to reach an external DNS server.
sudo iptables -t nat -A POSTROUTING \
    -s 10.88.0.0/16 ! -d 10.88.0.0/16 -j MASQUERADE 2>/dev/null || true
# On firewalld systems (RHEL/Fedora), enable masquerade permanently so it
# survives firewalld restarts and reboots, then reload to activate immediately.
sudo firewall-cmd --permanent --add-masquerade 2>/dev/null || true
sudo firewall-cmd --reload 2>/dev/null || true

# Register runner using docker executor backed by Podman
# --docker-privileged is required for Podman: containers need CAP_SYS_ADMIN to mount /proc
sudo gitlab-runner register \
  --non-interactive \
  --url "{{ .RepoURL }}" \
  --token "{{ .Token }}" \
  --name "{{ .Name }}" \
  --executor "docker" \
  --docker-image "fedora:latest" \
  --docker-host "unix:///run/podman/podman.sock" \
  --docker-privileged \
  "${_docker_dns_args[@]}" \
  --docker-volumes "/etc/containers/executor-containers.conf:/etc/containers/containers.conf:ro"

{{- if not .Unsecure}}
# Create a dedicated system user for running CI jobs
sudo useradd --system \
  --shell /bin/bash \
  --create-home \
  --home-dir /home/gitlab-runner \
  gitlab-runner

RUNNER_USER=gitlab-runner
{{- else}}
RUNNER_USER={{ .User }}
{{- end}}

# Install and start as service
sudo gitlab-runner install --user="${RUNNER_USER}"
{{- if .Concurrent}}
sudo sed -i "s/^concurrent = .*/concurrent = {{.Concurrent}}/" /etc/gitlab-runner/config.toml
{{- end}}
# Increase per-runner log limit (default 4 MB is too small for long builds like PyTorch)
sudo sed -i '/^\[\[runners\]\]/a\  output_limit = 65536' /etc/gitlab-runner/config.toml
sudo systemctl daemon-reload
sudo systemctl enable --now gitlab-runner
