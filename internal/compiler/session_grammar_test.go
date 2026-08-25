package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestSessionGrammarVocabulary(t *testing.T) {
	s := NewSession(WithGrammarLanguage(grammar.Hindi))
	id := s.AddFile("main.gogo", "बनाओ चर user के_रूप_में 1")
	parsed, ok := s.ParseFile(id)
	if !ok || s.HasErrors() {
		t.Fatalf("parse failed: %#v", s.Diagnostics.All())
	}
	if parsed.Statements[0].(ast.VariableDecl).Name.Name != "user" {
		t.Fatalf("unexpected AST: %#v", parsed.Statements[0])
	}
}

func TestMultipleSessionsUseIndependentVocabularies(t *testing.T) {
	en := NewSession()
	bn := NewSession(WithGrammarLanguage(grammar.Bengali))
	enID := en.AddFile("en.gogo", "create variable user as 1")
	bnID := bn.AddFile("bn.gogo", "তৈরি চলক user হিসেবে 1")
	if _, ok := en.ParseFile(enID); !ok || en.HasErrors() {
		t.Fatalf("English session failed: %#v", en.Diagnostics.All())
	}
	if _, ok := bn.ParseFile(bnID); !ok || bn.HasErrors() {
		t.Fatalf("Bengali session failed: %#v", bn.Diagnostics.All())
	}
	badID := en.AddFile("bad.gogo", "তৈরি চলক user হিসেবে 1")
	en.ParseFile(badID)
	if !en.HasErrors() {
		t.Fatal("English session should reject Bengali keywords without changing Bengali session")
	}
	if bn.HasErrors() {
		t.Fatalf("Bengali session was affected: %#v", bn.Diagnostics.All())
	}
}
