package lexer

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

func lex(text string) (*Lexer, []token.Token) {
	l := New(source.File{ID: 1, Path: "main.gogo", Text: text})
	return l, l.LexAll()
}

func TestLexBasicProgram(t *testing.T) {
	l, tokens := lex(`create variable user as "Alex"`)
	want := []token.Kind{token.Identifier, token.Identifier, token.Identifier, token.Identifier, token.String, token.EOF}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d: %#v", len(tokens), len(want), tokens)
	}
	for i, kind := range want {
		if tokens[i].Kind != kind {
			t.Fatalf("token %d: got %v, want %v", i, tokens[i].Kind, kind)
		}
	}
	if len(l.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", l.Diagnostics())
	}
}

func TestLexUnicodeIdentifiers(t *testing.T) {
	l, tokens := lex("বাংলা নাম\nहिन्दी शहर")
	if len(tokens) != 5 || tokens[0].Kind != token.Identifier || tokens[1].Kind != token.Identifier || tokens[3].Kind != token.Identifier {
		t.Fatalf("unexpected Unicode tokens: %#v", tokens)
	}
	if len(l.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", l.Diagnostics())
	}
}

func TestLexNumbersAndOperators(t *testing.T) {
	_, tokens := lex("42 10.5 1e6 2.5E-3 == != += -> && || <= >=")
	want := []token.Kind{token.Number, token.Number, token.Number, token.Number, token.EqualEqual, token.BangEqual, token.PlusEqual, token.Arrow, token.AndAnd, token.OrOr, token.LessEqual, token.GreaterEqual, token.EOF}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(want))
	}
	for i, kind := range want {
		if tokens[i].Kind != kind {
			t.Errorf("token %d: got %v, want %v", i, tokens[i].Kind, kind)
		}
	}
}

func TestLexInvalidNumberDiagnostics(t *testing.T) {
	for _, input := range []string{"10.", "1e", "1e+"} {
		l, tokens := lex(input)
		if len(tokens) != 2 || tokens[0].Kind != token.Invalid || tokens[1].Kind != token.EOF {
			t.Fatalf("%q: unexpected tokens: %#v", input, tokens)
		}
		if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1002" {
			t.Fatalf("%q: expected G1002, got %v", input, l.Diagnostics())
		}
	}
}

func TestLexCommentsAreSkippedByDefault(t *testing.T) {
	l, tokens := lex("x // hello\ny /* block */ z")
	if len(tokens) != 4 || tokens[0].Kind != token.Identifier || tokens[1].Kind != token.Identifier || tokens[2].Kind != token.Identifier || tokens[3].Kind != token.EOF {
		t.Fatalf("unexpected comment handling: %#v", tokens)
	}
	if len(l.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", l.Diagnostics())
	}
}

func TestLexCommentsCanBePreserved(t *testing.T) {
	l := New(source.File{ID: 1, Path: "main.gogo", Text: "x // hello\ny"})
	l.IncludeComments(true)
	tokens := l.LexAll()
	if len(tokens) != 4 || tokens[0].Kind != token.Identifier || tokens[1].Kind != token.Comment || tokens[2].Kind != token.Identifier || tokens[3].Kind != token.EOF {
		t.Fatalf("unexpected preserved comments: %#v", tokens)
	}
}

func TestLexUnterminatedString(t *testing.T) {
	l, tokens := lex(`"hello`)
	if len(tokens) != 2 || tokens[0].Kind != token.Invalid || tokens[1].Kind != token.EOF {
		t.Fatalf("unexpected unterminated string tokens: %#v", tokens)
	}
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1003" {
		t.Fatalf("expected G1003, got %v", l.Diagnostics())
	}
}

func TestLexInvalidEscape(t *testing.T) {
	l, tokens := lex(`"hello\q"`)
	if len(tokens) != 2 || tokens[0].Kind != token.Invalid || tokens[1].Kind != token.EOF {
		t.Fatalf("unexpected invalid escape tokens: %#v", tokens)
	}
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1004" {
		t.Fatalf("expected G1004, got %v", l.Diagnostics())
	}
}

func TestLexUnterminatedBlockComment(t *testing.T) {
	l, tokens := lex("/* hello")
	if len(tokens) != 1 || tokens[0].Kind != token.EOF {
		t.Fatalf("unexpected block comment tokens: %#v", tokens)
	}
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1001" {
		t.Fatalf("expected G1001, got %v", l.Diagnostics())
	}
}

func TestLexMalformedUTF8(t *testing.T) {
	l := New(source.File{ID: 1, Path: "main.gogo", Text: string([]byte{'x', 0xff, 'y'})})
	tokens := l.LexAll()
	if len(tokens) != 4 || tokens[0].Kind != token.Identifier || tokens[1].Kind != token.Invalid || tokens[2].Kind != token.Identifier || tokens[3].Kind != token.EOF {
		t.Fatalf("unexpected malformed UTF-8 tokens: %#v", tokens)
	}
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1000" {
		t.Fatalf("expected G1000, got %v", l.Diagnostics())
	}
}

func TestLexSpansUseSourcePositions(t *testing.T) {
	_, tokens := lex("α\nবাংলা")
	if tokens[0].Span.Start.Line != 1 || tokens[0].Span.Start.Column != 1 {
		t.Fatalf("unexpected first span: %+v", tokens[0].Span)
	}
	if tokens[1].Span.Start.Line != 2 || tokens[1].Span.Start.Column != 1 {
		t.Fatalf("unexpected second span: %+v", tokens[1].Span)
	}
}

func FuzzLexNeverPanics(f *testing.F) {
	f.Add("create variable user as \"Alex\"")
	f.Add("বাংলা নাম")
	f.Add("1.25e-4 && value")
	f.Add("/* comment */ x")
	f.Fuzz(func(t *testing.T, text string) {
		l := New(source.File{ID: 1, Path: "fuzz.gogo", Text: text})
		_ = l.LexAll()
		_ = l.Diagnostics()
	})
}
