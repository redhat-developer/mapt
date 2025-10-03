package userdata

type CloudConfig interface {
	CloudConfig() (*string, error)
}
