#!/bin/bash
# Download GitLab Runner
sudo curl -L -o /usr/local/bin/gitlab-runner "{{ .CliURL }}"
sudo chmod +x /usr/local/bin/gitlab-runner

# Register runner
sudo gitlab-runner register \
  --non-interactive \
  --url "{{ .RepoURL }}" \
  --token "{{ .Token }}" \
  --executor "shell"

# Install and start as LaunchDaemon
sudo gitlab-runner install
sudo gitlab-runner start
