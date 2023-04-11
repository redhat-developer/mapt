param(
    [Parameter(Mandatory,HelpMessage='password for the user to be set for autologin')]
    $userPass,
    [Parameter(Mandatory,HelpMessage='name for the user to be set for autologin')]
    $user,
    [Parameter(Mandatory,HelpMessage='hostname for the current machine')]
    $hostname,
    [Parameter(Mandatory,HelpMessage='authorizedkey for ssh private key for the user')]
    $authorizedKey
)
# Change password for user
$Password = ConvertTo-SecureString $userPass -AsPlainText -Force
$UserAccount = Get-LocalUser -Name "$user"
$UserAccount | Set-LocalUser -Password $Password

# Also need to set new password on autologin
$RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
Set-ItemProperty $RegistryPath -Name AutoAdminLogon -Value 1 -type String
Set-ItemProperty $RegistryPath -Name DefaultUsername -Value "$user" -type String 
Set-ItemProperty $RegistryPath -Name DefaultPassword -Value $userPass -type String

# Disable initial setup for devices on first startup
$RegistryPath = 'HKLM:\Software\Policies\Microsoft\Windows\OOBE'
New-Item $RegistryPath
Set-ItemProperty $RegistryPath -Name DisablePrivacyExperience -Value 1 -type DWord

# Enable localhost remote connection from locally
Set-Item WSMan:\localhost\Client\TrustedHosts -Value "$hostname" -Force
Set-NetConnectionProfile -NetworkCategory Private
Enable-PSRemoting -SkipNetworkProfileCheck -Force

# Set the authorized keys according to the private key
# Due to acl on the files Adminstrator can not write authorized_keys
# So we need to invoke it as the default user
$UserName = "localhost\$user"
$Cred = New-Object System.Management.Automation.PSCredential ($UserName, $Password)
$parameters = @{
    ComputerName = 'localhost'
    Credential = $Cred
    ScriptBlock  = {
        Param ($param1, $param2)
	    New-Item -Path "C:\Users\$param1\.ssh" -Name "authorized_keys" -ItemType "file" -Value "$param2" -Force
    }
    ArgumentList = "$user", "$authorizedKey"
}
Invoke-Command -asjob @parameters
Get-Job | Wait-Job

# Restart computer to have the ssh connection available with setup from this script
Start-Process powershell -verb runas -ArgumentList "Restart-Computer -Force"

# Finish success
Exit 0 
