# Create local user
$Password = ConvertTo-SecureString "$Env:PASSWORD" -AsPlainText -Force
New-LocalUser $Env:USERNAME -Password $Password
# Run a process with new local user to create profile, so it will create home folder
$credential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $Env:USERNAME, $Password
Start-Process cmd /c -WindowStyle Hidden -Wait -Credential $credential -ErrorAction SilentlyContinue
# Add user to required groups
# Administrators group 
Add-LocalGroupMember -Member $Env:USERNAME -SID S-1-5-32-544

# Check if this speed insall of crc...if msi installer checks if no reboot required
# It is suppose doing os we do not need restart on install
Add-LocalGroupMember -Member $Env:USERNAME -SID S-1-5-32-578

# Set autologon to user to allow start sshd for the user
# Check requirements for domain user
# https://docs.microsoft.com/en-us/troubleshoot/windows-server/user-profiles-and-logon/turn-on-automatic-logon
$RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
Set-ItemProperty $RegistryPath 'AutoAdminLogon' -Value "1" -Type String
Set-ItemProperty $RegistryPath 'DefaultUsername' -Value "$Env:USERNAME" -type String
Set-ItemProperty $RegistryPath 'DefaultPassword' -Value "$Env:PASSWORD" -type String

# Install HyperV
$osProductType = Get-ComputerInfo | select -ExpandProperty OSProductType | Out-String -Stream | Where { $_.Trim().Length -gt 0 }
switch ($osProductType)
{
    "WorkStation" {Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All -NoRestart}
    "Server" {Install-WindowsFeature -Name Hyper-V -IncludeManagementTools}
}

# Install sshd
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Set-Service -Name sshd -StartupType 'Manual'
# This generate ssh certs + config file for us
Start-Service sshd
# Disable the service as need to start it as a user process on startup
Stop-Service sshd
# Add pub key for the user as authorized_key
New-Item -Path "C:\Users\$Env:USERNAME\.ssh" -ItemType Directory -Force
New-Item -Path C:\Users\$Env:USERNAME\.ssh -Name "authorized_keys" -ItemType "file" -Value "$Env:AUTHORIZEDKEY"
# Set permissions valid permissions for hyper_user on authorized_keys + host_keys
$acl = Get-Acl C:\Users\$Env:USERNAME\.ssh\authorized_keys
$acl.SetOwner([System.Security.Principal.NTAccount] "$Env:USERNAME")
$acl.SetAccessRuleProtection($True, $False)
$AccessRule = New-Object System.Security.AccessControl.FileSystemAccessRule([System.Security.Principal.NTAccount] "$Env:USERNAME","FullControl","Allow")
$acl.SetAccessRule($AccessRule)
Set-Acl C:\Users\$Env:USERNAME\.ssh\authorized_keys $acl
Set-Acl -Path "C:\ProgramData\ssh\*key" $acl
# Create bat script to start sshd as a user process on startup
New-Item -Path "C:\Users\$Env:USERNAME\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup" -Name start-openssh.bat -ItemType "file" -Value 'powershell -command "sshd -f C:\ProgramData\ssh\sshd_config"'

# Disable UAC
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "ConsentPromptBehaviorAdmin" -Value "0"
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "ConsentPromptBehaviorUser" -Value "3"
Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name "EnableLUA" -Value "1"

# Install powershellcore
curl.exe -LO https://github.com/PowerShell/PowerShell/releases/download/v7.3.0/PowerShell-7.3.0-win-x64.msi
Start-Process C:\Windows\System32\msiexec.exe -ArgumentList '/qb /i PowerShell-7.3.0-win-x64.msi ADD_EXPLORER_CONTEXT_MENU_OPENPOWERSHELL=1 ENABLE_PSREMOTING=1 REGISTER_MANIFEST=1 USE_MU=1 ENABLE_MU=1 ADD_PATH=1' -wait
# Set powershell as default shell on openssh
New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value "C:\Program Files\PowerShell\7\pwsh.exe" -PropertyType String -Force

