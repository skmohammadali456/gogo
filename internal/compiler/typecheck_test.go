package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestStep10TypedDeclarationsAndSignatures(t *testing.T) {
	s := NewSession(WithGrammarLanguage(grammar.English))
	id := s.AddFile("types.gogo", `create variable names as Array<String> as ["Ada"]
create variable lookup as Map<String, Number> as lookup
fn format(value as Tuple<String, Number>) as Record{name: String, score: Number} { return value }`)
	_, ok := s.ParseFile(id)
	if !ok || s.HasErrors() {
		t.Fatalf("unexpected diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep10InvalidTypeAndAssignmentDiagnostics(t *testing.T) {
	s := NewSession()
	id := s.AddFile("bad.gogo", `create variable value as Boolean as "no"
create variable unknown as Mystery as 1`)
	s.ParseFile(id)
	codes := map[string]bool{}
	for _, d := range s.Diagnostics.All() {
		codes[d.Code] = true
	}
	if !codes["G3001"] || !codes["G3002"] {
		t.Fatalf("want G3001 and G3002, got %#v", s.Diagnostics.All())
	}
}

func TestStep10MultilingualTypesUseSameCanonicalModel(t *testing.T) {
	for _, language := range []grammar.Language{grammar.English, grammar.Bengali, grammar.Hindi} {
		s := NewSession(WithGrammarLanguage(language))
		text := `create variable x as Array<String> as ["a"]`
		if language == grammar.Bengali {
			text = `তৈরি চলক নাম হিসেবে Array<String> হিসেবে ["a"]`
		}
		if language == grammar.Hindi {
			text = `बनाओ चर नाम रूप Array<String> रूप ["a"]`
		}
		id := s.AddFile("multi.gogo", text)
		s.ParseFile(id)
		if s.HasErrors() {
			t.Fatalf("%s diagnostics: %#v", language, s.Diagnostics.All())
		}
	}
}

func TestStep10BindingMutabilityAndReassignment(t *testing.T) {
	s := NewSession()
	id := s.AddFile("mutability.gogo", `const locked as Number as 1
locked = 2
let open as Number as 1
open = "wrong"`)
	s.ParseFile(id)
	codes := map[string]bool{}
	for _, d := range s.Diagnostics.All() {
		codes[d.Code] = true
	}
	if !codes["G3002"] || !codes["G3003"] {
		t.Fatalf("expected type and immutable-binding diagnostics: %#v", s.Diagnostics.All())
	}
}
