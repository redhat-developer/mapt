package supportmatrix

const (
	// Specific to Openshift Local
	olRHELID          string = "ol-rhel"
	olWindowsID       string = "ol-windows"
	olWindowsNonEngID string = "ol-windows-non-eng"
	olFedoraID        string = "ol-fedora"
	sSNCID            string = "s-snc"

	// General services
	sProxyID string = "s-proxy"

	OwnerSelf string = "self"
)

type SupportedType int

const (
	RHEL SupportedType = iota
	Windows
	MacM1
	Fedora
)
