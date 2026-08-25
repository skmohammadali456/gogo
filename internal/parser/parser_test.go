package parser

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
	"github.com/skmohammadali786/gogo/internal/ast"
)

func parse(text string) (*Parser, ast.File) {
	l := lexer.New(source.File{ID: 1, Path: "main.gogo", Text: text})
	return func() (*Parser, ast.File) { p := New(l.LexAll()); return p, p.ParseFile() }()
}

func TestParseVariableDeclaration(t *testing.T) {
	p, file := parse(`create variable user as "Alex"`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	if len(file.Statements) != 1 { t.Fatalf("got %d statements", len(file.Statements)) }
	decl, ok := file.Statements[0].(ast.VariableDecl)
	if !ok || decl.Name.Name != "user" { t.Fatalf("unexpected declaration: %#v", file.Statements[0]) }
}

func TestParseExpressionPrecedence(t *testing.T) {
	p, file := parse(`a + b * c`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	expr := file.Statements[0].(ast.ExprStmt).Expression
	root, ok := expr.(ast.BinaryExpr)
	if !ok || root.Operator != "+" { t.Fatalf("unexpected root: %#v", expr) }
	if _, ok := root.Right.(ast.BinaryExpr); !ok { t.Fatalf("expected multiplication on right: %#v", root.Right) }
}

func TestParseCall(t *testing.T) {
	p, file := parse(`greet("Alex", 42)`)
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	call, ok := file.Statements[0].(ast.ExprStmt).Expression.(ast.CallExpr)
	if !ok || len(call.Arguments) != 2 { t.Fatalf("unexpected call: %#v", file.Statements[0]) }
}

func TestParseMultilingualIdentifiers(t *testing.T) {
	p, file := parse("বাংলা + हिन्दी + english")
	if len(p.Diagnostics()) != 0 { t.Fatalf("unexpected diagnostics: %v", p.Diagnostics()) }
	if len(file.Statements) != 1 { t.Fatalf("expected one statement") }
}

func TestParseMissingBraceDiagnostic(t *testing.T) {
	p, _ := parse(`{ create variable user as "Alex"`)
	if len(p.Diagnostics()) == 0 || p.Diagnostics()[0].Code != "G2006" { t.Fatalf("expected G2006, got %v", p.Diagnostics()) }
}

func TestParseMissingValueDiagnostic(t *testing.T) {
	p, _ := parse(`create variable user as`)
	if len(p.Diagnostics()) == 0 || p.Diagnostics()[0].Code != "G2004" { t.Fatalf("expected G2004, got %v", p.Diagnostics()) }
}

func TestParserNeverUsesInvalidTokensSilently(t *testing.T) {
	tokens := []token.Token{token.New(token.Invalid, "@", source.Span{}), token.New(token.EOF, "", source.Span{})}
	p := New(tokens)
	p.ParseFile()
	if len(p.Diagnostics()) == 0 { t.Fatal("expected diagnostic for invalid token") }
}
