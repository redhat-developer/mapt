New-Item -Path C:\actions-runner -Type Directory ; cd C:\actions-runner
Invoke-WebRequest -Uri {{ .CliURL }} -OutFile actions-runner-win.zip
Add-Type -AssemblyName System.IO.Compression.FileSystem ;
[System.IO.Compression.ZipFile]::ExtractToDirectory("$PWD\actions-runner-win.zip", "$PWD")
./config.cmd --token $ghToken --url {{ .RepoURL }} --name {{ .Name }} --unattended --runasservice --replace --labels {{ .Labels }}
