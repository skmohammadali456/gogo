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

func parseStep7(t *testing.T, lang grammar.Language, text string) (ast.File, []diagnostics.Diagnostic) {
	t.Helper()
	lx := lexer.New(source.File{ID: 1, Path: "step7.gogo", Text: text})
	p := New(lx.LexAll(), WithVocabulary(grammar.Must(lang)))
	file := p.ParseFile()
	return file, p.Diagnostics()
}

func diagnosticCodes(ds []diagnostics.Diagnostic) []string {
	codes := make([]string, len(ds))
	for i, d := range ds {
		codes[i] = d.Code
	}
	return codes
}

type canonicalFile struct{ Statements []any }
type canonicalVar struct{ Name, Type, Value string }
type canonicalImport struct{ Path, Alias string }
type canonicalFunction struct {
	Name               string
	Params, ParamTypes []string
	ReturnType         string
	Body               []any
}
type canonicalIf struct {
	Condition  string
	Then, Else []any
}
type canonicalReturn struct{ Value string }
type canonicalComponent struct {
	Name  string
	Props []canonicalProp
	Body  []any
}
type canonicalProp struct{ Name, Value string }
type canonicalExpr struct {
	Kind, Text string
	Args       []canonicalExpr
}

func canonicalizeFile(f ast.File) canonicalFile {
	out := canonicalFile{Statements: make([]any, 0, len(f.Statements))}
	for _, s := range f.Statements {
		out.Statements = append(out.Statements, canonicalizeStmt(s))
	}
	return out
}
func canonicalizeStmt(s ast.Stmt) any {
	switch v := s.(type) {
	case ast.VariableDecl:
		typ := ""
		if v.Type != nil {
			typ = v.Type.Name
		}
		return canonicalVar{v.Name.Name, typ, canonicalizeExpr(v.Value).Text}
	case ast.ImportDecl:
		alias := ""
		if v.Alias != nil {
			alias = v.Alias.Name
		}
		return canonicalImport{v.Path, alias}
	case ast.FunctionDecl:
		pts := make([]string, len(v.ParameterTypes))
		ps := make([]string, len(v.Parameters))
		for i := range v.Parameters {
			ps[i] = v.Parameters[i].Name
			if v.ParameterTypes[i] != nil {
				pts[i] = v.ParameterTypes[i].Name
			}
		}
		rt := ""
		if v.ReturnType != nil {
			rt = v.ReturnType.Name
		}
		return canonicalFunction{v.Name.Name, ps, pts, rt, canonicalizeBlock(v.Body)}
	case ast.IfStmt:
		var els []any
		if v.Else != nil {
			els = canonicalizeBlock(*v.Else)
		}
		return canonicalIf{canonicalizeExpr(v.Condition).Text, canonicalizeBlock(v.Then), els}
	case ast.ReturnStmt:
		return canonicalReturn{canonicalizeExpr(v.Value).Text}
	case ast.ComponentDecl:
		props := make([]canonicalProp, len(v.Properties))
		for i, p := range v.Properties {
			props[i] = canonicalProp{p.Name, canonicalizeExpr(p.Value).Text}
		}
		return canonicalComponent{v.Name.Name, props, canonicalizeBlock(v.Body)}
	case ast.ExprStmt:
		return canonicalizeExpr(v.Expression)
	default:
		return reflect.TypeOf(s).String()
	}
}
func canonicalizeBlock(b ast.BlockStmt) []any {
	out := make([]any, 0, len(b.Statements))
	for _, s := range b.Statements {
		out = append(out, canonicalizeStmt(s))
	}
	return out
}
func canonicalizeExpr(e ast.Expr) canonicalExpr {
	switch v := e.(type) {
	case ast.Identifier:
		return canonicalExpr{Kind: "id", Text: v.Name}
	case ast.Literal:
		return canonicalExpr{Kind: "lit", Text: v.Text}
	case ast.BinaryExpr:
		return canonicalExpr{Kind: "binary", Text: canonicalizeExpr(v.Left).Text + v.Operator + canonicalizeExpr(v.Right).Text}
	case ast.CallExpr:
		args := make([]canonicalExpr, len(v.Arguments))
		for i, a := range v.Arguments {
			args[i] = canonicalizeExpr(a)
		}
		return canonicalExpr{Kind: "call", Text: canonicalizeExpr(v.Callee).Text, Args: args}
	default:
		return canonicalExpr{Kind: reflect.TypeOf(e).String()}
	}
}

