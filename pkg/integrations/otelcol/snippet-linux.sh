#!/bin/bash
set -euo pipefail

PROXY_URL=""
if ! curl -sf --connect-timeout 5 --head {{.Endpoint}} > /dev/null 2>&1; then
  PROXY_URL="http://squid.corp.redhat.com:3128"
fi

# Download binary tarball — works on any Linux distro, no package manager needed
TAR_URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v{{.ColVersion}}/otelcol-contrib_{{.ColVersion}}_linux_{{.Arch}}.tar.gz"
HTTPS_PROXY="$PROXY_URL" curl -fsSL -o /tmp/otelcol-contrib.tar.gz "$TAR_URL"
tar -xzf /tmp/otelcol-contrib.tar.gz -C /tmp otelcol-contrib
mv /tmp/otelcol-contrib /usr/local/bin/otelcol-contrib
chmod 755 /usr/local/bin/otelcol-contrib
chcon -t bin_t /usr/local/bin/otelcol-contrib 2>/dev/null || restorecon -v /usr/local/bin/otelcol-contrib 2>/dev/null || true
rm -f /tmp/otelcol-contrib.tar.gz

# Create dedicated system user (idempotent)
useradd --system --no-create-home --shell /sbin/nologin otelcol-contrib 2>/dev/null || true

# Create config and drop-in directories
mkdir -p /etc/otelcol-contrib /etc/systemd/system/otelcol-contrib.service.d

# Write systemd service unit
cat > /etc/systemd/system/otelcol-contrib.service << 'SVCEOF'
[Unit]
Description=OpenTelemetry Collector Contrib
After=network.target

[Service]
User=otelcol-contrib
Group=otelcol-contrib
EnvironmentFile=/etc/otelcol-contrib/otelcol-contrib.conf
EnvironmentFile=/etc/otelcol-contrib/auth_token
ExecStart=/usr/local/bin/otelcol-contrib $OTELCOL_OPTIONS
Restart=on-failure
RestartSec=5s
KillMode=process
SyslogIdentifier=otelcol-contrib

[Install]
WantedBy=multi-user.target
SVCEOF

# Drop-in: grant read access to all log files and expose hostname
cat > /etc/systemd/system/otelcol-contrib.service.d/capabilities.conf << 'CAPEOF'
[Service]
AmbientCapabilities=CAP_DAC_READ_SEARCH
Environment="HOSTNAME=%H"
CAPEOF

# Options file consumed by ExecStart
printf 'OTELCOL_OPTIONS=--config /etc/otelcol-contrib/config.yaml\n' \
  > /etc/otelcol-contrib/otelcol-contrib.conf
chmod 640 /etc/otelcol-contrib/otelcol-contrib.conf

# Auth token (mode 600 — readable only by otelcol-contrib user after chown)
printf 'OTEL_AUTH_TOKEN={{.AuthToken}}\n' > /etc/otelcol-contrib/auth_token
chmod 600 /etc/otelcol-contrib/auth_token

