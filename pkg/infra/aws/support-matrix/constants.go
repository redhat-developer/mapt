package supportmatrix

const (
	// Specific to Openshift Local
	olRHELID    string = "ol-rhel"
	olWindowsID string = "ol-windows"

	// General mac m1
	gMacOSM1ID string = "g-macos-m1"

	// General services
	sBastionID string = "s-bastion"
	sProxyID   string = "s-proxy"
)

type SupportedType int

const (
	RHEL SupportedType = iota
	Windows
	MacM1
)