func TestStep7BengaliDeclarationsAndUnicodeIdentifiers(t *testing.T) {
	file, diags := parseStep7(t, grammar.Bengali, "তৈরি চলক ব্যবহারকারী হিসেবে \"Alex\"\nধরি বার্তা রূপে Text হিসেবে ব্যবহারকারী\nতৈরি ধ্রুবক সংখ্যা হিসেবে 42")
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %v", diagnosticCodes(diags))
	}
	if len(file.Statements) != 3 {
		t.Fatalf("statement count=%d", len(file.Statements))
	}
	if got := file.Statements[0].(ast.VariableDecl).Name.Name; got != "ব্যবহারকারী" {
		t.Fatalf("identifier changed: %q", got)
	}
	if got := file.Statements[1].(ast.VariableDecl).Type.Name; got != "Text" {
		t.Fatalf("type changed: %q", got)
	}
	if got := file.Statements[2].(ast.VariableDecl).Name.Name; got != "সংখ্যা" {
		t.Fatalf("identifier changed: %q", got)
	}
	mixed, mixedDiags := parseStep7(t, grammar.Bengali, "তৈরি চলক userবাংলা হিসেবে 1\nতৈরি চলক বাংলাUser হিসেবে 2\nতৈরি চলক ব্যবহারকারী123 হিসেবে 3\nতৈরি চলক परिणाम হিসেবে 4")
	if len(mixedDiags) != 0 {
		t.Fatalf("mixed identifier diagnostics: %v", diagnosticCodes(mixedDiags))
	}
	for i, want := range []string{"userবাংলা", "বাংলাUser", "ব্যবহারকারী123", "परिणाम"} {
		if got := mixed.Statements[i].(ast.VariableDecl).Name.Name; got != want {
			t.Fatalf("mixed identifier %d = %q want %q", i, got, want)
		}
	}
}

func TestStep7BengaliFunctionsControlFlowCallsComponentsImports(t *testing.T) {
	src := "আমদানি \"ui/card\" হিসেবে ui\nতৈরি ফাংশন শুভেচ্ছা(নাম হিসেবে Text, countASCII হিসেবে Number) হিসেবে Text {\nযদি নাম {\nফেরত নাম\n} নইলে {\nফেরাও fallback1\n}\n}\nতৈরি কম্পোনেন্ট কার্ড(শিরোনাম হিসেবে নাম) {\nশুভেচ্ছা(শিরোনাম, 1)\n}\n"
	file, diags := parseStep7(t, grammar.Bengali, src)
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
	if fn.Name.Name != "শুভেচ্ছা" || fn.Parameters[1].Name != "countASCII" || fn.ReturnType.Name != "Text" {
		t.Fatalf("bad function %#v", fn)
	}
	if _, ok := fn.Body.Statements[0].(ast.IfStmt); !ok {
		t.Fatalf("missing if: %#v", fn.Body.Statements[0])
	}
	cmp := file.Statements[2].(ast.ComponentDecl)
	if cmp.Name.Name != "কার্ড" || cmp.Properties[0].Name != "শিরোনাম" {
		t.Fatalf("bad component %#v", cmp)
	}
}

func TestStep7BengaliAndEnglishASTEquivalence(t *testing.T) {
	en, enDiags := parseStep7(t, grammar.English, "import \"ui\" as ui\ncreate variable user as Text as \"Alex\"\nfn greet(name as Text) as Text { return name }\ncreate component Card(title as user) { greet(title) }")
	bn, bnDiags := parseStep7(t, grammar.Bengali, "আমদানি \"ui\" হিসেবে ui\nতৈরি চলক user হিসেবে Text হিসেবে \"Alex\"\nকাজ greet(name হিসেবে Text) হিসেবে Text { ফেরত name }\nতৈরি উপাদান Card(title হিসেবে user) { greet(title) }")
	if len(enDiags) != 0 || len(bnDiags) != 0 {
		t.Fatalf("diagnostics en=%v bn=%v", diagnosticCodes(enDiags), diagnosticCodes(bnDiags))
	}
	if !reflect.DeepEqual(canonicalizeFile(en), canonicalizeFile(bn)) {
		t.Fatalf("not equivalent\nen=%#v\nbn=%#v", canonicalizeFile(en), canonicalizeFile(bn))
	}
}

