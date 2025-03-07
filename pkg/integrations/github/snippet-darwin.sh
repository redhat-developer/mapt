mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-osx.tar.gz -L {{ .RunnerURL }}
tar xzf ./actions-runner-osx.tar.gz
./config.sh --token {{ .Token }} --url {{ .RepoURL }} --name {{ .Name }} --unattended --replace --labels {{ .Labels }}
./svc.sh install
plistName=$(basename $(./svc.sh status | grep "plist$"))
mkdir -p ~/Library/LaunchDaemons
mv ~/Library/LaunchAgents/"${plistName}" ~/Library/LaunchDaemons/"${plistName}"
launchctl load ~/Library/LaunchDaemons/"${plistName}"
