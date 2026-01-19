#!/bin/bash
set -e

# Download GitLab Runner
curl -L -o /tmp/gitlab-runner "{{ .CliURL }}"
chmod +x /tmp/gitlab-runner

# Move to trusted path
sudo mv /tmp/gitlab-runner /usr/bin/gitlab-runner

# Fix SELinux context
sudo restorecon -v /usr/bin/gitlab-runner

# Register runner
sudo gitlab-runner register \
  --non-interactive \
  --url "{{ .RepoURL }}" \
  --token "{{ .Token }}" \
  --executor "shell"

# Install and start as service
sudo gitlab-runner install --user={{ .User }}
sudo systemctl daemon-reload
sudo systemctl enable --now gitlab-runner
