package compiler

import (
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/source"
)

// Session contains state shared by compiler phases for one compilation request.
type Session struct {
	Files       *source.FileMap
	Diagnostics diagnostics.Bag
}

func NewSession() *Session {
	return &Session{Files: source.NewFileMap()}
}

func (s *Session) AddFile(path, text string) uint32 {
	if !source.ValidatePath(path) {
		s.Diagnostics.Add(diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Code:     "G0001",
			Message:  "I need a source file path before I can compile this file.",
			Hint:     "Give the source file a non-empty path, such as main.gogo.",
		})
		return 0
	}

	id := s.Files.Add(path, text)
	file, ok := s.Files.Get(id)
	if !ok || !source.ValidateFile(file) {
		s.Diagnostics.Add(diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Code:     "G0002",
			Message:  "This source file is not valid UTF-8 or has invalid source metadata.",
			Hint:     "Save the file as UTF-8 and make sure its path is valid.",
		})
		return 0
	}
	return id
}

func (s *Session) HasErrors() bool {
	return s.Diagnostics.HasErrors()
}
