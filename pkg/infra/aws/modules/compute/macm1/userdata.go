package macm1

import (
	"bytes"
	"text/template"
)

var script string = `
#!/bin/sh

# Enable remote control (vnc)
defaults write /var/db/launchd.db/com.apple.launchd/overrides.plist com.apple.screensharing -dict Disabled -bool false
launchctl load -w /System/Library/LaunchDaemons/com.apple.screensharing.plist

# Set user password
dscl . -passwd /Users/{{.Username}} {{.Password}}

# Autologin
curl -o /tmp/kcpassword https://raw.githubusercontent.com/xfreebird/kcpassword/master/kcpassword
chmod +x /tmp/kcpassword
/tmp/kcpassword {{.Password}}
defaults write /Library/Preferences/com.apple.loginwindow autoLoginUser "{{.Username}}"

defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.securitypref.logoutvalue -int 1200
defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.autologout.AutoLogOutDelay -int 1200

# autologin to take effect
# This breaks script
reboot
`

type UserDataValues struct {
	Username string
	Password string
}

func getUserData(username, password string) (string, error) {
	data := UserDataValues{username, password}
	tmpl, err := template.New("userdata").Parse(script)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
