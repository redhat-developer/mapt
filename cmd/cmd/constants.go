package cmd

const (
	projectName                 string = "project-name"
	projectNameDesc             string = "project name to identify the instance of the stack"
	backedURL                   string = "backed-url"
	backedURLDesc               string = "backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket"
	connectionDetailsOutput     string = "conn-details-output"
	connectionDetailsOutputDesc string = "path to export host connection information. Here 3 files will be created holding: host, username and privateKey"
	supportedHostID             string = "host-id"
	supportedHostIDDesc         string = "host id from supported hosts list"
	availabilityZones           string = "availability-zones"
	availabilityZonesDesc       string = "List of comma separated azs to check. If empty all will be searched"
	rhMajorVersion              string = "rh-major-version"
	rhMajorVersionDesc          string = "major version for rhel image 7, 8 or 9"
	rhSubcriptionUsername       string = "rh-subscription-username"
	rhSubcriptionUsernameDesc   string = "username for rhel subcription"
	rhSubcriptionPassword       string = "rh-subscription-password"
	rhSubcriptionPasswordDesc   string = "password for rhel subcription"

	createCmdName  string = "create"
	destroyCmdName string = "destroy"
)
