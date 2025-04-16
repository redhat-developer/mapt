package cirrus

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/util"
)

//go:embed snippet-darwin.sh
var snippetDarwin []byte

//go:embed snippet-linux.sh
var snippetLinux []byte

//go:embed snippet-windows.ps1
var snippetWindows []byte

var snippets map[Platform][]byte = map[Platform][]byte{
	Darwin:  snippetDarwin,
	Linux:   snippetLinux,
	Windows: snippetWindows}

var (
	version = "v0.135.0"
	baseURL = "https://github.com/cirruslabs/cirrus-cli/releases/download/%s/cirrus-%s-%s"
)

var pwa *PersistentWorkerArgs

func Init(args *PersistentWorkerArgs) {
	pwa = args
}

func (args *PersistentWorkerArgs) GetUserDataValues() *integrations.UserDataValues {
	if args == nil {
		return nil
	}
	return &integrations.UserDataValues{
		CliURL: downloadURL(),
		Name:   pwa.Name,
		Token:  pwa.Token,
		Labels: GetLabelsAsString(),
		Port:   cirrusPort,
	}
}

func (args *PersistentWorkerArgs) GetSetupScriptTemplate() string {
	templateConfig := string(snippets[*pwa.Platform][:])
	return templateConfig
}

func GetRunnerArgs() *PersistentWorkerArgs {
	return pwa
}

// If cirrus is enable it will return
// the port to be opened
func CirrusPort() (*int, error) {
	if pwa == nil {
		return nil, nil
	}
	port, err := strconv.Atoi(cirrusPort)
	return &port, err
}

// Get token
func GetToken() string {
	return util.IfNillable(pwa != nil,
		func() string { return pwa.Token },
		"")
}

// Get labels in format
func GetLabelsAsString() string {
	return util.IfNillable(pwa != nil,
		func() string {
			if len(pwa.Labels) > 0 {
				return strings.Join(func() []string {
					out := make([]string, 0, len(pwa.Labels))
					for k, v := range pwa.Labels {
						out = append(out, fmt.Sprintf("%s=%s", k, v))
					}
					return out
				}(), ",")
			}
			return ""
		},
		"")
}

// platform: darwin, linux, windows
// arch: amd64, arm64
func downloadURL() string {
	url := fmt.Sprintf(baseURL, version, *pwa.Platform, *pwa.Arch)
	if pwa.Platform == &Windows {
		url = fmt.Sprintf("%s.exe", url)
	}
	return url
}
