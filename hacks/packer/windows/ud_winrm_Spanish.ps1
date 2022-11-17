<powershell>
# Set administrator password
net user Administrador SuperS3cr3t!!!!
wmic useraccount where "name='Administrador'" set PasswordExpires=FALSE

# First, make sure WinRM can't be connected to
netsh advfirewall firewall set rule name="Administración remota de Windows - Modo de compatibilidad (HTTP de entrada)" new enable=yes action=block

# Delete any existing WinRM listeners
winrm delete winrm/config/listener?Address=*+Transport=HTTP  2>$Null
winrm delete winrm/config/listener?Address=*+Transport=HTTPS 2>$Null

# Disable group policies which block basic authentication and unencrypted login

Set-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\WinRM\Client -Name AllowBasic -Value 1
Set-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\WinRM\Client -Name AllowUnencryptedTraffic -Value 1
Set-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\WinRM\Service -Name AllowBasic -Value 1
Set-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\WinRM\Service -Name AllowUnencryptedTraffic -Value 1


# Create a new WinRM listener and configure
winrm create winrm/config/listener?Address=*+Transport=HTTP
winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="0"}'
winrm set winrm/config '@{MaxTimeoutms="7200000"}'
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
winrm set winrm/config/service '@{MaxConcurrentOperationsPerUser="12000"}'
winrm set winrm/config/service/auth '@{Basic="true"}'
winrm set winrm/config/client/auth '@{Basic="true"}'

# Configure UAC to allow privilege elevation in remote shells
$Key = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System'
$Setting = 'LocalAccountTokenFilterPolicy'
Set-ItemProperty -Path $Key -Name $Setting -Value 1 -Force

# Configure and restart the WinRM Service; Enable the required firewall exception
Stop-Service -Name WinRM
Set-Service -Name WinRM -StartupType Automatic
netsh advfirewall firewall set rule name="Administración remota de Windows - Modo de compatibilidad (HTTP de entrada)" new action=allow localip=any remoteip=any
netsh advfirewall firewall set rule group=”Detección de redes” new enable=yes
Start-Service -Name WinRM
Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled False
</powershell>


