# Restart computer to have the ssh connection available with setup from this script
Start-Process powershell -verb runas -ArgumentList "Restart-Computer -Force"
Exit 0