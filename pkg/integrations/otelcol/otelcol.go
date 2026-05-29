package otelcol

import (
	_ "embed"

	cloudinit "github.com/redhat-developer/mapt/pkg/util/cloud-init"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

// version is overridden at build time via -ldflags.
var version = "0.151.0"

//go:embed snippet-linux.sh
var snippetLinux []byte

// GetSnippet renders the install script template with the provided args and
// returns the shell script as a string. Returns an empty string when args is nil.
func GetSnippet(args *OtelcolArgs) (*string, error) {
	if args == nil {
		empty := ""
		return &empty, nil
	}
	if args.ColVersion == "" {
		args.ColVersion = version
	}
	snippet, err := file.Template(args, string(snippetLinux))
	return &snippet, err
}

// GetSnippetAsCloudInitWritableFile returns the rendered install script indented
// with 6 spaces so it can be embedded directly as the content of a cloud-init
// write_files entry.
func GetSnippetAsCloudInitWritableFile(args *OtelcolArgs) (*string, error) {
	snippet, err := GetSnippet(args)
	if err != nil || len(*snippet) == 0 {
		return snippet, err
	}
	return cloudinit.IndentWriteFile(snippet)
}
