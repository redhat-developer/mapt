package supportmatrix

type SupportedHost struct {
	ID          string
	Description string
	// Filter for machine
	ProductDescription string
	// Manually check the map on https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#SpotPlacementScore:
	//between instances types and Requirements
	InstaceTypes []string
	// Requirements Requirements
	// true if spot instances are supported
	Spot bool
	// AMI Pattern
	AMI AMI
}

type AMI struct {
	RegexName string
	// In case wanna compose the regex pattern
	RegexPattern string
	Filters      map[string]string
	DefaultUser  string
}
