package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
)

func TestSessionParseFile(t *testing.T) {
	s := NewSession()
	id := s.AddFile("main.gogo", `create variable user as "Alex"`)
	file, ok := s.ParseFile(id)
	if !ok || s.HasErrors() {
		t.Fatalf("parse failed: ok=%v diagnostics=%v", ok, s.Diagnostics.All())
	}
	if len(file.Statements) != 1 {
		t.Fatalf("got %d statements", len(file.Statements))
	}
	if _, ok := file.Statements[0].(ast.VariableDecl); !ok {
		t.Fatalf("unexpected statement: %#v", file.Statements[0])
	}
}

func TestSessionParseMissingFile(t *testing.T) {
	s := NewSession()
	_, ok := s.ParseFile(999)
	if ok || !s.HasErrors() {
		t.Fatal("expected missing-file diagnostic")
	}
}
