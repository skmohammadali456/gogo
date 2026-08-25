package source

import (
	"strings"
	"unicode/utf8"
)

// ValidatePath checks the minimum invariants required for a source path.
// Path normalization and project-root policy belong to the project layer.
func ValidatePath(path string) bool {
	return strings.TrimSpace(path) != "" && !strings.ContainsRune(path, '\x00')
}

// ValidateFile checks source metadata and source encoding invariants.
func ValidateFile(file File) bool {
	if file.ID == 0 || !ValidatePath(file.Path) {
		return false
	}
	return utf8.ValidString(file.Text)
}
