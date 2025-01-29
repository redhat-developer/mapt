# Download cirrus cli
sudo curl -o /usr/local/bin/cirrus -L {{.CliURL}}
sudo chmod +x /usr/local/bin/cirrus

# Create launchd

cd /tmp

cat << EOF > com.cirrus.client.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.cirrus.client</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/cirrus</string>
        <string>worker</string>
        <string>run</string>
        <string>--name</string>
        <string>{{.Name}}</string>
        <string>--token</string>
        <string>{{.Token}}</string>
{{ if .Labels }}
        <string>--labels</string>
        <string>{{.Labels}}</string>
{{ end }}
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>UserName</key>
    <string>{{.User}}</string>
    <key>GroupName</key>
    <string>staff</string>  
    <key>StandardOutPath</key>
    <string>/tmp/com.cirrus.client.out.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/com.cirrus.client.err.log</string>
</dict>
</plist>
EOF

sudo mv com.cirrus.client.plist /Library/LaunchDaemons/
sudo chown root:wheel /Library/LaunchDaemons/com.cirrus.client.plist
sudo chmod 644 /Library/LaunchDaemons/com.cirrus.client.plist
sudo launchctl load /Library/LaunchDaemons/com.cirrus.client.plist