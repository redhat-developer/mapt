package macm1

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"

	"bytes"
	"fmt"
	"text/template"

	"github.com/adrianriobo/qenvs/pkg/infra"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
)

const vncDefaultPort int = 5900

func (r *MacM1Request) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return ami.GetAMIByName(ctx, r.Specs.AMI.RegexName, r.Specs.AMI.Owner, r.Specs.AMI.Filters)
}

func (r *MacM1Request) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return ec2.NewDedicatedHost(ctx,
		r.GetName(),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(r.AvailabilityZones[0]),
			InstanceType:     pulumi.String(r.Specs.InstaceTypes[0]),
		})
}

func (r *MacM1Request) CustomIngressRules() []securityGroup.IngressRules {
	return []securityGroup.IngressRules{
		{
			Description: fmt.Sprintf("VNC port for %s", r.Specs.ID),
			FromPort:    vncDefaultPort,
			ToPort:      vncDefaultPort,
			Protocol:    "tcp",
			CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
		},
	}
}

func (r *MacM1Request) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *MacM1Request) GetPostScript() (string, error) {
	return getUserData(r.Specs.AMI.DefaultUser, "crcqe")
}

var script string = `
#!/bin/sh

# Enable remote control (vnc)
sudo defaults write /var/db/launchd.db/com.apple.launchd/overrides.plist com.apple.screensharing -dict Disabled -bool false
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.screensharing.plist

# Set user password
sudo dscl . -passwd /Users/{{.Username}} {{.Password}}

# Autologin
sudo curl -o /tmp/kcpassword https://raw.githubusercontent.com/xfreebird/kcpassword/master/kcpassword
sudo chmod +x /tmp/kcpassword
sudo /tmp/kcpassword {{.Password}}
sudo defaults write /Library/Preferences/com.apple.loginwindow autoLoginUser "{{.Username}}"

sudo defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.securitypref.logoutvalue -int 1200
sudo defaults write /Library/Preferences/.GlobalPreferences.plist com.apple.autologout.AutoLogOutDelay -int 1200

# autologin to take effect
# run reboot on background to successfully finish the remote exec of the script
(sleep 2 && sudo reboot)&
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
