package cmd

const (
	projectName           string = "project-name"
	projectNameDesc       string = "project name to identify the instance of the stack"
	backedURL             string = "backed-url"
	backedURLDesc         string = "backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket"
	supportedHostID       string = "host-id"
	supportedHostIDDesc   string = "host id from supported hosts list"
	availabilityZones     string = "availability-zones"
	availabilityZonesDesc string = "List of comma separated azs to check. If empty all will be searched"

	createCmdName  string = "create"
	destroyCmdName string = "destroy"
)
