param(
    [Parameter(Mandatory,HelpMessage='password for the user to be set for autologin')]
    $userPass,
    [Parameter(Mandatory,HelpMessage='name for the user to be set for autologin')]
    $user,
    [Parameter(Mandatory,HelpMessage='hostname for the current machine')]
    $hostname,
    [Parameter(Mandatory,HelpMessage='authorizedkey for ssh private key for the user')]
    $authorizedKey,
    [Parameter(HelpMessage='token for the github actions runner')]
    $ghToken,
    [Parameter(HelpMessage='token for the cirrus persistent worker')]
    $cirrusToken,
    [switch]$crcProfile=$false
)
# Create local user
$Password = ConvertTo-SecureString $userPass -AsPlainText -Force
New-LocalUser $user -Password $Password

# Add user to required groups
# Administrators group 
Add-LocalGroupMember -Member $user -SID S-1-5-32-544

# Check if this speed insall of crc...if msi installer checks if no reboot required
# It is suppose doing os we do not need restart on install
Add-LocalGroupMember -Member $user -SID S-1-5-32-578

# Enable localhost remote connection from locally
Set-NetConnectionProfile -NetworkCategory Private
Enable-PSRemoting -SkipNetworkProfileCheck -Force
Set-Item WSMan:\localhost\Client\TrustedHosts -Value "$hostname" -Force

# Credentials for the new user
$Cred = New-Object System.Management.Automation.PSCredential ("localhost\$user", $Password)
# $Cred = New-Object System.Management.Automation.PSCredential ($user, $Password)

# Run a process with new local user to create profile, so it will create home folder
$parameters = @{
    ComputerName = 'localhost'
    Credential = $Cred
    ScriptBlock  = {
        Start-Process cmd /c -WindowStyle Hidden -Wait
    }
}
Invoke-Command -asjob @parameters

# Set autologon to user to allow start sshd for the user
# Check requirements for domain user
# https://docs.microsoft.com/en-us/troubleshoot/windows-server/user-profiles-and-logon/turn-on-automatic-logon
# Also need to set new password on autologin
$RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
Set-ItemProperty $RegistryPath -Name AutoAdminLogon -Value 1 -type String
Set-ItemProperty $RegistryPath -Name DefaultUsername -Value "$user" -type String 
Set-ItemProperty $RegistryPath -Name DefaultPassword -Value $userPass -type String

# Disable initial setup for devices on first startup
$RegistryPath = 'HKLM:\Software\Policies\Microsoft\Windows\OOBE'
New-Item $RegistryPath
Set-ItemProperty $RegistryPath -Name DisablePrivacyExperience -Value 1 -type DWord

# Install HyperV
$osProductType = Get-ComputerInfo | select -ExpandProperty OSProductType | Out-String -Stream | Where { $_.Trim().Length -gt 0 }
switch ($osProductType)
{
    "WorkStation" {Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All -NoRestart}
    "Server" {Install-WindowsFeature -Name Hyper-V -IncludeManagementTools}
}

# Install wsl2
Enable-WindowsOptionalFeature -Online -NoRestart -FeatureName VirtualMachinePlatform 
Enable-WindowsOptionalFeature -Online -NoRestart -FeatureName Microsoft-Windows-Subsystem-Linux 
New-Item -Path "C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp" -Name wsl-install.bat -ItemType "file" -Value 'wsl --update'

# Install sshd
# Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
curl.exe -LO https://github.com/PowerShell/Win32-OpenSSH/releases/download/v9.5.0.0p1-Beta/OpenSSH-Win64-v9.5.0.0.msi
Start-Process C:\Windows\System32\msiexec.exe -ArgumentList '/qb /i OpenSSH-Win64-v9.5.0.0.msi' -wait
Set-Service -Name sshd -StartupType 'Manual'
# This generate ssh certs + config file for us
Start-Service sshd
# Disable the service as need to start it as a user process on startup
Stop-Service sshd
# Add pub key for the user as authorized_key
New-Item -Path "C:\Users\$user\.ssh" -ItemType Directory -Force
New-Item -Path C:\Users\$user\.ssh -Name "authorized_keys" -ItemType "file" -Value "$authorizedKey"
# Set permissions valid permissions for hyper_user on authorized_keys + host_keys
$acl = Get-Acl C:\Users\$user\.ssh\authorized_keys
$acl.SetOwner([System.Security.Principal.NTAccount] "$user")
$acl.SetAccessRuleProtection($True, $False)
$AccessRule = New-Object System.Security.AccessControl.FileSystemAccessRule([System.Security.Principal.NTAccount] "$user","FullControl","Allow")
$acl.SetAccessRule($AccessRule)
Set-Acl C:\Users\$user\.ssh\authorized_keys $acl
Set-Acl -Path "C:\ProgramData\ssh\*key" $acl
# Create bat script to start sshd as a user process on startup
# New-Item -Path "C:\Users\$Env:USERNAME\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup" -Name start-openssh.bat -ItemType "file" -Value 'powershell -command "sshd -f C:\ProgramData\ssh\sshd_config"'
New-Item -Path "C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp" -Name start-openssh.bat -ItemType "file" -Value 'powershell -command "sshd -f C:\ProgramData\ssh\sshd_config"'

# Disable UAC
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "ConsentPromptBehaviorAdmin" -Value "0"
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "ConsentPromptBehaviorUser" -Value "3"
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "EnableLUA" -Value "1"

# Uninstall OneDrive
$onedrive = "$env:SYSTEMROOT\SysWOW64\OneDriveSetup.exe"
If (!(Test-Path $onedrive)) {
    $onedrive = "$env:SYSTEMROOT\System32\OneDriveSetup.exe"
}
Start-Process $onedrive "/uninstall" -NoNewWindow -Wait

# Install powershellcore
curl.exe -LO https://github.com/PowerShell/PowerShell/releases/download/v7.4.2/PowerShell-7.4.2-win-x64.msi
Start-Process C:\Windows\System32\msiexec.exe -ArgumentList '/qb /i PowerShell-7.4.2-win-x64.msi ADD_EXPLORER_CONTEXT_MENU_OPENPOWERSHELL=1 ENABLE_PSREMOTING=1 REGISTER_MANIFEST=1 USE_MU=1 ENABLE_MU=1 ADD_PATH=1' -wait
# Set powershell as default shell on openssh
New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value "C:\Program Files\PowerShell\7\pwsh.exe" -PropertyType String -Force

# Remove curl alias
$profilePath="C:\Users\$user\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
if (!(Test-Path -Path $profilePath)) {
    New-Item -Path $profilePath -Force
}
Add-Content -Path $profilePath -Value "Remove-Item alias:curl"

# Profiles it can have several profiles at once

# crc profile; create crc-users and add the user
# this is to avoid reboot requirement when installing crc
if ($crcProfile) {
    New-LocalGroup -Name "crc-users"
    Add-LocalGroupMember -Group "crc-users" -Member $user
}

# Install github-actions-runner if needed
{{ .ActionsRunnerSnippet }}

{{ .CirrusSnippet }}

# Restart computer to have the ssh connection available with setup from this script
Start-Process powershell -verb runas -ArgumentList "Restart-Computer -Force"

# Finish success
Exit 0 
