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

func assertKinds(t *testing.T, tokens []token.Token, want ...token.Kind) {
	t.Helper()
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d: %#v", len(tokens), len(want), tokens)
	}
	for i, kind := range want {
		if tokens[i].Kind != kind {
			t.Fatalf("token %d: got %v, want %v", i, tokens[i].Kind, kind)
		}
	}
}

func TestLexEnglishBengaliHindiIdentifiers(t *testing.T) {
	l, tokens := lex("create user\nবাংলা নাম\nहिन्दी नाम")
	assertKinds(t, tokens, token.Identifier, token.Identifier, token.Identifier, token.Identifier, token.Identifier, token.Identifier, token.EOF)
	for _, tok := range tokens[:6] {
		if tok.Text == "" || tok.Span.Start.Line < 1 {
			t.Fatalf("invalid token: %+v", tok)
		}
	}
	if len(l.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", l.Diagnostics())
	}
}

func TestLexNumbersAndJSStyleNumericForms(t *testing.T) {
	_, tokens := lex("42 10.5 1e6 2.5E-3 1_000 0xff 0b1010 0o755 42n")
	assertKinds(t, tokens, token.Number, token.Number, token.Number, token.Number, token.Number, token.Number, token.Number, token.Number, token.Number, token.EOF)
}

func TestLexInvalidNumberDiagnostics(t *testing.T) {
	for _, input := range []string{"10.", "1e", "1e+", "1_", "1__2", "0x", "0b", "0o", "123abc", "0x1g", "0b1012", "42n4"} {
		l, tokens := lex(input)
		if tokens[0].Kind != token.Invalid || len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1002" {
			t.Fatalf("%q: unexpected result: tokens=%#v diagnostics=%v", input, tokens, l.Diagnostics())
		}
	}
}

func TestLexUnicodeDigitsDoNotBecomeDecimalDigits(t *testing.T) {
	for _, input := range []string{"1.٢", "٢"} {
		l, tokens := lex(input)
		if tokens[0].Kind != token.Invalid || len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1002" {
			t.Fatalf("%q: unexpected Unicode-digit result: tokens=%#v diagnostics=%v", input, tokens, l.Diagnostics())
		}
	}
}

func TestLexCompleteOperatorSet(t *testing.T) {
	input := "= == === ! != !== + += ++ - -= -- -> => * *= ** **= / /= % %= < <= << <<= > >= >> >>= >>> >>>= & && &= | || |= ^ ^= ~ ? ?? ?. ..."
	_, tokens := lex(input)
	assertKinds(t, tokens,
		token.Equal, token.EqualEqual, token.EqualEqualEqual, token.Bang, token.BangEqual, token.BangEqualEqual,
		token.Plus, token.PlusEqual, token.PlusPlus, token.Minus, token.MinusEqual, token.MinusMinus, token.Arrow, token.FatArrow,
		token.Star, token.StarEqual, token.StarStar, token.StarStarEqual, token.Slash, token.SlashEqual, token.Percent, token.PercentEqual,
		token.Less, token.LessEqual, token.ShiftLeft, token.ShiftLeftEqual, token.Greater, token.GreaterEqual, token.ShiftRight,
		token.ShiftRightEqual, token.UnsignedShiftRight, token.UnsignedShiftRightEqual, token.And, token.AndAnd, token.AndEqual,
		token.Or, token.OrOr, token.OrEqual, token.Caret, token.CaretEqual, token.Tilde, token.Question, token.QuestionQuestion,
		token.QuestionDot, token.Ellipsis, token.EOF,
	)
}

func TestLexPunctuation(t *testing.T) {
	_, tokens := lex("{} () [] , . : ;")
	assertKinds(t, tokens, token.LBrace, token.RBrace, token.LParen, token.RParen, token.LBracket, token.RBracket, token.Comma, token.Dot, token.Colon, token.Semicolon, token.EOF)
}

