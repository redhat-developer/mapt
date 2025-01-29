# Download cirrus cli
curl.exe -o C:\Users\{{.User}}\AppData\Local\Microsoft\WindowsApps\cirrus.exe -L {{.CliURL}}

# Create startup bat to run cirrus
$cirrusCMD="cirrus.exe worker run --name {{.Name}} --token $cirrusToken "
{{ if .Labels }}
$cirrusCMD="$cirrusCMD --labels {{.Labels}}"
{{ end }}
New-Item -Path "C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp" -Name start-cirrus.bat -ItemType "file" -Value "powershell -command `"$cirrusCMD`""