# Collector configuration
cat > /etc/otelcol-contrib/config.yaml << 'OTELEOF'
receivers:
  filelog/syslog:
    include:
    - {{.SyslogPath}}
    start_at: end
    include_file_path: true
    include_file_name: true
    exclude_older_than: 24h
    operators:
    - type: move
      id: move_to_source_name
      from: attributes["log.file.path"]
      to: attributes["_sourceName"]
    - type: remove
      id: remove_file_name
      field: attributes["log.file.name"]
    - type: time_parser
      id: parse_timestamp
      layout: '%b %e %H:%M:%S'
      parse_from: body
      on_error: send
    attributes:
      index: "{{.Index}}"
      _sourceCategory: syslog
      _sourceHost: ${env:HOSTNAME}
  filelog/secure:
    include:
    - {{.SecurePath}}
    start_at: end
    include_file_path: true
    include_file_name: true
    exclude_older_than: 24h
    operators:
    - type: move
      id: move_to_source_name
      from: attributes["log.file.path"]
      to: attributes["_sourceName"]
    - type: remove
      id: remove_file_name
      field: attributes["log.file.name"]
    - type: time_parser
      id: parse_timestamp
      layout: '%b %e %H:%M:%S'
      parse_from: body
      on_error: send
    attributes:
      index: "{{.Index}}"
      _sourceCategory: secure
      _sourceHost: ${env:HOSTNAME}
  filelog/audit:
    include:
    - /var/log/audit/audit.log
    start_at: end
    include_file_path: true
    include_file_name: true
    exclude_older_than: 24h
    operators:
    - type: move
      id: move_to_source_name
      from: attributes["log.file.path"]
      to: attributes["_sourceName"]
    - type: remove
      id: remove_file_name
      field: attributes["log.file.name"]
    attributes:
      index: "{{.Index}}"
      _sourceCategory: audit
      _sourceHost: ${env:HOSTNAME}
{{- if .MonitorGitLabRunner}}
  filelog/gitlab-runner:
    include:
    - /var/log/gitlab-runner/runner.log
    start_at: end
    include_file_path: true
    include_file_name: true
    operators:
    - type: move
      id: move_to_source_name
      from: attributes["log.file.path"]
      to: attributes["_sourceName"]
    - type: remove
      id: remove_file_name
      field: attributes["log.file.name"]
    - type: regex_parser
      id: parse_job_id
      parse_from: body
      regex: '\bjob=(?P<job_id>\d+)'
      on_error: send
    - type: regex_parser
      id: parse_runner_token
      parse_from: body
      regex: '\brunner=(?P<runner_token>\w+)'
      on_error: send
    attributes:
      index: "{{.Index}}"
      _sourceCategory: gitlab-runner
      _sourceHost: ${env:HOSTNAME}
  journald/gitlab-jobs:
    operators:
    - type: regex_parser
      id: parse_container_name
      parse_from: attributes["CONTAINER_NAME"]
      regex: '^runner-(?P<runner_token>.+?)-project-(?P<project_id>\d+)-concurrent-(?P<concurrent_id>\d+)-(?P<job_id>\d+)$'
      on_error: send
    attributes:
      index: "{{.Index}}"
      _sourceCategory: gitlab-runner-jobs
      _sourceHost: ${env:HOSTNAME}
{{- end}}
processors:
  filter/drop_null_bytes:
    logs:
      log_record:
        - 'IsMatch(body, "^\x00+$")'
  batch:
    timeout: "1s"
    send_batch_size: 1024
  resource:
    attributes:
      - key: appcode
        value: "{{.AppCode}}"
        action: upsert
      - key: com.redhat.otel.auth_token
        value: "${env:OTEL_AUTH_TOKEN}"
        action: upsert
      - key: arch
        value: "{{.Arch}}"
        action: upsert
{{- range $k, $v := .ExtraAttrs}}
      - key: {{$k}}
        value: "{{$v}}"
        action: upsert
{{- end}}
exporters:
  otlphttp:
    endpoint: "{{.Endpoint}}"
    tls:
      insecure_skip_verify: true
service:
  telemetry:
    logs:
      level: "fatal"
    metrics:
      level: "basic"
  pipelines:
    logs:
      receivers: [filelog/syslog, filelog/secure, filelog/audit{{if .MonitorGitLabRunner}}, filelog/gitlab-runner, journald/gitlab-jobs{{end}}]
      processors: [filter/drop_null_bytes, resource, batch]
      exporters: [otlphttp]
OTELEOF

# Transfer ownership to the service user
chown -R otelcol-contrib:otelcol-contrib /etc/otelcol-contrib

# Set up proxy drop-in if direct access to the endpoint was unavailable
if [ -n "$PROXY_URL" ]; then
  printf '[Service]\nEnvironment="HTTPS_PROXY=%s/"\nEnvironment="NO_PROXY=10.*,192.168.*,localhost,127.0.0.1"\n' "$PROXY_URL" \
    > /etc/systemd/system/otelcol-contrib.service.d/proxy.conf
fi

systemctl daemon-reload
systemctl enable --now otelcol-contrib
