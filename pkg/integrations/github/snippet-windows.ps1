New-Item -Path C:\actions-runner -Type Directory ; cd C:\actions-runner
Invoke-WebRequest -Uri {{ .CliURL }} -OutFile actions-runner-win.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win.zip", "$PWD")
{{ if .JITConfig -}}
# A startup task brings the runner back after the reboot that ends the setup; no time limit, it waits for its job
$runnerAction = New-ScheduledTaskAction -Execute "$PWD\run.cmd" -Argument "--jitconfig {{ .JITConfig }}" -WorkingDirectory "$PWD"
$runnerSettings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero)
Register-ScheduledTask -TaskName "actions-runner-jit" -Action $runnerAction -Trigger (New-ScheduledTaskTrigger -AtStartup) -Settings $runnerSettings -User "SYSTEM" -RunLevel Highest -Force | Out-Null
Start-ScheduledTask -TaskName "actions-runner-jit"
{{- else -}}
./config.cmd --token $ghToken --url {{ .RepoURL }} --name {{ .Name }} --unattended --runasservice --replace --labels {{ .Labels }}{{ if .Ephemeral }} --ephemeral{{ end }}
{{- end }}
