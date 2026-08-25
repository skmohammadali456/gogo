package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/token"
)

func TestSessionAcceptsMultilingualSource(t *testing.T) {
	s := NewSession()
	id := s.AddFile("main.gogo", "create variable user as \"মোহাম্মদ\"\ncreate variable city as \"दिल्ली\"")
	if id == 0 || s.HasErrors() {
		t.Fatalf("expected multilingual source to be accepted: id=%d diagnostics=%v", id, s.Diagnostics.All())
	}
}

func TestSessionRejectsEmptyPath(t *testing.T) {
	s := NewSession()
	if id := s.AddFile("", "hello"); id != 0 {
		t.Fatalf("expected rejected file ID 0, got %d", id)
	}
	if !s.HasErrors() {
		t.Fatal("expected diagnostic for empty path")
	}
}

func TestSessionRejectsInvalidUTF8WithoutMutatingExistingFile(t *testing.T) {
	s := NewSession()
	id := s.AddFile("main.gogo", "create variable value as 1")
	if id == 0 {
		t.Fatal("expected initial file to be accepted")
	}

	bad := string([]byte{'x', 0xff})
	if got := s.AddFile("main.gogo", bad); got != 0 {
		t.Fatalf("expected invalid replacement to be rejected, got %d", got)
	}
	file, ok := s.Files.Get(id)
	if !ok || file.Text != "create variable value as 1" {
		t.Fatalf("invalid replacement mutated existing file: %+v", file)
	}
}

func TestSessionLexFileIntegratesLexerDiagnostics(t *testing.T) {
	s := NewSession()
	id := s.AddFile("main.gogo", "create variable value as 10.")
	if id == 0 {
		t.Fatal("expected source file to be added")
	}

	tokens := s.LexFile(id)
	if len(tokens) != 6 {
		t.Fatalf("got %d tokens, want 6: %#v", len(tokens), tokens)
	}
	if tokens[0].Kind != token.Identifier || tokens[len(tokens)-1].Kind != token.EOF {
		t.Fatalf("unexpected lexer output: %#v", tokens)
	}
	if !s.HasErrors() {
		t.Fatal("expected lexer diagnostic to be copied into session")
	}
	if got := s.Diagnostics.All()[0].Code; got != "G1002" {
		t.Fatalf("expected G1002, got %s", got)
	}
}

func TestSessionLexMissingFile(t *testing.T) {
	s := NewSession()
	if tokens := s.LexFile(999); tokens != nil {
		t.Fatalf("expected nil tokens for missing file, got %#v", tokens)
	}
	if len(s.Diagnostics.All()) != 1 || s.Diagnostics.All()[0].Code != "G0003" {
		t.Fatalf("expected G0003, got %v", s.Diagnostics.All())
	}
}
