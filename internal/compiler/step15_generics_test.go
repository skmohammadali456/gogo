package compiler

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func step15(text string, v grammar.Vocabulary) *Session {
	s := NewSession(WithGrammarVocabulary(v))
	id := s.AddFile("step15.gogo", text)
	f, _ := s.ParseFile(id)
	s.checkTypes(f)
	return s
}
func hasDiag15(s *Session, code string) bool {
	for _, d := range s.Diagnostics.All() {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestStep15GenericFunctionExplicitInferredContextualAndNested(t *testing.T) {
	s := step15(`
create function identity<T>(x as T) as T { return x }
create function choose<A,B>(a as A, b as B) as A { return a }
create variable a as Number as identity(1)
create variable b as String as identity<String>("x")
create variable c as Number as choose(identity(1), identity("x"))
`, grammar.DefaultVocabulary())
	if s.Diagnostics.HasErrors() {
		t.Fatalf("diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep15GenericConstraintsAndFailures(t *testing.T) {
	s := step15(`
create function needObj<T extends Object{name: String}>(x as T) as String { return x.name }
create variable good as String as needObj({name:"ok", age:1})
create variable bad as String as needObj<Number>(1)
create function leak<T>(x as T) as T { return x }
create variable escaped as T as 1
`, grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3206") || !hasDiag15(s, "G3001") {
		t.Fatalf("want constraint and unresolved parameter diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep15GenericAliasesNestedCollectionsAndArity(t *testing.T) {
	s := step15(`
create type Box<T> as Object{value: T}
create type Many<T> as Array<Optional<Result<T, String>>>
create variable ok as Box<Number> as {value:1}
create variable wrong as Box as {value:1}
`, grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3001") {
		t.Fatalf("want arity diagnostic: %#v", s.Diagnostics.All())
	}
}

func TestStep15ParserGenericComponentsUnicodeAndLocales(t *testing.T) {
	cases := []struct {
		v   grammar.Vocabulary
		src string
	}{
		{grammar.DefaultVocabulary(), `create component Card<T>(value as "x") { }`},
		{grammar.Must(grammar.Bengali), `তৈরি কম্পোনেন্ট কার্ড<টি>(মান হিসেবে "x") { }`},
		{grammar.Must(grammar.Hindi), `बनाओ घटक कार्ड<टी>(मान रूप "x") { }`},
	}
	for _, tc := range cases {
		s := step15(tc.src, tc.v)
		if s.Diagnostics.HasErrors() {
			t.Fatalf("%s diagnostics: %#v", tc.v.Language, s.Diagnostics.All())
		}
	}
}

func TestStep15GenericDiagnosticsJSONLocalizedAndConcurrent(t *testing.T) {
	s := step15(`create function id<T>(x as T) as T { return x }
create variable bad as Number as id<String>("x")`, grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3002") {
		t.Fatalf("want assignability diagnostic: %#v", s.Diagnostics.All())
	}
	b, err := json.Marshal(s.Diagnostics.All())
	if err != nil || !strings.Contains(string(b), "G3002") {
		t.Fatalf("json: %s %v", b, err)
	}
	_ = diagnostics.Renderer{Files: s.Files, Locale: diagnostics.Bengali}.Text(s.Diagnostics.All())
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss := step15(`create function id<T>(x as T) as T { return x }
create variable a as Number as id(1)`, grammar.DefaultVocabulary())
			if ss.Diagnostics.HasErrors() {
				t.Errorf("session leaked: %#v", ss.Diagnostics.All())
			}
		}()
	}
	wg.Wait()
}
