package compiler

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestStep14EnumAnnotationsPayloadAndNarrowing(t *testing.T) {
	ds := compileStep13(t, grammar.English, `create enum State { Ready, Loading as String }
let state as State as State.Loading("wait")
if state == State.Loading {
 let loading as State as state
} else {
 let ready as State as state
}
let invalid as State as State.Loading(1)`)
	if !hasCode(ds, "G3002") {
		t.Fatalf("expected payload mismatch: %#v", ds)
	}
}
func TestStep14InterfaceContractsAndExtensionErrors(t *testing.T) {
	ds := compileStep13(t, grammar.English, `create interface Base { readonly id as Number }
create interface User extends Base { name as String }
let user as User as {id: 1, name: "Ada"}
let bad as User as {id: 1}
create interface A extends B { a as String }
create interface B extends A { b as String }`)
	if !hasCode(ds, "G3002") || !hasCode(ds, "G3103") {
		t.Fatalf("want implementation and cycle diagnostics: %#v", ds)
	}
}
func TestStep14EnumMultilingual(t *testing.T) {
	for _, x := range []struct {
		l grammar.Language
		s string
	}{{grammar.Bengali, `তৈরি এনাম অবস্থা { প্রস্তুত } চলক x হিসেবে অবস্থা হিসেবে অবস্থা.প্রস্তুত`}, {grammar.Hindi, `बनाओ एनम स्थिति { तैयार } चर x रूप स्थिति रूप स्थिति.तैयार`}} {
		if ds := compileStep13(t, x.l, x.s); len(ds) != 0 {
			t.Fatalf("%s: %#v", x.l, ds)
		}
	}
}

func TestStep14DiagnosticJSONAndUTF8Positions(t *testing.T) {
	cases := []struct{ code, source string }{
		{"G3100", "create enum 状態 { A, A }"},
		{"G3101", "create interface A {} create interface A {}"},
		{"G3102", "create interface A extends Missing {}"},
		{"G3103", "create interface A extends B {} create interface B extends A {}"},
		{"G3104", "create interface A { x as String } create interface B { x as Number } create interface C extends A, B {}"},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			s := NewSession(WithGrammarLanguage(grammar.English))
			id := s.AddFile("utf8.gogo", "// বাংলা\n"+tc.source)
			s.ParseFile(id)
			var d diagnostics.Diagnostic
			found := false
			for _, candidate := range s.Diagnostics.All() {
				if candidate.Code == tc.code {
					d = candidate
					found = true
					break
				}
			}
			if !found || d.Severity != diagnostics.Error || d.Span.Start.Line < 2 || d.Span.Start.Column < 1 {
				t.Fatalf("bad diagnostic %#v", d)
			}
			data, err := (diagnostics.Renderer{Files: s.Files, Locale: diagnostics.Bengali}).JSON([]diagnostics.Diagnostic{d})
			if err != nil {
				t.Fatal(err)
			}
			var out []diagnostics.JSONDiagnostic
			if err := json.Unmarshal(data, &out); err != nil || len(out) != 1 {
				t.Fatalf("json=%s err=%v", data, err)
			}
			got := out[0]
			if got.Code != tc.code || got.Severity != "error" || got.Language != diagnostics.Bengali || got.File != "utf8.gogo" || got.Span.Start.Line != d.Span.Start.Line || got.Span.Start.Column != d.Span.Start.Column || got.Message == "" {
				t.Fatalf("bad JSON %#v", got)
			}
		})
	}
}

func TestStep14ConcurrentSessionsDoNotLeakDeclarations(t *testing.T) {
	cases := []struct {
		lang grammar.Language
		text string
	}{{grammar.English, `create enum State { Ready } create interface Card { state as State } let c as Card as {state: State.Ready}`}, {grammar.Bengali, `তৈরি এনাম অবস্থা { প্রস্তুত } তৈরি ইন্টারফেস কার্ড { মান হিসেবে অবস্থা } চলক c হিসেবে কার্ড হিসেবে {মান: অবস্থা.প্রস্তুত}`}, {grammar.Hindi, `बनाओ एनम स्थिति { तैयार } बनाओ इंटरफ़ेस कार्ड { मान रूप स्थिति } चर c रूप कार्ड रूप {मान: स्थिति.तैयार}`}}
	var wg sync.WaitGroup
	errs := make(chan []diagnostics.Diagnostic, len(cases))
	for _, tc := range cases {
		tc := tc
		wg.Add(1)
		go func() { defer wg.Done(); errs <- compileStep13(t, tc.lang, tc.text) }()
	}
	wg.Wait()
	close(errs)
	for ds := range errs {
		if len(ds) != 0 {
			t.Fatalf("leaked/failed session: %#v", ds)
		}
	}
}

func TestStep14EnumAndInterfaceControlFlowRegression(t *testing.T) {
	noStep13Diag(t, `create enum State { Ready, Loading as String }
create type StateAlias as State
create interface Base { kind as "user" }
create interface User extends Base { name as String }
create type Failure as Object{kind: "error", message: String}
create type Response as User | Failure
create variable state as StateAlias as State.Loading("wait")
if state == State.Loading {
 let loading as State as state
} else {
 let ready as State as state
}
state = State.Ready
let afterAssignment as State as state
let response as Response as {kind: "user", name: "Ada"}
if response.kind == "user" {
 let user as User as response
} else {
 let failure as Failure as response
}
fn display(input as State) as State {
 if input == State.Ready { return input }
 return input
}`)
}

func TestStep14InterfaceNegativeCompatibilityMatrix(t *testing.T) {
	ds := compileStep13(t, grammar.English, `create interface Child { name as String }
create interface Contract { child as Child, required as String, mutable as Number }
let missing as Contract as {mutable: 1, child: {name: "ok"}}
let wrong as Contract as {required: "ok", mutable: "bad", child: {name: "ok"}}
let nested as Contract as {required: "ok", mutable: 1, child: {name: 1}}
let optional as Object{required?: String, mutable: Number, child: Child} as {mutable: 1, child: {name: "ok"}}
let mismatch as Contract as optional
let readonlySource as Object{required: String, readonly mutable: Number, child: Child} as {required: "ok", mutable: 1, child: {name: "ok"}}
let readonlyMismatch as Contract as readonlySource`)
	if count := countCode(ds, "G3002"); count < 5 {
		t.Fatalf("want all contract incompatibilities, got %d: %#v", count, ds)
	}
}
