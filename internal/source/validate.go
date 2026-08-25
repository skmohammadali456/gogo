package source

import (
	"unicode/utf8"
)

// ValidateUTF8 reports whether source text is valid UTF-8.
func ValidateUTF8(text string) bool {
	return utf8.ValidString(text)
}

// ValidatePath performs the compiler's minimum source-file identity checks.
// Empty paths are rejected because diagnostics and future module resolution need
// a stable human-readable source identity.
func ValidatePath(path string) bool {
	return path != ""
}

// ValidateFile checks invariants required by compiler phases.
func ValidateFile(file File) bool {
	return file.ID != 0 && ValidatePath(file.Path) && ValidateUTF8(file.Text)
}
