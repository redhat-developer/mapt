#!/bin/sh

# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-mac-instances.html
# Internal disk will be accessible on workspace dir 
# mkdir -p "/Users/{{.Username}}/workspace"
echo 'if [[ ! -d /Volumes/InternalDisk ]]; then' | tee -a "/Users/{{.Username}}/.zshrc"
echo 'APFSVolumeName="InternalDisk" ; SSDContainer=$(diskutil list | grep "Physical Store disk0" -B 3 | grep "/dev/disk" | awk {'"'"'print $1'"'"'} ) ; diskutil apfs addVolume $SSDContainer APFS $APFSVolumeName' | tee -a "/Users/{{.Username}}/.zshrc"
echo 'sudo chown {{.Username}}:staff "/Volumes/InternalDisk"' | tee -a "/Users/{{.Username}}/.zshrc"
echo 'sudo chmod 0750 "/Volumes/InternalDisk"' | tee -a "/Users/{{.Username}}/.zshrc"
echo 'fi' | tee -a "/Users/{{.Username}}/.zshrc"

# Allow run x86 binaries on arm64 
# TODO review if still required for CrC
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
sudo sysadminctl -screenLock off -password "{{.Password}}"

# Override authorized key
mkdir -p /Users/{{.Username}}/.ssh
echo "{{.AuthorizedKey}}" | tee /Users/{{.Username}}/.ssh/authorized_keys

# Install github-actions-runner if needed
{{ if .InstallActionsRunner }}
    {{- .ActionsRunnerSnippet }}
{{ end }}

{{ .CirrusSnippet }}

# autologin to take effect
# run reboot on background to successfully finish the remote exec of the script
(sleep 2 && sudo reboot)&
