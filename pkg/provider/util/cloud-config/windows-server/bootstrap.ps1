<powershell>
# Change password for user
$Password = ConvertTo-SecureString "{{.Password}}" -AsPlainText -Force
$UserAccount = Get-LocalUser -Name "{{.Username}}"
$UserAccount | Set-LocalUser -Password $Password
# Also need to set new password on autologin
$RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
Set-ItemProperty $RegistryPath 'DefaultPassword' -Value "{{.Password}}" -type String

# On windows server 2022 need to ensure unencrypted communications are allowed for this setup
cd WSMan:\localhost\Client
Set-Item .\allowunencrypted $true
# now Create a session
$UserName = "{{.Username}}"
$Cred = New-Object System.Management.Automation.PSCredential ($UserName, $Password)
$s = New-PSSession -ComputerName localhost -Credential $Cred -Authentication Basic
# Set the authorized keys according to the private key
Invoke-Command -ScriptBlock { New-Item -Path C:\Users\{{.Username}}\.ssh -Name "authorized_keys" -ItemType "file" -Value "{{.AuthorizedKey}}" -Force } -Session $s
# Set back disable unnecrypted communications
Set-Item .\allowunencrypted $false

# Install github-actions-runner if needed
{{ if .ActionsRunnerSnippet }}$ghToken = {{ .RunnerToken }}
    {{- .ActionsRunnerSnippet }}
{{ end }}

{{ if .CirrusSnippet }}
$cirrusToken = {{ .CirrusToken }}
{{ .CirrusSnippet }}
{{ end }}

{{ if .GitLabSnippet }}
$gitlabToken = {{ .GitLabToken }}
{{ .GitLabSnippet }}
{{ end }}

netsh advfirewall firewall add rule name="Open SSH Port 22" dir=in action=allow protocol=TCP localport=22 remoteip=any
# Restart computer to have the ssh connection available with setup from this script
Start-Process powershell -verb runas -ArgumentList "Restart-Computer -Force"

</powershell>
