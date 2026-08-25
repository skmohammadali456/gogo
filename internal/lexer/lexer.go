package lexer

import (
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

type Lexer struct {
	file            source.File
	cursor          *source.Cursor
	diagnostics     diagnostics.Bag
	includeComments bool
}

func New(file source.File) *Lexer { return &Lexer{file: file, cursor: source.NewCursor(file.Text)} }
func (l *Lexer) IncludeComments(enabled bool) { l.includeComments = enabled }
func (l *Lexer) Diagnostics() []diagnostics.Diagnostic { return l.diagnostics.All() }

func (l *Lexer) LexAll() []token.Token {
	var out []token.Token
	for {
		t := l.Next()
		if t.Kind != token.Comment || l.includeComments { out = append(out, t) }
		if t.Kind == token.EOF { return out }
	}
}

func (l *Lexer) Next() token.Token {
	for !l.cursor.Done() {
		if t, ok := l.skipSpaceAndComments(); ok {
			if l.includeComments { return t }
			continue
		}
		break
	}
	start := l.cursor.Offset
	if l.cursor.Done() { return l.make(token.EOF, start) }

	if !utf8.ValidString(l.file.Text[l.cursor.Offset:]) {
		l.advanceRune()
		return l.invalid(start, "I found invalid UTF-8 in this source file.", "Save the GOGO source as UTF-8 and try again.")
	}

	r, _ := l.cursor.Peek()
	if isIdentifierStart(r) { return l.identifier(start) }
	if unicode.IsDigit(r) { return l.number(start) }
	if r == '"' || r == '\'' { return l.stringLiteral(start, r) }

	pairs := []struct { text string; kind token.Kind }{
		{">>>=", token.UnsignedShiftRightEqual}, {"===", token.EqualEqualEqual}, {"!==", token.BangEqualEqual},
		{"<<=", token.ShiftLeftEqual}, {">>=", token.ShiftRightEqual}, {"**=", token.StarStarEqual},
		{"+=", token.PlusEqual}, {"-=", token.MinusEqual}, {"*=", token.StarEqual}, {"/=", token.SlashEqual},
		{"%=", token.PercentEqual}, {"&=", token.AndEqual}, {"|=", token.OrEqual}, {"^=", token.CaretEqual},
		{"==", token.EqualEqual}, {"!=", token.BangEqual}, {"++", token.PlusPlus}, {"--", token.MinusMinus},
		{"->", token.Arrow}, {"=>", token.FatArrow}, {"**", token.StarStar}, {"<<", token.ShiftLeft},
		{">>>", token.UnsignedShiftRight}, {">>", token.ShiftRight}, {"&&", token.AndAnd}, {"||", token.OrOr},
		{"??", token.QuestionQuestion}, {"?.", token.QuestionDot}, {"...", token.Ellipsis},
		{"<=", token.LessEqual}, {">=", token.GreaterEqual},
	}
	for _, p := range pairs {
		if l.cursor.Match(p.text) { return l.make(p.kind, start) }
	}

	single := map[rune]token.Kind{
		'{': token.LBrace, '}': token.RBrace, '(': token.LParen, ')': token.RParen,
		'[': token.LBracket, ']': token.RBracket, ',': token.Comma, '.': token.Dot,
		':': token.Colon, ';': token.Semicolon, '?': token.Question, '=': token.Equal,
		'!': token.Bang, '+': token.Plus, '-': token.Minus, '*': token.Star, '/': token.Slash,
		'%': token.Percent, '<': token.Less, '>': token.Greater, '&': token.And, '|': token.Or,
		'^': token.Caret, '~': token.Tilde,
	}
	if kind, ok := single[r]; ok { l.advanceRune(); return l.make(kind, start) }
	l.advanceRune()
	return l.invalid(start, "I don't recognize this character yet.", "Check the spelling or use a supported GOGO operator or punctuation mark.")
}

func (l *Lexer) skipSpaceAndComments() (token.Token, bool) {
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if unicode.IsSpace(r) { l.advanceRune(); continue }
		start := l.cursor.Offset
		if l.cursor.Match("//") {
			for !l.cursor.Done() { r, _ = l.cursor.Peek(); if r == '\n' || r == '\r' { break }; l.advanceRune() }
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
				l.diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G1001", Message: "This block comment never closes.", Hint: "Add */ to close the comment.", Span: l.span(start)})
			}
			if l.includeComments { return l.make(token.Comment, start), true }
			continue
		}
		return token.Token{}, false
	}
	return token.Token{}, false
}

func (l *Lexer) identifier(start int) token.Token {
	for !l.cursor.Done() { r, _ := l.cursor.Peek(); if !isIdentifierContinue(r) { break }; l.advanceRune() }
	return l.make(token.Identifier, start)
}

func (l *Lexer) number(start int) token.Token {
	if l.cursor.Match("0x") || l.cursor.Match("0X") { return l.radixNumber(start, 16) }
	if l.cursor.Match("0b") || l.cursor.Match("0B") { return l.radixNumber(start, 2) }
	if l.cursor.Match("0o") || l.cursor.Match("0O") { return l.radixNumber(start, 8) }

	if !l.consumeDigits(10) { return l.invalidNumber(start, "This number is missing digits.", "Write a number such as 42.") }
	if !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if r == '.' {
			l.advanceRune()
			if l.cursor.Done() || !l.hasDigit() { return l.invalidNumber(start, "This decimal point must be followed by a digit.", "Write a decimal such as 10.5, not 10.") }
			l.consumeDigits(10)
		}
	}
	if !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if r == 'e' || r == 'E' {
			l.advanceRune()
			if !l.cursor.Done() { r, _ = l.cursor.Peek(); if r == '+' || r == '-' { l.advanceRune() } }
			if l.cursor.Done() || !l.hasDigit() { return l.invalidNumber(start, "This exponent is missing digits.", "Write an exponent such as 1e6 or 2.5E-3.") }
			l.consumeDigits(10)
		}
	}
	if !l.cursor.Done() { r, _ := l.cursor.Peek(); if r == 'n' { l.advanceRune() } }
	return l.make(token.Number, start)
}