func TestStep7BengaliSourcePositionsDiagnosticsAndRecovery(t *testing.T) {
	src := "তৈরি চলক ব্যবহারকারী হিসেবে \"Alex\"\nতৈরি চলক ফলাফল হিসেবে;\nতৈরি চলক পরে হিসেবে 1"
	file, diags := parseStep7(t, grammar.Bengali, src)
	if len(diags) == 0 || diags[0].Code != "G2004" {
		t.Fatalf("want G2004 got %v", diagnosticCodes(diags))
	}
	want := source.PositionAt(src, strings.Index(src, ";"))
	if diags[0].Span.Start != want || diags[0].Span.End != source.PositionAt(src, strings.Index(src, ";")+1) {
		t.Fatalf("bad diagnostic span: %#v want start %#v", diags[0].Span, want)
	}
	if want.Line != 2 || want.Column != 22 {
		t.Fatalf("unexpected UTF-8 human column: %#v", want)
	}
	if len(file.Statements) != 2 {
		t.Fatalf("recovery should preserve later valid statement, got %d", len(file.Statements))
	}
}

func TestStep7VocabularyScopedIdentifierRules(t *testing.T) {
	file, diags := parseStep7(t, grammar.English, "create variable তৈরি as \"identifier\"")
	if len(diags) != 0 {
		t.Fatalf("English diagnostics: %v", diagnosticCodes(diags))
	}
	if file.Statements[0].(ast.VariableDecl).Name.Name != "তৈরি" {
		t.Fatalf("Bengali word should remain an English identifier")
	}
	_, wrong := parseStep7(t, grammar.English, "তৈরি চলক user হিসেবে 1")
	if len(wrong) == 0 {
		t.Fatalf("English session should not parse Bengali keywords as English grammar")
	}
}

func TestStep7BengaliDiagnosticsLocalized(t *testing.T) {
	text := "তৈরি চলক নাম হিসেবে\n"
	files := source.NewFileMap()
	id := files.Add("বাংলা.gogo", text)
	_, diags := parseStep7(t, grammar.Bengali, text)
	if len(diags) == 0 {
		t.Fatal("expected diagnostic")
	}
	diags[0].FileID = id
	if diags[0].Severity != diagnostics.Error || diags[0].Code != "G2004" {
		t.Fatalf("bad diagnostic identity: %#v", diags[0])
	}
	out := diagnostics.Renderer{Files: files, Locale: diagnostics.Bengali}.Text(diags)
	if !strings.Contains(out, "G2004") || !strings.Contains(out, "একটি মান প্রত্যাশিত") || !strings.Contains(out, "বাংলা.gogo") {
		t.Fatalf("not localized/stable:\n%s", out)
	}
	if strings.Contains(out, "G২০০৪") {
		t.Fatalf("code localized: %s", out)
	}
	jsonData, err := diagnostics.Renderer{Files: files, Locale: diagnostics.Bengali}.JSON(diags)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(jsonData)
	for _, want := range []string{"\"language\": \"bn\"", "\"code\": \"G2004\"", "বাংলা.gogo", "একটি মান প্রত্যাশিত"} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("JSON diagnostic missing %q in %s", want, jsonText)
		}
	}
}

func TestStep7BengaliRecoveryMalformedForms(t *testing.T) {
	for _, src := range []string{
		"তৈরি চলক হিসেবে 1; তৈরি চলক পরে হিসেবে 2",
		"তৈরি ফাংশন ভাঙা(নাম হিসেবে Text { ফেরত নাম }; তৈরি চলক পরে হিসেবে 2",
		"যদি { ফেরত x }; তৈরি চলক পরে হিসেবে 2",
		"তৈরি কম্পোনেন্ট (শিরোনাম হিসেবে x) {}; তৈরি চলক পরে হিসেবে 2",
		"আমদানি হিসেবে bad; তৈরি চলক পরে হিসেবে 2",
	} {
		file, diags := parseStep7(t, grammar.Bengali, src)
		if len(diags) == 0 {
			t.Fatalf("expected diagnostic for %q", src)
		}
		if len(file.Statements) == 0 {
			t.Fatalf("parser failed to recover for %q with %v", src, diagnosticCodes(diags))
		}
	}
}
