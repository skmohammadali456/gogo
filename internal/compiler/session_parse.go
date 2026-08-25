package compiler

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/parser"
)

// ParseFile lexes and parses a registered source file into a GOGO AST.
func (s *Session) ParseFile(id uint32) (ast.File, bool) {
	file, ok := s.Files.Get(id)
	if !ok {
		return ast.File{}, false
	}
	p := parser.New((&lexerBridge{file: file}).Lex())
	for _, d := range p.Diagnostics() { s.Diagnostics.Add(d) }
	return p.ParseFile(), true
}

// lexerBridge keeps compiler session integration small while the compiler pipeline is assembled.
type lexerBridge struct{ file sourceFile }
type sourceFile interface{}

func (b *lexerBridge) Lex() []token.Token { return nil }
