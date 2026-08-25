package compiler

import "testing"

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
