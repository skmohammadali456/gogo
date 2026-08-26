package compiler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestStep15AuditWrongArgCountsAndUnsafeInferenceFail(t *testing.T) {
	s := step15(`
create function id<T>(x as T) as T { return x }
create function same<T>(a as T, b as T) as T { return a }
create variable tooFew as Number as id()
create variable tooMany as Number as id(1, 2)
create variable conflict as Number as same(1, "x")
`, grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3204") || !hasDiag15(s, "G3205") {
		t.Fatalf("want arity and unsafe inference diagnostics: %#v", s.Diagnostics.All())
	}
}

func TestStep15AuditGenericInterfaceAliasAndRecursiveTermination(t *testing.T) {
	s := step15(`
create interface Box<T> { value as T }
create type Loop<T> as Loop<T>
create variable ok as Box<String> as {value:"x"}
create variable bad as Loop<Number> as 1
`, grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3001") {
		t.Fatalf("want recursive generic resolution diagnostic: %#v", s.Diagnostics.All())
	}
}

func TestStep15AuditGenericNarrowingBranchAndBoundary(t *testing.T) {
	s := step15(`
create function branch<T>(flag as Boolean, a as T, b as T) as T {
 if flag { return a } else { return b }
}
create variable value as Number as branch(true, 1, 2)
`, grammar.DefaultVocabulary())
	if s.Diagnostics.HasErrors() {
		t.Fatalf("generic optional narrowing failed: %#v", s.Diagnostics.All())
	}
}

func TestStep15AuditMalformedSyntaxLocalizedJSONPositions(t *testing.T) {
	s := step15("create function bad<টি, টি(x as টি) as টি { return x }", grammar.DefaultVocabulary())
	if !hasDiag15(s, "G3202") {
		t.Fatalf("want malformed generic diagnostic: %#v", s.Diagnostics.All())
	}
	out := diagnostics.Renderer{Files: s.Files, Locale: diagnostics.Hindi}.Text(s.Diagnostics.All())
	if !strings.Contains(out, "generic") && !strings.Contains(out, "सूची") {
		t.Fatalf("localized output missing generic content: %s", out)
	}
	data, err := diagnostics.Renderer{Files: s.Files, Locale: diagnostics.Hindi}.JSON(s.Diagnostics.All())
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil || len(decoded) == 0 {
		t.Fatalf("json decode: %v %s", err, data)
	}
}
