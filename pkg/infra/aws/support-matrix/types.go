package supportmatrix

type SupportedHost struct {
	ID          string
	Description string
	// Filter for machine
	ProductDescription string
	InstaceTypes       []string
	// true if spot instances are supported
	Spot bool
	// If spot not true we can setup a specific Az
	FixedAMI *FixedAMI
	// AMI Pattern
	AMI AMI
	// RHEL, Windows, MacM1
	Type SupportedType
}

type AMI struct {
	RegexName string
	// In case wanna compose the regex pattern
	RegexPattern    string
	Filters         map[string]string
	DefaultUser     string
	Owner           string
	AMITargetName   string
	AMISourceID     string
	AMISourceRegion string
}

// When no replication is possible and only match
// one region / az
type FixedAMI struct {
	AvailavilityZone string
	Region           string
}
