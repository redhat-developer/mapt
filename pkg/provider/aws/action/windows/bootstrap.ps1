<powershell>
# Change password for user
$Password = ConvertTo-SecureString "{{.Password}}" -AsPlainText -Force
$UserAccount = Get-LocalUser -Name "{{.Username}}"
$UserAccount | Set-LocalUser -Password $Password
# Also need to set new password on autologin
$RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
Set-ItemProperty $RegistryPath 'DefaultPassword' -Value "{{.Password}}" -type String

# Set the authorized keys according to the private key
# Due to acl on the files Adminstrator can not write authorized_keys
# So we need to invoke it as the default user
$UserName = "{{.Username}}"
$Cred = New-Object System.Management.Automation.PSCredential ($UserName, $Password)
Invoke-Command -ScriptBlock { New-Item -Path C:\Users\{{.Username}}\.ssh -Name "authorized_keys" -ItemType "file" -Value "{{.AuthorizedKey}}" -Force } -Credential $Cred -computername localhost
</powershell>