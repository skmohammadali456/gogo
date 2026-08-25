package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

func parseStep8(t *testing.T, lang grammar.Language, text string) (ast.File, []diagnostics.Diagnostic) {
	t.Helper()
	lx := lexer.New(source.File{ID: 1, Path: "step8.gogo", Text: text})
	p := New(lx.LexAll(), WithVocabulary(grammar.Must(lang)))
	file := p.ParseFile()
	return file, p.Diagnostics()
}

func TestStep8HindiDeclarationsAndUnicodeIdentifiers(t *testing.T) {
	file, diags := parseStep8(t, grammar.Hindi, "बनाओ चर उपयोगकर्ता रूप \"Alex\"\nमान संदेश जैसा Text रूप उपयोगकर्ता\nबनाओ स्थिर संख्या रूप 42")
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %v", diagnosticCodes(diags))
	}
	if len(file.Statements) != 3 {
		t.Fatalf("statement count=%d", len(file.Statements))
	}
	if got := file.Statements[0].(ast.VariableDecl).Name.Name; got != "उपयोगकर्ता" {
		t.Fatalf("identifier changed: %q", got)
	}
	if got := file.Statements[1].(ast.VariableDecl).Type.Name; got != "Text" {
		t.Fatalf("type changed: %q", got)
	}
	if got := file.Statements[2].(ast.VariableDecl).Name.Name; got != "संख्या" {
		t.Fatalf("identifier changed: %q", got)
	}
	mixed, mixedDiags := parseStep8(t, grammar.Hindi, "बनाओ चर userहिंदी रूप 1\nबनाओ चर हिंदीUser रूप 2\nबनाओ चर उपयोगकर्ता123 रूप 3\nबनाओ चर ব্যবহারকারী रूप 4")
	if len(mixedDiags) != 0 {
		t.Fatalf("mixed identifier diagnostics: %v", diagnosticCodes(mixedDiags))
	}
	for i, want := range []string{"userहिंदी", "हिंदीUser", "उपयोगकर्ता123", "ব্যবহারকারী"} {
		if got := mixed.Statements[i].(ast.VariableDecl).Name.Name; got != want {
			t.Fatalf("mixed identifier %d = %q want %q", i, got, want)
		}
	}
}

func TestStep8HindiFunctionsControlFlowCallsComponentsImports(t *testing.T) {
	src := "आयात \"ui/card\" रूप ui\nबनाओ फलन अभिवादन(नाम रूप Text, countASCII रूप Number) रूप Text {\nअगर नाम {\nलौटाओ नाम\n} वरना {\nवापस fallback1\n}\n}\nबनाओ घटक कार्ड(शीर्षक रूप नाम) {\nअभिवादन(शीर्षक, 1)\n}\n"
	file, diags := parseStep8(t, grammar.Hindi, src)
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %v", diagnosticCodes(diags))
	}
	if len(file.Statements) != 3 {
		t.Fatalf("statement count=%d", len(file.Statements))
	}
	if imp := file.Statements[0].(ast.ImportDecl); imp.Path != "\"ui/card\"" || imp.Alias.Name != "ui" {
		t.Fatalf("bad import %#v", imp)
	}
	fn := file.Statements[1].(ast.FunctionDecl)
	if fn.Name.Name != "अभिवादन" || fn.Parameters[1].Name != "countASCII" || fn.ReturnType.Name != "Text" {
		t.Fatalf("bad function %#v", fn)
	}
	if _, ok := fn.Body.Statements[0].(ast.IfStmt); !ok {
		t.Fatalf("missing if: %#v", fn.Body.Statements[0])
	}
	cmp := file.Statements[2].(ast.ComponentDecl)
	if cmp.Name.Name != "कार्ड" || cmp.Properties[0].Name != "शीर्षक" {
		t.Fatalf("bad component %#v", cmp)
	}
}

func TestStep8EnglishBengaliHindiASTEquivalence(t *testing.T) {
	en, enDiags := parseStep8(t, grammar.English, "import \"ui\" as ui\ncreate variable user as Text as \"Alex\"\nfn greet(name as Text) as Text { return name }\ncreate component Card(title as user) { greet(title) }")
	bn, bnDiags := parseStep8(t, grammar.Bengali, "আমদানি \"ui\" হিসেবে ui\nতৈরি চলক user হিসেবে Text হিসেবে \"Alex\"\nকাজ greet(name হিসেবে Text) হিসেবে Text { ফেরত name }\nতৈরি উপাদান Card(title হিসেবে user) { greet(title) }")
	hi, hiDiags := parseStep8(t, grammar.Hindi, "आयात \"ui\" रूप ui\nबनाओ चर user रूप Text रूप \"Alex\"\nकार्य greet(name रूप Text) रूप Text { लौटाओ name }\nबनाओ अवयव Card(title रूप user) { greet(title) }")
	if len(enDiags) != 0 || len(bnDiags) != 0 || len(hiDiags) != 0 {
		t.Fatalf("diagnostics en=%v bn=%v hi=%v", diagnosticCodes(enDiags), diagnosticCodes(bnDiags), diagnosticCodes(hiDiags))
	}
	if !reflect.DeepEqual(canonicalizeFile(en), canonicalizeFile(bn)) || !reflect.DeepEqual(canonicalizeFile(en), canonicalizeFile(hi)) {
		t.Fatalf("not equivalent\nen=%#v\nbn=%#v\nhi=%#v", canonicalizeFile(en), canonicalizeFile(bn), canonicalizeFile(hi))
	}
}

