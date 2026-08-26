package compiler

import (
	"strings"
	"sync"
	"testing"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func compileStep13(t *testing.T, lang grammar.Language, text string) []diagnostics.Diagnostic {
	t.Helper()
	s := NewSession(WithGrammarLanguage(lang))
	id := s.AddFile("step13.gogo", text)
	s.ParseFile(id)
	return s.Diagnostics.All()
}
func hasCode(ds []diagnostics.Diagnostic, code string) bool {
	for _, d := range ds {
		if d.Code == code {
			return true
		}
	}
	return false
}
func noStep13Diag(t *testing.T, text string) {
	t.Helper()
	if ds := compileStep13(t, grammar.English, text); len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", ds)
	}
}

func TestStep13InferenceLiteralsPrimitivesCollectionsObjectsAndContext(t *testing.T) {
	noStep13Diag(t, `let s as "Ada"
let n as 1
let b as true
let xs as Array<String | Number> as ["Ada", 1]
let obj as Object{name: String, score: Number} as {name: "Ada", score: 1}`)
}
func TestStep13ExplicitAnnotationCompatibilityAndInvalidInference(t *testing.T) {
	ds := compileStep13(t, grammar.English, `let bad as Number as "no"
let unknown as missing + 1`)
	if !hasCode(ds, "G3002") || !hasCode(ds, "G3007") {
		t.Fatalf("want G3002 and G3007: %#v", ds)
	}
}
func TestStep13OptionalUnionTruthinessNarrowingBranchesAndJoins(t *testing.T) {
	noStep13Diag(t, `let value as Optional<String> as "Ada"
if value { let inner as String as value } else { let absent as Optional<String> as value }
let again as Optional<String> as value
let mixed as String | Number as "Ada"
if mixed == "Ada" { let narrowed as String as mixed } else { let sibling as String | Number as mixed }
let joined as String | Number as mixed`)
}
func TestStep13DiscriminatedPropertyUnionAliasesNestedAliases(t *testing.T) {
	noStep13Diag(t, `create type A as Object{kind: "user", user: Object{name: String}}
create type B as Object{kind: "error", error: String}
create type State as A | B
create type Alias as State
let state as Alias as {kind: "user", user: {name: "Ada"}}
if state.kind == "user" { let user as A as state; let name as String as state.user.name } else { let err as B as state }`)
}
func TestStep13ResultOkErrNarrowingAndEarlyReturn(t *testing.T) {
	noStep13Diag(t, `create type User as Object{name: String}
let load as Result<User, String> as {name: "Ada"}
if load { let user as User as load } else { let message as String as load }
fn show(input as Optional<String>) as String { if input { return input } return "fallback" }`)
}
func TestStep13InvalidNarrowingAndTruthinessDiagnostics(t *testing.T) {
	ds := compileStep13(t, grammar.English, `let n as Number as 1
if n { let x as Number as n }
let v as String | Number as 1
if v == "x" { let bad as Boolean as v }`)
	if !hasCode(ds, "G3008") || !hasCode(ds, "G3002") {
		t.Fatalf("want truthiness and assignability diagnostics: %#v", ds)
	}
}
func TestStep13IntersectionBehaviorAndPropertyChecks(t *testing.T) {
	noStep13Diag(t, `create type HasName as Object{name: String}
create type HasAge as Object{age: Number}
let both as HasName & HasAge as {name: "Ada", age: 1}
let name as String as both.name`)
}
func TestStep13EnglishBengaliHindiUnicodeSource(t *testing.T) {
	cases := []struct {
		lang grammar.Language
		text string
	}{
		{grammar.English, `let নাম as Optional<String> as "আদা"
if নাম { let inner as String as নাম }`},
		{grammar.Bengali, `চলক नाम হিসেবে Optional<String> হিসেবে "आशा"
যদি नाम { চলক inner হিসেবে String হিসেবে name }`},
		{grammar.Hindi, `चर নাম रूप Optional<String> रूप "আশা"
अगर নাম { चर inner रूप String रूप নাম }`},
	}
	for _, tc := range cases {
		if ds := compileStep13(t, tc.lang, tc.text); len(ds) != 0 {
			t.Fatalf("%s: %#v", tc.lang, ds)
		}
	}
}
func TestStep13ConcurrentCompilerSessions(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ds := compileStep13(t, grammar.English, `let x as Optional<String> as "ok"
if x { let y as String as x }`)
			if len(ds) != 0 {
				t.Errorf("%#v", ds)
			}
		}()
	}
	wg.Wait()
}
func TestStep13MalformedSyntaxRecoveryAndLocalizedDiagnostics(t *testing.T) {
	ds := compileStep13(t, grammar.Hindi, `चर x रूप Optional<String रूप "oops"
अगर x { चर y रूप String रूप x `)
	if len(ds) == 0 {
		t.Fatal("expected diagnostics")
	}
	r := diagnostics.Renderer{Locale: diagnostics.Hindi}
	if !strings.Contains(r.Text(ds), "इस") && len(ds) > 0 {
		t.Fatalf("expected localized rendering: %s", r.Text(ds))
	}
}

func TestStep13ExhaustivenessNonExhaustiveCases(t *testing.T) {
	ds := compileStep13(t, grammar.English, `create type A as Object{kind: "a", a: String}
create type B as Object{kind: "b", b: Number}
let state as A | B as {kind: "a", a: "x"}
if state.kind == "a" { let a as A as state }`)
	if !hasCode(ds, "G3009") {
		t.Fatalf("want non-exhaustive diagnostic: %#v", ds)
	}
}
