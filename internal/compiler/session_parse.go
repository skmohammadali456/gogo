package compiler

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/parser"
)

// ParseFile runs the lexer and parser for a registered source file and merges diagnostics into the session.
func (s *Session) ParseFile(id uint32) (ast.File, bool) {
	file, ok := s.Files.Get(id)
	if !ok {
		s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G0004", Message: "The compiler cannot parse a source file that is not in this session.", Hint: "Add the file to the compilation session before parsing it."})
		return ast.File{}, false
	}
	l := lexer.New(file)
	p := parser.New(l.LexAll())
	for _, d := range l.Diagnostics() {
		s.Diagnostics.Add(d)
	}
	result := p.ParseFile()
	for _, d := range p.Diagnostics() {
		s.Diagnostics.Add(d)
	}
	return result, true
}
