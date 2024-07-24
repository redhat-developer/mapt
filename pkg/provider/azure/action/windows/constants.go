package windows

const (
	stackCreateWindowsDesktop = "stackCreateWindowsDesktop"
	stackSyncWindowsDesktop   = "stackSyncWindowsDesktop"

	azureWindowsDesktopID = "awd"

	scriptName = "setup.ps1"

	outputHost              = "awdHost"
	outputUsername          = "awdUsername"
	outputUserPassword      = "awdUserPassword"
	outputUserPrivateKey    = "awdUserPrivatekey"
	outputAdminUsername     = "awdAdminUsername"
	outputAdminUserPassword = "awdAdminUserPassword"

	ProfileCRC = "crc"
)

var profiles = []string{ProfileCRC}
