package windows

// type userDataValues struct {
// 	Username      string
// 	Password      string
// 	AuthorizedKey string
// }

// var userdata string = `
// <powershell>
// # Create local user
// $Password = ConvertTo-SecureString "{{.Password}}" -AsPlainText -Force
// New-LocalUser {{.Username}} -Password $Password
// # Run a process with new local user to create profile, so it will create home folder
// $credential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList {{.Username}}, $Password
// Start-Process cmd /c -WindowStyle Hidden -Wait -Credential $credential -ErrorAction SilentlyContinue
// # Add user to required groups
// Add-LocalGroupMember -Group "Administrators" -Member {{.Username}}
// # Check if this speed insall of crc...if msi installer checks if no reboot required
// # Non eng Add-LocalGroupMember -Group "Hyper-V Administrators" -Member {{.Username}}
// Add-LocalGroupMember -Member {{.Username}} -SID S-1-5-32-578
// # Set autologon to user to allow start sshd for the user
// # Check requirements for domain user
// # https://docs.microsoft.com/en-us/troubleshoot/windows-server/user-profiles-and-logon/turn-on-automatic-logon
// $RegistryPath = 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon'
// Set-ItemProperty $RegistryPath 'AutoAdminLogon' -Value "1" -Type String
// Set-ItemProperty $RegistryPath 'DefaultUsername' -Value "{{.Username}}" -type String
// Set-ItemProperty $RegistryPath 'DefaultPassword' -Value "{{.Password}}" -type String
// # Install sshd
// Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
// Set-Service -Name sshd -StartupType 'Manual'
// # This generate ssh certs + config file for us
// Start-Service sshd
// # Disable the service as need to start it as a user process on startup
// Stop-Service sshd
// # Add pub key for the user as authorized_key
// New-Item -Path "C:\Users\{{.Username}}\.ssh" -ItemType Directory -Force
// New-Item -Path C:\Users\{{.Username}}\.ssh -Name "authorized_keys" -ItemType "file" -Value "{{.AuthorizedKey}}"
// # Set permissions valid permissions for hyper_user on authorized_keys + host_keys
// $acl = Get-Acl C:\Users\{{.Username}}\.ssh\authorized_keys
// $acl.SetOwner([System.Security.Principal.NTAccount] "{{.Username}}")
// $acl.SetAccessRuleProtection($True, $False)
// $AccessRule = New-Object System.Security.AccessControl.FileSystemAccessRule([System.Security.Principal.NTAccount] "{{.Username}}","FullControl","Allow")
// $acl.SetAccessRule($AccessRule)
// Set-Acl C:\Users\{{.Username}}\.ssh\authorized_keys $acl
// Set-Acl -Path "C:\ProgramData\ssh\*key" $acl
// # Create bat script to start sshd as a user process on startup
// New-Item -Path "C:\Users\{{.Username}}\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup" -Name start-openssh.bat -ItemType "file" -Value 'powershell -command "sshd -f C:\ProgramData\ssh\sshd_config"'
// Restart-Computer
// </powershell>
// `
