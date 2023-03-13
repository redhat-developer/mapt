package macm1

var script string = `
#!/bin/sh

# Allow run x86 binaries on arm64 
sudo softwareupdate --install-rosetta --agree-to-license

# Enable remote control (vnc)
sudo defaults write /var/db/launchd.db/com.apple.launchd/overrides.plist com.apple.screensharing -dict Disabled -bool false
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.screensharing.plist

# Set user password
sudo dscl . -passwd /Users/{{.Username}} "{{.Password}}"

# Autologin
sudo curl -o /tmp/kcpassword https://raw.githubusercontent.com/xfreebird/kcpassword/master/kcpassword
sudo chmod +x /tmp/kcpassword
sudo /tmp/kcpassword "{{.Password}}"
sudo defaults write /Library/Preferences/com.apple.loginwindow autoLoginUser "{{.Username}}"

sudo defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.securitypref.logoutvalue -int 0
sudo defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.autologout.AutoLogOutDelay -int 0

# autologin to take effect
# run reboot on background to successfully finish the remote exec of the script
(sleep 2 && sudo reboot)&
`

type scriptDataValues struct {
	Username string
	Password string
}
