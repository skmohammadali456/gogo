package token

import "github.com/skmohammadali786/gogo/internal/source"

type Kind uint16

const (
	EOF Kind = iota
	Identifier
	Number
	String
	Invalid

	LBrace
	RBrace
	LParen
	RParen
	LBracket
	RBracket
	Comma
	Dot
	Colon
	Semicolon

	Equal
	EqualEqual
	Bang
	BangEqual
	Plus
	PlusEqual
	Minus
	MinusEqual
	Star
	StarEqual
	Slash
	SlashEqual
	Percent
	PercentEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	AndAnd
	OrOr
	Arrow

	Comment
)

type Token struct {
	Kind tokenKindAlias
	Text string
	Span source.Span
}

// tokenKindAlias keeps Token's public representation stable while allowing
// Kind to remain the canonical token enumeration used by the lexer and parser.
type tokenKindAlias = Kind

func New(kind Kind, text string, span source.Span) Token {
	return Token{Kind: kind, Text: text, Span: span}
}

func (t Token) IsTrivia() bool { return t.Kind == Comment }

func (k Kind) String() string {
	switch k {
	case EOF:
		return "end of file"
	case Identifier:
		return "identifier"
	case Number:
		return "number"
	case String:
		return "string"
	case Invalid:
		return "invalid token"
	case LBrace:
		return "{"
	case RBrace:
		return "}"
	case LParen:
		return "("
	case RParen:
		return ")"
	case LBracket:
		return "["
	case RBracket:
		return "]"
	case Comma:
		return ","
	case Dot:
		return "."
	case Colon:
		return ":"
	case Semicolon:
		return ";"
	case Equal:
		return "="
	case EqualEqual:
		return "=="
	case Bang:
		return "!"
	case BangEqual:
		return "!="
	case Plus:
		return "+"
	case PlusEqual:
		return "+="
	case Minus:
		return "-"
	case MinusEqual:
		return "-="
	case Star:
		return "*"
	case StarEqual:
		return "*="
	case Slash:
		return "/"
	case SlashEqual:
		return "/="
	case Percent:
		return "%"
	case PercentEqual:
		return "%="
	case Less:
		return "<"
	case LessEqual:
		return "<="
	case Greater:
		return ">"
	case GreaterEqual:
		return ">="
	case AndAnd:
		return "&&"
	case OrOr:
		return "||"
	case Arrow:
		return "->"
	case Comment:
		return "comment"
	default:
		return "unknown token"
	}
}
