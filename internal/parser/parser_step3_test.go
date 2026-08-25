package parser

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

func parseStep3(text string) (*Parser, ast.File) {
	l := lexer.New(source.File{ID: 1, Path: "step3.gogo", Text: text})
	p := New(l.LexAll())
	return p, p.ParseFile()
}

func TestParseFunctionIfElse(t *testing.T) {
	p, file := parseStep3(`create function greet(name) { if name { return name } else { return "unknown" } }`)
	if len(p.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics())
	}
	fn, ok := file.Statements[0].(ast.FunctionDecl)
	if !ok || len(fn.Parameters) != 1 {
		t.Fatalf("unexpected function AST: %#v", file.Statements[0])
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("unexpected function body: %#v", fn.Body)
	}
	if _, ok := fn.Body.Statements[0].(ast.IfStmt); !ok {
		t.Fatalf("expected if statement")
	}
}

func TestParseArraysObjectsMembersAndIndexes(t *testing.T) {
	p, file := parseStep3(`items[0].name`)
	if len(p.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics())
	}
	member, ok := file.Statements[0].(ast.ExprStmt).Expression.(ast.MemberExpr)
	if !ok || member.Name != "name" {
		t.Fatalf("unexpected member AST: %#v", file.Statements[0])
	}
	if _, ok := member.Object.(ast.IndexExpr); !ok {
		t.Fatalf("expected index before member")
	}

	p, file = parseStep3(`[1, 2, 3]`)
	if len(p.Diagnostics()) != 0 {
		t.Fatalf("unexpected array diagnostics: %v", p.Diagnostics())
	}
	array, ok := file.Statements[0].(ast.ExprStmt).Expression.(ast.ArrayExpr)
	if !ok || len(array.Items) != 3 {
		t.Fatalf("unexpected array AST: %#v", file.Statements[0])
	}

	p, file = parseStep3(`create variable data as {name: "Alex", age: 20}`)
	if len(p.Diagnostics()) != 0 {
		t.Fatalf("unexpected object diagnostics: %v", p.Diagnostics())
	}
	decl, ok := file.Statements[0].(ast.VariableDecl)
	if !ok {
		t.Fatalf("expected variable declaration, got %#v", file.Statements[0])
	}
	object, ok := decl.Value.(ast.ObjectExpr)
	if !ok || len(object.Properties) != 2 {
		t.Fatalf("unexpected object AST: %#v", decl.Value)
	}
}

func TestParseMultilingualFunctionNamesAndParameters(t *testing.T) {
	p, file := parseStep3(`create function অভিবাদন(नाम) { return नाम }`)
	if len(p.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics())
	}
	fn := file.Statements[0].(ast.FunctionDecl)
	if fn.Name.Name != "অভিবাদন" || fn.Parameters[0].Name != "नाम" {
		t.Fatalf("unexpected multilingual AST: %#v", fn)
	}
}

func TestParseRecoveryAfterInvalidStatement(t *testing.T) {
	p, file := parseStep3(`create function broken( { return 1 } create variable good as 2`)
	if len(p.Diagnostics()) == 0 {
		t.Fatal("expected parser diagnostics")
	}
	if len(file.Statements) == 0 {
		t.Fatal("parser should recover and continue")
	}
}