func (l *Lexer) radixNumber(start, base int) token.Token {
	if !l.consumeDigits(base) { return l.invalidNumber(start, "This based number is missing valid digits.", "Use hexadecimal, binary, or octal digits after its prefix.") }
	if !l.cursor.Done() { r, _ := l.cursor.Peek(); if r == 'n' { l.advanceRune() } }
	return l.make(token.Number, start)
}

func (l *Lexer) consumeDigits(base int) bool {
	count := 0
	lastUnderscore := false
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if r == '_' {
			if count == 0 || lastUnderscore { break }
			lastUnderscore = true
			l.advanceRune()
			continue
		}
		if digitValue(r) >= base { break }
		lastUnderscore = false
		count++
		l.advanceRune()
	}
	if lastUnderscore { return false }
	return count > 0
}

func (l *Lexer) hasDigit() bool { r, _ := l.cursor.Peek(); return unicode.IsDigit(r) }
func digitValue(r rune) int {
	switch { case r >= '0' && r <= '9': return int(r-'0'); case r >= 'a' && r <= 'f': return int(r-'a')+10; case r >= 'A' && r <= 'F': return int(r-'A')+10; default: return -1 }
}

func (l *Lexer) stringLiteral(start int, quote rune) token.Token {
	l.advanceRune()
	for !l.cursor.Done() {
		r, _ := l.cursor.Peek()
		if r == '\\' {
			l.advanceRune()
			if l.cursor.Done() { break }
			escaped, _ := l.cursor.Peek()
			if isSimpleEscape(escaped) { l.advanceRune(); continue }
			if escaped == 'x' { if !l.consumeHexEscape(2) { return l.invalidString(start, "This string contains an invalid hexadecimal escape.", "Use an escape such as \\x41.") }; continue }
			if escaped == 'u' { if !l.consumeUnicodeEscape() { return l.invalidString(start, "This string contains an invalid Unicode escape.", "Use \\u0041 or \\u{1F600} with valid hexadecimal digits.") }; continue }
			l.advanceRune()
			return l.invalidString(start, "This string contains an unsupported escape sequence.", "Use a supported escape sequence.")
		}
		if r == quote { l.advanceRune(); return l.make(token.String, start) }
		if r == '\n' || r == '\r' { break }
		l.advanceRune()
	}
	return l.invalidString(start, "This string is missing its closing quote.", "Close the string with the same quote that opened it.")
}

func (l *Lexer) consumeHexEscape(n int) bool {
	l.advanceRune()
	for i := 0; i < n; i++ { if l.cursor.Done() { return false }; r, _ := l.cursor.Peek(); if digitValue(r) < 0 || digitValue(r) >= 16 { return false }; l.advanceRune() }
	return true
}

func (l *Lexer) consumeUnicodeEscape() bool {
	l.advanceRune()
	if l.cursor.Done() { return false }
	r, _ := l.cursor.Peek()
	if r == '{' {
		l.advanceRune(); digits := 0
		for !l.cursor.Done() { r, _ = l.cursor.Peek(); if r == '}' { l.advanceRune(); return digits > 0 && digits <= 6 }; if digitValue(r) < 0 || digitValue(r) >= 16 { return false }; digits++; l.advanceRune(); if digits > 6 { return false } }
		return false
	}
	for i := 0; i < 4; i++ { if l.cursor.Done() { return false }; r, _ = l.cursor.Peek(); if digitValue(r) < 0 || digitValue(r) >= 16 { return false }; l.advanceRune() }
	return true
}

func (l *Lexer) invalidString(start int, message, hint string) token.Token {
	l.diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G1004", Message: message, Hint: hint, Span: l.span(start)})
	return token.New(token.Invalid, l.file.Text[start:l.cursor.Offset], l.span(start))
}
func (l *Lexer) advanceRune() { l.cursor.Advance() }
func (l *Lexer) make(kind token.Kind, start int) token.Token { return token.New(kind, l.file.Text[start:l.cursor.Offset], l.span(start)) }
func (l *Lexer) invalid(start int, message, hint string) token.Token { l.diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G1000", Message: message, Hint: hint, Span: l.span(start)}); return l.make(token.Invalid, start) }
func (l *Lexer) invalidNumber(start int, message, hint string) token.Token { l.diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G1002", Message: message, Hint: hint, Span: l.span(start)}); return l.make(token.Invalid, start) }
func (l *Lexer) span(start int) source.Span { return source.Span{Start: source.PositionAt(l.file.Text, start), End: source.PositionAt(l.file.Text, l.cursor.Offset)} }
func isIdentifierStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }
func isIdentifierContinue(r rune) bool { return isIdentifierStart(r) || unicode.IsDigit(r) || unicode.IsMark(r) }
func isSimpleEscape(r rune) bool { switch r { case 'n','r','t','b','f','v','0','\\','"','\'': return true; default: return false } }

// Keep strconv linked into the lexer package so the lexical grammar can share
// Go's exact rune classification conventions when numeric decoding is added.
var _ = strconv.IntSize
