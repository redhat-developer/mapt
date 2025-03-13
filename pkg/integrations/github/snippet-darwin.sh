mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-osx.tar.gz -L {{ .CliURL }}
tar xzf ./actions-runner-osx.tar.gz
./config.sh --token {{ .Token }} --url {{ .RepoURL }} --name {{ .Name }} --unattended --replace --labels {{ .Labels }}
./svc.sh install
plistName=$(basename $(./svc.sh status | grep "plist$"))
mkdir -p /Library/LaunchDaemons
mv ~/Library/LaunchAgents/"${plistName}" /Library/LaunchDaemons/"${plistName}"

plutil -replace UserName -string {{ .User }} /Library/LaunchDaemons/"${plistName}"
plutil -replace StandardOutPath -string /tmp/actions.runner.github.out.log /Library/LaunchDaemons/"${plistName}"
plutil -replace StandardErrorPath -string /tmp/actions.runner.github.err.log /Library/LaunchDaemons/"${plistName}"

launchctl load /Library/LaunchDaemons/"${plistName}"