func TestStep8HindiSourcePositionsDiagnosticsAndRecovery(t *testing.T) {
	src := "बनाओ चर उपयोगकर्ता रूप \"Alex\"\nबनाओ चर परिणाम रूप;\nबनाओ चर बाद रूप 1"
	file, diags := parseStep8(t, grammar.Hindi, src)
	if len(diags) == 0 || diags[0].Code != "G2004" {
		t.Fatalf("want G2004 got %v", diagnosticCodes(diags))
	}
	semi := strings.Index(src, ";")
	want := source.PositionAt(src, semi)
	if diags[0].Span.Start != want || diags[0].Span.End != source.PositionAt(src, semi+1) {
		t.Fatalf("bad diagnostic span: %#v want start %#v", diags[0].Span, want)
	}
	if want.Line != 2 || want.Column != 19 || want.Offset != semi {
		t.Fatalf("unexpected UTF-8 position: %#v", want)
	}
	if len(file.Statements) != 2 {
		t.Fatalf("recovery should preserve later valid statement, got %d", len(file.Statements))
	}
}

func TestStep8HindiDiagnosticsLocalized(t *testing.T) {
	text := "बनाओ चर नाम रूप\n"
	files := source.NewFileMap()
	id := files.Add("हिन्दी.gogo", text)
	_, diags := parseStep8(t, grammar.Hindi, text)
	if len(diags) == 0 {
		t.Fatal("expected diagnostic")
	}
	diags[0].FileID = id
	if diags[0].Severity != diagnostics.Error || diags[0].Code != "G2004" {
		t.Fatalf("bad diagnostic identity: %#v", diags[0])
	}
	out := diagnostics.Renderer{Files: files, Locale: diagnostics.Hindi}.Text(diags)
	if !strings.Contains(out, "G2004") || !strings.Contains(out, "एक मान अपेक्षित") || !strings.Contains(out, "हिन्दी.gogo") {
		t.Fatalf("not localized/stable:\n%s", out)
	}
	jsonData, err := diagnostics.Renderer{Files: files, Locale: diagnostics.Hindi}.JSON(diags)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(jsonData)
	for _, want := range []string{"\"language\": \"hi\"", "\"code\": \"G2004\"", "हिन्दी.gogo", "एक मान अपेक्षित"} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("JSON diagnostic missing %q in %s", want, jsonText)
		}
	}
}

func TestStep8HindiRecoveryMalformedForms(t *testing.T) {
	for _, src := range []string{
		"बनाओ चर रूप 1; बनाओ चर बाद रूप 2",
		"बनाओ फलन टूटा(नाम रूप Text { लौटाओ नाम }; बनाओ चर बाद रूप 2",
		"अगर { लौटाओ x }; बनाओ चर बाद रूप 2",
		"बनाओ घटक (शीर्षक रूप x) {}; बनाओ चर बाद रूप 2",
		"आयात रूप bad; बनाओ चर बाद रूप 2",
		"बनाओ चर नाम रूप Text रूप; बनाओ चर बाद रूप 2",
	} {
		file, diags := parseStep8(t, grammar.Hindi, src)
		if len(diags) == 0 {
			t.Fatalf("expected diagnostic for %q", src)
		}
		if len(file.Statements) == 0 {
			t.Fatalf("parser failed to recover for %q with %v", src, diagnosticCodes(diags))
		}
	}
}

func TestStep8VocabularyScopedIdentifierRules(t *testing.T) {
	file, diags := parseStep8(t, grammar.English, "create variable बनाओ as \"identifier\"")
	if len(diags) != 0 {
		t.Fatalf("English diagnostics: %v", diagnosticCodes(diags))
	}
	if file.Statements[0].(ast.VariableDecl).Name.Name != "बनाओ" {
		t.Fatalf("Hindi word should remain an English identifier")
	}
	_, wrong := parseStep8(t, grammar.English, "बनाओ चर user रूप 1")
	if len(wrong) == 0 {
		t.Fatalf("English session should not parse Hindi keywords as English grammar")
	}
}

