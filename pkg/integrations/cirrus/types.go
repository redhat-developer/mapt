package cirrus

type Platform string
type Arch string

var (
	cirrusPort = "3010"

	Windows Platform = "windows"
	Linux   Platform = "linux"
	Darwin  Platform = "darwin"

	Arm64 Arch = "arm64"
	Amd64 Arch = "amd64"
)

type PersistentWorkerArgs struct {
	Name     string
	Token    string
	Platform *Platform
	Arch     *Arch
	Labels   map[string]string
}
