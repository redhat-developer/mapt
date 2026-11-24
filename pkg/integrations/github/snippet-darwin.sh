mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-osx.tar.gz -L {{ .CliURL }}
tar xzf ./actions-runner-osx.tar.gz
{{ if .JITConfig -}}
# A LaunchDaemon starts the runner now and again after the reboot that ends the setup.
# env.sh saves this session's PATH for jobs, since launchd starts the runner with a minimal one.
./env.sh
sudo tee /Library/LaunchDaemons/actions.runner.jit.plist > /dev/null << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>actions.runner.jit</string>
  <key>ProgramArguments</key>
  <array>
    <string>$PWD/run.sh</string>
    <string>--jitconfig</string>
    <string>{{ .JITConfig }}</string>
  </array>
  <key>WorkingDirectory</key><string>$PWD</string>
  <key>UserName</key><string>{{ .User }}</string>
  <key>RunAtLoad</key><true/>
  <key>StandardOutPath</key><string>/tmp/actions.runner.github.out.log</string>
  <key>StandardErrorPath</key><string>/tmp/actions.runner.github.err.log</string>
</dict>
</plist>
EOF
sudo chmod 600 /Library/LaunchDaemons/actions.runner.jit.plist
sudo launchctl load -w /Library/LaunchDaemons/actions.runner.jit.plist
{{- else -}}
./config.sh --token {{ .Token }} --url {{ .RepoURL }} --name {{ .Name }} --unattended --replace --labels {{ .Labels }}{{ if .Ephemeral }} --ephemeral{{ end }}
./svc.sh install
plistName=$(basename $(./svc.sh status | grep "plist$"))
mkdir -p /Library/LaunchDaemons
mv ~/Library/LaunchAgents/"${plistName}" /Library/LaunchDaemons/"${plistName}"

plutil -replace UserName -string {{ .User }} /Library/LaunchDaemons/"${plistName}"
plutil -replace StandardOutPath -string /tmp/actions.runner.github.out.log /Library/LaunchDaemons/"${plistName}"
plutil -replace StandardErrorPath -string /tmp/actions.runner.github.err.log /Library/LaunchDaemons/"${plistName}"

launchctl load /Library/LaunchDaemons/"${plistName}"
{{- end }}
