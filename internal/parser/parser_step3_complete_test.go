package parser

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

func parseStep3Complete(text string) (*Parser, ast.File) {
	l := lexer.New(source.File{ID: 1, Path: "step3.gogo", Text: text})
	p := New(l.LexAll())
	return p, p.ParseFile()
}

func TestStep3AssignmentAndConditional(t *testing.T) {
	p, file := parseStep3Complete(`create variable result as 0; result = ready ? 10 : 20`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	if len(file.Statements) != 2 { t.Fatalf("expected 2 statements, got %d", len(file.Statements)) }
	expr := file.Statements[1].(ast.ExprStmt).Expression
	assignment, ok := expr.(ast.AssignmentExpr)
	if !ok || assignment.Operator != "=" { t.Fatalf("expected assignment, got %#v", expr) }
	if _, ok := assignment.Right.(ast.ConditionalExpr); !ok { t.Fatalf("expected conditional right side, got %#v", assignment.Right) }
}

func TestStep3RightAssociativeOperators(t *testing.T) {
	p, file := parseStep3Complete(`a ** b ** c`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	root := file.Statements[0].(ast.ExprStmt).Expression.(ast.BinaryExpr)
	if root.Operator != "**" { t.Fatalf("unexpected root operator: %s", root.Operator) }
	if _, ok := root.Right.(ast.BinaryExpr); !ok { t.Fatalf("expected right associative exponentiation, got %#v", root.Right) }
}

func TestStep3NestedPostfixExpressions(t *testing.T) {
	p, file := parseStep3Complete(`[1, 2, 3][0].name(42)`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	call, ok := file.Statements[0].(ast.ExprStmt).Expression.(ast.CallExpr)
	if !ok || len(call.Arguments) != 1 { t.Fatalf("expected call with one argument, got %#v", file.Statements[0]) }
	member, ok := call.Callee.(ast.MemberExpr)
	if !ok || member.Name != "name" { t.Fatalf("expected member access, got %#v", call.Callee) }
	if _, ok := member.Object.(ast.IndexExpr); !ok { t.Fatalf("expected index before member, got %#v", member.Object) }
}

func TestStep3AssignmentTargetDiagnostic(t *testing.T) {
	p, _ := parseStep3Complete(`1 = value`)
	if len(p.Diagnostics()) == 0 || p.Diagnostics()[0].Code != "G2028" { t.Fatalf("expected G2028, got %v", p.Diagnostics()) }
}

func TestStep3MalformedCollectionsRecover(t *testing.T) {
	p, file := parseStep3Complete(`create variable broken as [1, 2; create variable good as 3`)
	if len(p.Diagnostics()) == 0 { t.Fatal("expected diagnostics for malformed array") }
	if len(file.Statements) < 1 { t.Fatal("parser should retain recoverable statements") }
}

func TestStep3FunctionTrailingComma(t *testing.T) {
	p, file := parseStep3Complete(`create function greet(first, second,) { return first }`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	fn, ok := file.Statements[0].(ast.FunctionDecl)
	if !ok || len(fn.Parameters) != 2 { t.Fatalf("unexpected function: %#v", file.Statements[0]) }
}

func TestStep3ObjectAndArrayTrailingCommas(t *testing.T) {
	p, file := parseStep3Complete(`create variable data as {name: "Alex", tags: ["go", "gogo",],}`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	decl := file.Statements[0].(ast.VariableDecl)
	obj, ok := decl.Value.(ast.ObjectExpr)
	if !ok || len(obj.Properties) != 2 { t.Fatalf("unexpected object: %#v", decl.Value) }
	array, ok := obj.Properties[1].Value.(ast.ArrayExpr)
	if !ok || len(array.Items) != 2 { t.Fatalf("unexpected array: %#v", obj.Properties[1].Value) }
}
