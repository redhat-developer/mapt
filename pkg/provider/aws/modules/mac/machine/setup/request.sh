#!/bin/zsh

# Change password
cat << EOF > change_password.exp
#!/usr/bin/expect

set old_password "{{.OldPassword}}"
set new_password "{{.NewPassword}}"

spawn sudo dscl . -passwd /Users/{{.Username}}
expect "New Password:"
send "$new_password\r"
expect "Permission denied. Please enter user's old password:"
send "$old_password\r"
expect eof
EOF
chmod +x change_password.exp
./change_password.exp
rm change_password.exp

# Autologin
sudo curl -o /tmp/kcpassword https://raw.githubusercontent.com/xfreebird/kcpassword/master/kcpassword
sudo chmod +x /tmp/kcpassword
sudo /tmp/kcpassword "{{.NewPassword}}"

# Override authorized key
mkdir -p /Users/{{.Username}}/.ssh
echo "{{.AuthorizedKey}}" | tee /Users/{{.Username}}/.ssh/authorized_keys

{{ if .InstallActionsRunner }}
    {{- .ActionsRunnerSnippet }}
{{ end }}

{{ .CirrusSnippet }}

# autologin to take effect
# run reboot on background to successfully finish the remote exec of the script
(sleep 2 && sudo reboot)&