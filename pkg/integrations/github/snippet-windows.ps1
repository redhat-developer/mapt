New-Item -Path C:\actions-runner -Type Directory ; cd C:\actions-runner
Invoke-WebRequest -Uri {{ .CliURL }} -OutFile actions-runner-win.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win.zip", "$PWD")
{{ if .JITConfig -}}
# A startup task brings the runner back after the reboot that ends the setup; no time limit, it waits for its job.
# Its wrapper deletes the .runner and .credentials* files a previous start left behind: the runner marks them hidden
# and then can't overwrite them, so every later --jitconfig start would fail. Only SYSTEM and Administrators can read it.
Set-Content -Path "$PWD\start-jit.cmd" -Value "@cd /d `"%~dp0`"`r`n@del /f /q /a .runner .credentials .credentials_rsaparams 2>nul`r`n@call run.cmd --jitconfig {{ .JITConfig }}"
icacls "$PWD\start-jit.cmd" /inheritance:r /grant:r "*S-1-5-18:F" "*S-1-5-32-544:F" | Out-Null
$runnerAction = New-ScheduledTaskAction -Execute "$PWD\start-jit.cmd" -WorkingDirectory "$PWD"
$runnerSettings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero)
Register-ScheduledTask -TaskName "actions-runner-jit" -Action $runnerAction -Trigger (New-ScheduledTaskTrigger -AtStartup) -Settings $runnerSettings -User "SYSTEM" -RunLevel Highest -Force | Out-Null
Start-ScheduledTask -TaskName "actions-runner-jit"
{{- else -}}
./config.cmd --token $ghToken --url {{ .RepoURL }} --name {{ .Name }} --unattended --runasservice --replace --labels {{ .Labels }}{{ if .Ephemeral }} --ephemeral{{ end }}
{{- end }}
