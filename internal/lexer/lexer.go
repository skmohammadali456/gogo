package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

type Lexer struct {
	file        source.File
	cursor      *source.Cursor
	diagnostics diagnostics.Bag
	includeComments bool
}

func New(file source.File) *Lexer {
	return &Lexer{file: file, cursor: source.NewCursor(file.Text)}
}

func (l *Lexer) IncludeComments(enabled bool) { l.includeComments = enabled }

func (l *Lexer) Diagnostics() []diagnostics.Diagnostic { return l.diagnostics.All() }

func (l *Lexer) LexAll() []token.Token {
	var out []token.Token
	for {
		t := l.Next()
		if t.Kind != token.Comment || l.includeComments {
			out = append(out, t)
		}
		if t.Kind == token.EOF {
			return out
		}
	}
}

func (l *Lexer) Next() token.Token {
	for !l.cursor.Done() {
		if t, ok := l.skipSpaceAndComments(); ok {
			if l.includeComments {
				return t
			}
			continue
		}
		break
	}

	start := l.cursor.Offset
	if l.cursor.Done() {
		return l.make(token.EOF, start)
	}

	if !utf8.ValidString(l.file.Text[l.cursor.Offset:]) {
		l.advanceRune()
		return l.invalid(start, "I found invalid UTF-8 in this source file.", "Save the GOGO source as UTF-8 and try again.")
	}

	r, _ := l.cursor.Peek()
	if isIdentifierStart(r) {
		return l.identifier(start)
	}
	if unicode.IsDigit(r) {
		return l.number(start)
	}
	if r == '"' || r == '\'' {
		return l.stringLiteral(start, r)
	}

	pairs := []struct{ text string; kind token.Kind }{
		{"==", token.EqualEqual}, {"!=", token.BangEqual}, {"+=", token.PlusEqual},
		{"-=", token.MinusEqual}, {"*=", token.StarEqual}, {"/=", token.SlashEqual},
		{"%=", token.PercentEqual}, {"<=", token.LessEqual}, {">=", token.GreaterEqual},
		{"&&", token.AndAnd}, {"||", token.OrOr}, {"->", token.Arrow},
	}
	for _, p := range pairs {
		if l.cursor.Match(p.text) {
			return l.make(p.kind, start)
		}
	}

	single := map[rune]token.Kind{
		'{': token.LBrace, '}': token.RBrace, '(': token.LParen, ')': token.RParen,
		'[': token.LBracket, ']': token.RBracket, ',': token.Comma, '.': token.Dot,
		':': token.Colon, ';': token.Semicolon, '=': token.Equal, '!': token.Bang,
		'+': token.Plus, '-': token.Minus, '*': token.Star, '/': token.Slash,
		'%': token.Percent, '<': token.Less, '>': token.Greater,
	}
	if kind, ok := single[r]; ok {
		l.advanceRune()
		return l.make(kind, start)
	}

	l.advanceRune()
	return l.invalid(start, "I don't recognize this character yet.", "Check the spelling or use a supported GOGO operator or punctuation mark.")
}

func (l *Lexer) skipSpaceAndComments() (token.Token, bool) {
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if unicode.IsSpace(r) {
			l.advanceRune()
			continue
		}
		start := l.cursor.Offset
		if l.cursor.Match("//") {
			for !l.cursor.Done() {
				r, _ := l.cursor.Peek()
				if r == '\n' || r == '\r' { break }
				l.advanceRune()
			}
			if l.includeComments { return l.make(token.Comment, start), true }
			continue
		}
		if l.cursor.Match("/*") {
			closed := false
			for !l.cursor.Done() {
				if l.cursor.Match("*/") { closed = true; break }
				l.advanceRune()
			}
			if !closed {
				l.diagnostics.Add(diagnostics.Diagnostic{
					Severity: diagnostics.Error, Code: "G1001",
					Message: "This block comment never closes.",
					Hint: "Add */ to close the comment.", Span: l.span(start),
				})
			}
			if l.includeComments { return l.make(token.Comment, start), true }
			continue
		}
		return token.Token{}, false
	}
	return token.Token{}, false
}

func (l *Lexer) identifier(start int) token.Token {
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if !isIdentifierContinue(r) { break }
		l.advanceRune()
	}
	return l.make(token.Identifier, start)
}

func (l *Lexer) number(start int) token.Token {
	seenDot := false
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if unicode.IsDigit(r) { l.advanceRune(); continue }
		if r == '.' && !seenDot {
			seenDot = true
			l.advanceRune()
			continue
		}
		break
	}
	text := l.file.Text[start:l.cursor.Offset]
	if strings.HasSuffix(text, ".") {
		l.diagnostics.Add(diagnostics.Diagnostic{
			Severity: diagnostics.Error, Code: "G1002",
			Message: "This number ends with a decimal point.",
			Hint: "Write a digit after the decimal point, for example 10.5.", Span: l.span(start),
		})
		return token.New(token.Invalid, text, l.span(start))
	}
	return l.make(token.Number, start)
}

func (l *Lexer) stringLiteral(start int, quote rune) token.Token {
	l.advanceRune()
	closed := false
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if r == '\\' {
			l.advanceRune()
			if !l.cursor.Done() { l.advanceRune() }
			continue
		}
		if r == quote {
			l.advanceRune()
			closed = true
			break
		}
		if r == '\n' || r == '\r' {
			break
		}
		l.advanceRune()
	}
	if !closed {
		l.diagnostics.Add(diagnostics.Diagnostic{
			Severity: diagnostics.Error, Code: "G1003",
			Message: "This string is missing its closing quote.",
			Hint: "Close the string with the same quote that opened it.", Span: l.span(start),
		})
		return token.New(token.Invalid, l.file.Text[start:l.cursor.Offset], l.span(start))
	}
	return l.make(token.String, start)
}

func (l *Lexer) advanceRune() { l.cursor.Advance() }

func (l *Lexer) make(kind token.Kind, start int) token.Token {
	return token.New(kind, l.file.Text[start:l.cursor.Offset], l.span(start))
}

func (l *Lexer) invalid(start int, message, hint string) token.Token {
	l.diagnostics.Add(diagnostics.Diagnostic{
		Severity: diagnostics.Error, Code: "G1000", Message: message, Hint: hint, Span: l.span(start),
	})
	return l.make(token.Invalid, start)
}

func (l *Lexer) span(start int) source.Span {
	return source.Span{Start: source.PositionAt(l.file.Text, start), End: source.PositionAt(l.file.Text, l.cursor.Offset)}
}

func isIdentifierStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }
func isIdentifierContinue(r rune) bool { return isIdentifierStart(r) || unicode.IsDigit(r) || unicode.IsMark(r) }