func TestStep8EveryHindiAliasParsesAsCanonicalSemantics(t *testing.T) {
	cases := []string{
		"बनाओ चर नाम रूप 1",
		"निर्माण मान नाम जैसा 1",
		"बनाओ स्थिर नाम के_रूप_में 1",
		"निर्माण अचर नाम रूप 1",
		"फलन f() रूप Text { लौटाओ x }",
		"कार्य f() जैसा Text { वापस x }",
		"फ़ंक्शन f(नाम के_रूप_में Text) के_रूप_में Text { लौटाओ नाम }",
		"अगर ready { लौटाओ yes } वरना { लौटाओ no }",
		"यदि ready { लौटाओ yes } अन्यथा { लौटाओ no }",
		"आयात \"pkg\" रूप pkg",
		"लाओ \"pkg\" जैसा pkg",
		"बनाओ घटक Card(शीर्षक रूप title) { f(शीर्षक) }",
		"निर्माण अवयव Card(शीर्षक जैसा title) { f(शीर्षक) }",
	}
	for _, src := range cases {
		file, diags := parseStep8(t, grammar.Hindi, src)
		if len(diags) != 0 {
			t.Fatalf("%q diagnostics: %v", src, diagnosticCodes(diags))
		}
		if len(file.Statements) == 0 {
			t.Fatalf("%q did not produce canonical AST statements", src)
		}
	}
}

func TestStep8FromAliasIsReservedOnlyInHindi(t *testing.T) {
	hi := grammar.Must(grammar.Hindi)
	if got, ok := hi.Lookup("से"); !ok || got != grammar.KeywordFrom {
		t.Fatalf("से = %v,%v want KeywordFrom,true", got, ok)
	}
	for _, lang := range []grammar.Language{grammar.English, grammar.Bengali} {
		if grammar.Must(lang).IsKeyword("से") {
			t.Fatalf("से must not be reserved in %s", lang)
		}
	}
}

func TestStep8EveryEnglishAliasParsesAsCanonicalSemantics(t *testing.T) {
	cases := []string{
		"create variable name as 1",
		"let name as 1",
		"var name as 1",
		"create constant name as 1",
		"const name as 1",
		"function f() as Text { return x }",
		"fn f() as Text { return x }",
		"if ready { return yes } else { return no }",
		"import \"pkg\" as pkg",
		"create component Card(title as value) { f(title) }",
	}
	for _, src := range cases {
		file, diags := parseStep8(t, grammar.English, src)
		if len(diags) != 0 {
			t.Fatalf("%q diagnostics: %v", src, diagnosticCodes(diags))
		}
		if len(file.Statements) == 0 {
			t.Fatalf("%q did not produce canonical AST statements", src)
		}
	}
	if got, ok := grammar.Must(grammar.English).Lookup("from"); !ok || got != grammar.KeywordFrom {
		t.Fatalf("from = %v,%v want KeywordFrom,true", got, ok)
	}
}

func TestStep8EveryBengaliAliasParsesAsCanonicalSemantics(t *testing.T) {
	cases := []string{
		"তৈরি চলক নাম হিসেবে 1",
		"ঘোষণা ধরি নাম রূপে 1",
		"তৈরি ধ্রুবক নাম হিসেবে 1",
		"ঘোষণা অপরিবর্তনীয় নাম রূপে 1",
		"ফাংশন f() হিসেবে Text { ফেরত x }",
		"কাজ f() রূপে Text { ফেরাও x }",
		"যদি ready { ফেরত yes } নইলে { ফেরত no }",
		"যদি ready { ফেরত yes } অন্যথায় { ফেরত no }",
		"আমদানি \"pkg\" হিসেবে pkg",
		"ইমপোর্ট \"pkg\" রূপে pkg",
		"তৈরি কম্পোনেন্ট Card(শিরোনাম হিসেবে value) { f(শিরোনাম) }",
		"ঘোষণা উপাদান Card(শিরোনাম রূপে value) { f(শিরোনাম) }",
	}
	for _, src := range cases {
		file, diags := parseStep8(t, grammar.Bengali, src)
		if len(diags) != 0 {
			t.Fatalf("%q diagnostics: %v", src, diagnosticCodes(diags))
		}
		if len(file.Statements) == 0 {
			t.Fatalf("%q did not produce canonical AST statements", src)
		}
	}
	if got, ok := grammar.Must(grammar.Bengali).Lookup("থেকে"); !ok || got != grammar.KeywordFrom {
		t.Fatalf("থেকে = %v,%v want KeywordFrom,true", got, ok)
	}
}
