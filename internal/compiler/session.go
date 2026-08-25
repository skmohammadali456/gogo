package compiler

import (
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
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

// LexFile runs the Step 2 lexer for a file already registered in the session.
// Lexer diagnostics are copied into the session so later compiler phases see
// one diagnostic stream while every token retains its original source span.
func (s *Session) LexFile(id uint32) []token.Token {
	file, ok := s.Files.Get(id)
	if !ok {
		s.Diagnostics.Add(diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Code:     "G0003",
			Message:  "The compiler cannot lex a source file that is not in this session.",
			Hint:     "Add the file to the compilation session before lexing it.",
		})
		return nil
	}

	l := lexer.New(file)
	tokens := l.LexAll()
	for _, diagnostic := range l.Diagnostics() {
		s.Diagnostics.Add(diagnostic)
	}
	return tokens
}

func (s *Session) HasErrors() bool {
	return s.Diagnostics.HasErrors()
}
