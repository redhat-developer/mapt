# Download cirrus
curl -o /usr/local/bin/cirrus \
    -L {{.CliURL}}
chmod +x /usr/local/bin/cirrus
# Create service
cd /etc/systemd/system
cat << EOF > cirrus.service
[Unit]
Description=cirrus client
After=network.target
[Service]
User={{.User}}
WorkingDirectory=/home/{{.User}}
ExecStart=/usr/local/bin/cirrus worker run --name {{.Name}} \
                            --token {{.Token}}{{ if .Labels }} \
                            --labels {{.Labels}}{{ end }}
Restart=always
RestartSec=5
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable cirrus.service
systemctl start cirrus.service
systemctl status cirrus.service