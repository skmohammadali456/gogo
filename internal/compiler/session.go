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
	return s.Files.Add(path, text)
}

func (s *Session) HasErrors() bool {
	return s.Diagnostics.HasErrors()
}
