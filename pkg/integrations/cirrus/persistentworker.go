package cirrus

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
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

type snippetDataValues struct {
	CliURL string
	User   string
	Name   string
	Token  string
	Labels string
	Port   string
}

var (
	version = "v0.135.0"
	baseURL = "https://github.com/cirruslabs/cirrus-cli/releases/download/%s/cirrus-%s-%s"
)

var pwa *PersistentWorkerArgs

func Init(args *PersistentWorkerArgs) {
	pwa = args
}

func PersistentWorkerSnippet(username string) (*string, error) {
	if pwa == nil {
		noSnippet := ""
		return &noSnippet, nil
	}
	templateConfig := string(snippets[*pwa.Platform][:])
	snippet, err := file.Template(
		snippetDataValues{
			CliURL: downloadURL(),
			User:   username,
			Name:   pwa.Name,
			Token:  pwa.Token,
			Labels: getLabelsAsString(),
			Port:   cirrusPort,
		},
		templateConfig)
	return &snippet, err
}

// If we add the snippet as part of a cloud init file the strategy
// would be create the file with write_files:
// i.e.
// write_files:
//
//	# Cirrus service setup
//	- content: |
//	    {{ .CirrusSnippet }} <----- 6 spaces
//
// to do so we need to indent 6 spaces each line of the snippet
func PersistentWorkerSnippetAsCloudInitWritableFile(username string) (*string, error) {
	snippet, err := PersistentWorkerSnippet(username)
	if err != nil || len(*snippet) == 0 {
		return snippet, err
	}
	lines := strings.Split(strings.TrimSpace(*snippet), "\n")
	for i, line := range lines {
		// Added 6 spaces before each line
		lines[i] = fmt.Sprintf("      %s", line)
	}
	identedSnippet := strings.Join(lines, "\n")
	return &identedSnippet, nil

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

// platform: darwin, linux, windows
// arch: amd64, arm64
func downloadURL() string {
	url := fmt.Sprintf(baseURL, version, *pwa.Platform, *pwa.Arch)
	if pwa.Platform == &Windows {
		url = fmt.Sprintf("%s.exe", url)
	}
	return url
}

// Get labels in format
func getLabelsAsString() string {
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
