package compiler

import (
	"github.com/skmohammadali786/gogo/internal/grammar"
	"strings"
	"testing"
)

func step11File(t *testing.T, text string, v grammar.Vocabulary) *Session {
	t.Helper()
	s := NewSession(WithGrammarVocabulary(v))
	id := s.AddFile("step11.gogo", text)
	s.ParseFile(id)
	return s
}
func TestStep11AliasesResolveCanonically(t *testing.T) {
	s := step11File(t, "create type User as Object{readonly name: String, tags?: Array<String>}\ncreate type Users as Array<User>\ncreate variable users as Users as []", grammar.DefaultVocabulary())
	if len(s.Diagnostics.All()) != 0 {
		t.Fatal(s.Diagnostics.All())
	}
}
func TestStep11AliasDiagnostics(t *testing.T) {
	for _, text := range []string{"create type A as Missing\ncreate variable x as A as 1", "create type A as B\ncreate type B as A\ncreate variable x as A as 1", "create type A as String\ncreate type A as Number"} {
		s := step11File(t, text, grammar.DefaultVocabulary())
		if len(s.Diagnostics.All()) == 0 {
			t.Fatalf("expected alias diagnostic: %s", text)
		}
		if !strings.Contains(s.Diagnostics.All()[0].Code, "G300") {
			t.Fatal(s.Diagnostics.All())
		}
	}
}
func TestStep11MultilingualAliasesShareIdentity(t *testing.T) {
	cases := []struct {
		text string
		lang grammar.Language
	}{{"create type User as Object{name: String}", grammar.English}, {"তৈরি ধরন User হিসেবে Object{name: String}", grammar.Bengali}, {"बनाओ प्रकार User रूप Object{name: String}", grammar.Hindi}}
	for _, c := range cases {
		v := grammar.Must(c.lang)
		s := step11File(t, c.text, v)
		if len(s.Diagnostics.All()) != 0 {
			t.Fatalf("%s: %v", c.lang, s.Diagnostics.All())
		}
	}
}
