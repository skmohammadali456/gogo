package parser

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

func parseStep6(t *testing.T, text string) (ast.File, []string) {
	t.Helper()
	lx := lexer.New(source.File{ID: 1, Path: "step6.gogo", Text: text})
	p := New(lx.LexAll(), WithVocabulary(grammar.Must(grammar.English)))
	file := p.ParseFile()
	var codes []string
	for _, d := range p.Diagnostics() {
		codes = append(codes, d.Code)
	}
	return file, codes
}

func TestStep6ReadableAndConciseDeclarationsEquivalent(t *testing.T) {
	readable, rdiag := parseStep6(t, `create variable user as "Alex"`)
	concise, cdiag := parseStep6(t, `let user as "Alex"`)
	if len(rdiag) != 0 || len(cdiag) != 0 {
		t.Fatalf("diagnostics readable=%v concise=%v", rdiag, cdiag)
	}
	r := readable.Statements[0].(ast.VariableDecl)
	c := concise.Statements[0].(ast.VariableDecl)
	if r.Name.Name != c.Name.Name || r.Value.(ast.Literal).Text != c.Value.(ast.Literal).Text {
		t.Fatalf("not equivalent: %#v %#v", r, c)
	}
}

func TestStep6TypedFunctionImportComponentAndNewlines(t *testing.T) {
	file, codes := parseStep6(t, "import \"ui\" as ui\ncreate variable user as Text as \"Alex\"\nfn greet(name as Text) as Text {\nreturn name\n}\ncreate component Card(title as user) {\nuser\n}\n")
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	if len(file.Statements) != 4 {
		t.Fatalf("statement count=%d", len(file.Statements))
	}
	if imp := file.Statements[0].(ast.ImportDecl); imp.Path != `"ui"` || imp.Alias == nil || imp.Alias.Name != "ui" {
		t.Fatalf("bad import %#v", imp)
	}
	if v := file.Statements[1].(ast.VariableDecl); v.Type == nil || v.Type.Name != "Text" {
		t.Fatalf("bad typed var %#v", v)
	}
	if fn := file.Statements[2].(ast.FunctionDecl); fn.ReturnType == nil || fn.ParameterTypes[0] == nil {
		t.Fatalf("bad function types %#v", fn)
	}
	if cmp := file.Statements[3].(ast.ComponentDecl); cmp.Name.Name != "Card" || len(cmp.Properties) != 1 {
		t.Fatalf("bad component %#v", cmp)
	}
}

func TestStep6UnicodeIdentifiersRemainIdentifiers(t *testing.T) {
	file, codes := parseStep6(t, "create variable ব্যবহারকারী as \"Alex\"\ncreate variable उपयोगकर्ता as ব্যবহারকারী")
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	if file.Statements[0].(ast.VariableDecl).Name.Name != "ব্যবহারকারী" {
		t.Fatalf("Bengali identifier changed")
	}
	if file.Statements[1].(ast.VariableDecl).Name.Name != "उपयोगकर्ता" {
		t.Fatalf("Hindi identifier changed")
	}
}

func TestStep6RecoveryDiagnostics(t *testing.T) {
	_, codes := parseStep6(t, "import as bad\ncreate component (x as 1) {}\ncreate variable as")
	if len(codes) == 0 {
		t.Fatalf("expected diagnostics")
	}
}
