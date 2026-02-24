package cloudinit

import (
	"fmt"
	"strings"
)

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
func IndentWriteFile(snippet *string) (*string, error) {
	lines := strings.Split(strings.TrimSpace(*snippet), "\n")
	for i, line := range lines {
		// Added 6 spaces before each line
		lines[i] = fmt.Sprintf("      %s", line)
	}
	identedSnippet := strings.Join(lines, "\n")
	return &identedSnippet, nil
}
