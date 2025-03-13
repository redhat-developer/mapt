mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-linux.tar.gz -L {{ .CliURL }}
tar xzf ./actions-runner-linux.tar.gz
sudo ./bin/installdependencies.sh
./config.sh --token {{ .Token }} --url {{ .RepoURL }} --name {{ .Name }} --unattended --replace --labels {{ .Labels }}
sudo ./svc.sh install
chcon system_u:object_r:usr_t:s0 $(pwd)/runsvc.sh
sudo ./svc.sh start
