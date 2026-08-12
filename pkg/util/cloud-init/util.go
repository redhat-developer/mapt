package cloudinit

import (
	"fmt"
	"strings"
)

// IndentWriteFile indents each line of snippet by 6 spaces so it can be
// embedded inside a cloud-init write_files content block.
func IndentWriteFile(snippet *string) (*string, error) {
	lines := strings.Split(strings.TrimSpace(*snippet), "\n")
	for i, line := range lines {
		// Added 6 spaces before each line
		lines[i] = fmt.Sprintf("      %s", line)
	}
	identedSnippet := strings.Join(lines, "\n")
	return &identedSnippet, nil
}
