package supportmatrix

type SupportedHost struct {
	ID          string
	Description string
	// Filter for machine
	ProductDescription string
	InstaceTypes       []string
	// true if spot instances are supported
	Spot bool
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
