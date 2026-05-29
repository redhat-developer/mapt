package otelcol

// Arch is a Linux architecture identifier used for binary download URLs.
type Arch string

var (
	Ppc64le Arch = "ppc64le"
	S390x   Arch = "s390x"
	Amd64   Arch = "amd64"
	Arm64   Arch = "arm64"
)

// OtelcolArgs holds all parameters needed to install and configure the
// otelcol-contrib filelog collector on a Linux host.
type OtelcolArgs struct {
	AppCode    string
	AuthToken  string
	Index      string
	Endpoint   string
	ColVersion string            // overridden from module var if empty
	Arch       Arch              // target linux arch (ppc64le, s390x, amd64, arm64)
	SyslogPath string            // distro-specific syslog path
	SecurePath string            // distro-specific auth/secure log path
	ExtraAttrs map[string]string // additional resource attributes

	// MonitorGitLabRunner adds a filelog/gitlab-runner receiver that tails
	// /var/log/gitlab-runner/runner.log when set to true.
	MonitorGitLabRunner bool
}
