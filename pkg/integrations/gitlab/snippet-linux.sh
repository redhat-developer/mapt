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

# Register runner using docker executor backed by Podman
# --docker-privileged is required for Podman: containers need CAP_SYS_ADMIN to mount /proc
sudo gitlab-runner register \
  --non-interactive \
  --url "{{ .RepoURL }}" \
  --token "{{ .Token }}" \
  --executor "docker" \
  --docker-image "fedora:latest" \
  --docker-host "unix:///run/podman/podman.sock" \
  --docker-privileged

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
sudo systemctl daemon-reload
sudo systemctl enable --now gitlab-runner
