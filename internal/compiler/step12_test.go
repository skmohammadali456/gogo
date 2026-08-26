package compiler

import (
	"strings"
	"sync"
	"testing"

	"github.com/skmohammadali786/gogo/internal/grammar"
)

func step12Session(text string, v grammar.Vocabulary) *Session {
	s := NewSession(WithGrammarVocabulary(v))
	id := s.AddFile("step12.gogo", text)
	s.ParseFile(id)
	return s
}

func TestStep12AliasesAndAssignments(t *testing.T) {
	text := `create type Name as Optional<String>
create type ID as String | Number | String
create type User as Object{name: String, tag?: Optional<String>}
create type Load as Result<User, String | Object{message: String}>
create variable a as ID as "abc"
create variable u as User as {name: "Ada"}
create variable r as User as {name: "Ada"}`
	s := step12Session(text, grammar.DefaultVocabulary())
	if s.HasErrors() {
		t.Fatalf("unexpected diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep12InvalidAssignmentsAndCyclicAliases(t *testing.T) {
	text := `create type A as B
create type B as A
create variable bad as Number as "not-number"
create variable no as Optional<Number> as "not-number"
create variable wrong as A as 1`
	s := step12Session(text, grammar.DefaultVocabulary())
	if len(s.Diagnostics.All()) == 0 {
		t.Fatalf("expected diagnostics")
	}
	codes := map[string]bool{}
	for _, d := range s.Diagnostics.All() {
		codes[d.Code] = true
	}
	if !codes["G3001"] || !codes["G3002"] {
		t.Fatalf("want type and assignment diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep12MultilingualSyntaxCanonical(t *testing.T) {
	cases := []struct {
		lang grammar.Language
		text string
	}{
		{grammar.English, `create type State as Result<Object{নাম: String}, String | Number>`},
		{grammar.Bengali, `তৈরি ধরন State হিসেবে Result<Object{নাম: String}, String | Number>`},
		{grammar.Hindi, `बनाओ प्रकार State रूप Result<Object{নাম: String}, String | Number>`},
	}
	for _, c := range cases {
		s := step12Session(c.text, grammar.Must(c.lang))
		if s.HasErrors() {
			t.Fatalf("%s: %#v", c.lang, s.Diagnostics.All())
		}
	}
}

func TestStep12ConcurrentCompilerSessionsIsolated(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "A"
			if i%2 == 0 {
				name = "B"
			}
			s := step12Session("create type "+name+" as String | Number\ncreate variable x as "+name+" as \"ok\"", grammar.DefaultVocabulary())
			if s.HasErrors() {
				t.Errorf("unexpected: %#v", s.Diagnostics.All())
			}
		}(i)
	}
	wg.Wait()
}

func TestStep12MalformedSyntaxRecovers(t *testing.T) {
	for _, text := range []string{`create type A as String |`, `create type A as Result<String>
create variable x as A as 1`, `create type A as Optional<>
create variable x as A as 1`, `create type A as String & & Number`} {
		s := step12Session(text, grammar.DefaultVocabulary())
		if len(s.Diagnostics.All()) == 0 {
			t.Fatalf("expected diagnostic for %q", text)
		}
		if strings.Contains(strings.ToLower(s.Diagnostics.All()[0].Message), "panic") {
			t.Fatalf("bad diagnostic: %#v", s.Diagnostics.All())
		}
	}
}
