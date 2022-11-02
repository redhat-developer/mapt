package snc

import (
	"encoding/base64"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	securityGroup "github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (r *SNCRequest) GetRequest() *compute.Request {
	return &r.Request
}

func (r *SNCRequest) GetAMI(ctx *pulumi.Context) (*ec2.LookupAmiResult, error) {
	return r.RHELRequest.GetAMI(ctx)
}

func (r *SNCRequest) GetUserdata(ctx *pulumi.Context) (pulumi.StringPtrInput, error) {
	userdata, err := util.Template(
		userDataValues{
			r.SubscriptionUsername,
			r.SubscriptionPassword},
		"userdata", cloudConfig)
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}

func (r *SNCRequest) GetDedicatedHost(ctx *pulumi.Context) (*ec2.DedicatedHost, error) {
	return nil, nil
}

func (r *SNCRequest) CustomIngressRules() []securityGroup.IngressRules {
	return nil
}

func (r *SNCRequest) CustomSecurityGroups(ctx *pulumi.Context) ([]*ec2.SecurityGroup, error) {
	return nil, nil
}

func (r *SNCRequest) GetPostScript(ctx *pulumi.Context) (string, error) {
	return "", nil
}

func (r *SNCRequest) Create(ctx *pulumi.Context,
	computeRequested compute.ComputeRequest) (*compute.Compute, error) {
	return r.RHELRequest.Create(ctx, r)
}

var cloudConfig string = `
#cloud-config  
rh_subscription:
  username: {{.SubscriptionUsername}}
  password: {{.SubscriptionPassword}}
  auto-attach: true
packages:
  - podman
  - "@virt"
  - jq
runcmd:
  - systemctl daemon-reload 
  - systemctl enable libvirtd-tcp.socket 
  - systemctl start --no-block libvirtd-tcp.socket 
  # Debug libvirt
  #- echo 'log_filters="1:libvirt 1:util 1:qemu"' | tee -a /etc/libvirt/libvirtd.conf
  #- echo 'log_outputs="1:file:/var/log/libvirt/libvirtd.log"' | tee -a /etc/libvirt/libvirtd.conf
  # https://libvirt.org/manpages/libvirtd.html#system-socket-activation
  - systemctl mask libvirtd.socket libvirtd-ro.socket libvirtd-admin.socket libvirtd-tls.socket libvirtd-tcp.socket
  - echo 'LIBVIRTD_ARGS="--listen"' | tee -a /etc/sysconfig/libvirtd
  - echo 'listen_tls = 0' | tee -a /etc/libvirt/libvirtd.conf
  - echo 'listen_tcp = 1' | tee -a /etc/libvirt/libvirtd.conf
  - echo 'tcp_port = "16509"' | tee -a /etc/libvirt/libvirtd.conf
  - echo 'auth_tcp = "none"' | tee -a /etc/libvirt/libvirtd.conf
  - systemctl enable libvirtd 
  - systemctl start --no-block libvirtd  
  - usermod -a -G libvirt ${username}
  - echo "user.max_user_namespaces=28633" | tee -a /etc/sysctl.d/userns.conf
  - sysctl -p /etc/sysctl.d/userns.conf
  - dnf upgrade -y curl openssl
  - dnf group install -y "Development Tools"
`

type userDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
}
