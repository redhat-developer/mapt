# Download GitLab Runner
New-Item -Path C:\GitLab-Runner -ItemType Directory -Force
Invoke-WebRequest -Uri "{{ .CliURL }}" -OutFile C:\GitLab-Runner\gitlab-runner.exe

# Register runner
C:\GitLab-Runner\gitlab-runner.exe register `
  --non-interactive `
  --url "{{ .RepoURL }}" `
  --token "{{ .Token }}" `
  --executor "shell"

# Install and start as Windows service
C:\GitLab-Runner\gitlab-runner.exe install
C:\GitLab-Runner\gitlab-runner.exe start