func TestLexStringsAndUnicodeEscapes(t *testing.T) {
	_, tokens := lex(`"hello\n\t\x41\u0042\u{1F600}" 'বাংলা' "हिन्दी"`)
	assertKinds(t, tokens, token.String, token.String, token.String, token.EOF)
}

func TestLexInvalidStringEscapes(t *testing.T) {
	for _, input := range []string{`"hello\q"`, `"\x1"`, `"\u12"`, `"\u{1G}"`, `"\u{}"`, `"\u{110000}"`, `"\u{D800}"`} {
		l, tokens := lex(input)
		if tokens[0].Kind != token.Invalid || len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1004" {
			t.Fatalf("%q: unexpected result: %#v %v", input, tokens, l.Diagnostics())
		}
	}
}

func TestLexCommentsAreSkippedByDefault(t *testing.T) {
	l, tokens := lex("x // hello\ny /* block */ z")
	assertKinds(t, tokens, token.Identifier, token.Identifier, token.Identifier, token.EOF)
	if len(l.Diagnostics()) != 0 {
		t.Fatalf("unexpected diagnostics: %v", l.Diagnostics())
	}
}

func TestLexCommentsCanBePreserved(t *testing.T) {
	l := New(source.File{ID: 1, Path: "main.gogo", Text: "x // hello\ny"})
	l.IncludeComments(true)
	assertKinds(t, l.LexAll(), token.Identifier, token.Comment, token.Identifier, token.EOF)
}

func TestLexUnterminatedString(t *testing.T) {
	l, tokens := lex(`"hello`)
	assertKinds(t, tokens, token.Invalid, token.EOF)
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1004" {
		t.Fatalf("expected G1004, got %v", l.Diagnostics())
	}
}

func TestLexUnterminatedBlockComment(t *testing.T) {
	l, tokens := lex("/* hello")
	assertKinds(t, tokens, token.EOF)
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1001" {
		t.Fatalf("expected G1001, got %v", l.Diagnostics())
	}
}

func TestLexMalformedUTF8(t *testing.T) {
	l := New(source.File{ID: 1, Path: "main.gogo", Text: string([]byte{'x', 0xff, 'y'})})
	assertKinds(t, l.LexAll(), token.Identifier, token.Invalid, token.Identifier, token.EOF)
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1000" {
		t.Fatalf("expected G1000, got %v", l.Diagnostics())
	}
}

func TestLexMalformedUTF8InsideComment(t *testing.T) {
	text := string([]byte{'/', '*', ' ', 0xff, ' ', '*', '/'})
	l := New(source.File{ID: 1, Path: "main.gogo", Text: text})
	assertKinds(t, l.LexAll(), token.EOF)
	if len(l.Diagnostics()) != 1 || l.Diagnostics()[0].Code != "G1000" {
		t.Fatalf("expected G1000, got %v", l.Diagnostics())
	}
}

func TestLexSpansUseSourcePositions(t *testing.T) {
	_, tokens := lex("α\nবাংলা\nहिन्दी")
	if tokens[0].Span.Start.Line != 1 || tokens[0].Span.Start.Column != 1 {
		t.Fatalf("unexpected first span: %+v", tokens[0].Span)
	}
	if tokens[1].Span.Start.Line != 2 || tokens[1].Span.Start.Column != 1 {
		t.Fatalf("unexpected Bengali span: %+v", tokens[1].Span)
	}
	if tokens[2].Span.Start.Line != 3 || tokens[2].Span.Start.Column != 1 {
		t.Fatalf("unexpected Hindi span: %+v", tokens[2].Span)
	}
}

func FuzzLexNeverPanics(f *testing.F) {
	f.Add("create variable user as \"Alex\"")
	f.Add("বাংলা নাম")
	f.Add("हिन्दी नाम")
	f.Add("1.25e-4 && value ?? fallback")
	f.Add("/* comment */ x")
	f.Fuzz(func(t *testing.T, text string) {
		l := New(source.File{ID: 1, Path: "fuzz.gogo", Text: text})
		_ = l.LexAll()
		_ = l.Diagnostics()
	})
}
