package parser

import (
	"os"
	"reflect"
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

type semanticFile struct{ Statements []any }

func parseWithGrammar(t *testing.T, lang grammar.Language, text string) (ast.File, []string) {
	t.Helper()
	file := source.File{ID: 1, Path: string(lang) + ".gogo", Text: text}
	l := lexer.New(file)
	p := New(l.LexAll(), WithVocabulary(grammar.Must(lang)))
	parsed := p.ParseFile()
	var codes []string
	for _, d := range append(l.Diagnostics(), p.Diagnostics()...) {
		codes = append(codes, d.Code)
	}
	return parsed, codes
}

func canonical(n ast.Node) any {
	switch v := n.(type) {
	case ast.File:
		items := make([]any, len(v.Statements))
		for i, s := range v.Statements {
			items[i] = canonical(s)
		}
		return semanticFile{items}
	case ast.VariableDecl:
		return map[string]any{"kind": "var", "name": v.Name.Name, "value": canonical(v.Value)}
	case ast.FunctionDecl:
		params := make([]string, len(v.Parameters))
		for i, p := range v.Parameters {
			params[i] = p.Name
		}
		return map[string]any{"kind": "fn", "name": v.Name.Name, "params": params, "body": canonical(v.Body)}
	case ast.ReturnStmt:
		return map[string]any{"kind": "return", "value": canonical(v.Value)}
	case ast.IfStmt:
		m := map[string]any{"kind": "if", "cond": canonical(v.Condition), "then": canonical(v.Then)}
		if v.Else != nil {
			m["else"] = canonical(*v.Else)
		}
		return m
	case ast.BlockStmt:
		items := make([]any, len(v.Statements))
		for i, s := range v.Statements {
			items[i] = canonical(s)
		}
		return map[string]any{"kind": "block", "stmts": items}
	case ast.ExprStmt:
		return map[string]any{"kind": "exprstmt", "expr": canonical(v.Expression)}
	case ast.Identifier:
		return map[string]any{"kind": "id", "name": v.Name}
	case ast.Literal:
		return map[string]any{"kind": "lit", "text": v.Text}
	case ast.BinaryExpr:
		return map[string]any{"kind": "binary", "op": v.Operator, "left": canonical(v.Left), "right": canonical(v.Right)}
	case ast.AssignmentExpr:
		return map[string]any{"kind": "assign", "op": v.Operator, "left": canonical(v.Left), "right": canonical(v.Right)}
	case ast.CallExpr:
		args := make([]any, len(v.Arguments))
		for i, a := range v.Arguments {
			args[i] = canonical(a)
		}
		return map[string]any{"kind": "call", "callee": canonical(v.Callee), "args": args}
	case ast.MemberExpr:
		return map[string]any{"kind": "member", "object": canonical(v.Object), "name": v.Name, "optional": v.Optional}
	case ast.IndexExpr:
		return map[string]any{"kind": "index", "object": canonical(v.Object), "index": canonical(v.Index)}
	case ast.ArrayExpr:
		items := make([]any, len(v.Items))
		for i, it := range v.Items {
			items[i] = canonical(it)
		}
		return map[string]any{"kind": "array", "items": items}
	case ast.ObjectExpr:
		props := make([]any, len(v.Properties))
		for i, p := range v.Properties {
			props[i] = map[string]any{"key": p.Key, "value": canonical(p.Value)}
		}
		return map[string]any{"kind": "object", "props": props}
	default:
		return nil
	}
}

func readTestSource(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestVocabularyASTEquivalence(t *testing.T) {
	programs := map[grammar.Language]string{
		grammar.English: readTestSource(t, "../../testdata/grammar/english.gogo"),
		grammar.Bengali: readTestSource(t, "../../testdata/grammar/bengali.gogo"),
		grammar.Hindi:   readTestSource(t, "../../testdata/grammar/hindi.gogo"),
	}
	var want any
	for lang, text := range programs {
		parsed, codes := parseWithGrammar(t, lang, text)
		if len(codes) != 0 {
			t.Fatalf("%s diagnostics: %v", lang, codes)
		}
		got := canonical(parsed)
		if want == nil {
			want = got
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s AST not equivalent\ngot %#v\nwant %#v", lang, got, want)
		}
	}
}

func TestLocalizedSourceSpansUseOriginalText(t *testing.T) {
	parsed, codes := parseWithGrammar(t, grammar.Bengali, "তৈরি চলক নাম হিসেবে \"Alex\"")
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	decl := parsed.Statements[0].(ast.VariableDecl)
	if decl.Name.Name != "নাম" || decl.Name.Span.Start.Offset != len("তৈরি চলক ") || decl.Name.Span.End.Offset != len("তৈরি চলক নাম") {
		t.Fatalf("bad Bengali identifier/span: %#v", decl.Name)
	}
}

func TestCrossVocabularyKeywordRejected(t *testing.T) {
	_, codes := parseWithGrammar(t, grammar.Bengali, "create চলক user হিসেবে 1")
	if len(codes) == 0 {
		t.Fatal("expected diagnostic for English keyword in Bengali grammar mode")
	}
}

func TestLocalizedUnknownIdentifierExpression(t *testing.T) {
	parsed, codes := parseWithGrammar(t, grammar.Hindi, "उपयोगकर्ता")
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	expr := parsed.Statements[0].(ast.ExprStmt).Expression.(ast.Identifier)
	if expr.Name != "उपयोगकर्ता" {
		t.Fatalf("identifier changed: %q", expr.Name)
	}
}

func TestLocalizedGrammarDiagnosticsAndRecovery(t *testing.T) {
	parsed, codes := parseWithGrammar(t, grammar.Hindi, "बनाओ चर user 1; बनाओ चर next के_रूप_में 2")
	if len(codes) == 0 || codes[0] != "G2003" {
		t.Fatalf("expected missing Hindi 'as' diagnostic, got %v", codes)
	}
	if len(parsed.Statements) != 2 {
		t.Fatalf("parser should recover to next Hindi declaration, got %d statements", len(parsed.Statements))
	}
	if got := parsed.Statements[1].(ast.VariableDecl).Name.Name; got != "next" {
		t.Fatalf("recovered declaration name = %q", got)
	}
}

func TestMixedUnicodeIdentifiersRemainSourceData(t *testing.T) {
	parsed, codes := parseWithGrammar(t, grammar.Bengali, "তৈরি চলক ব্যবহারकर्ता হিসেবে user_নাম + नाम")
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	decl := parsed.Statements[0].(ast.VariableDecl)
	if decl.Name.Name != "ব্যবহারकर्ता" {
		t.Fatalf("mixed-script variable name changed: %q", decl.Name.Name)
	}
	expr := decl.Value.(ast.BinaryExpr)
	if expr.Left.(ast.Identifier).Name != "user_নাম" || expr.Right.(ast.Identifier).Name != "नाम" {
		t.Fatalf("mixed Unicode identifiers changed: %#v", expr)
	}
}
