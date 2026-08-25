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
	Question
	Ellipsis

	Equal
	EqualEqual
	EqualEqualEqual
	Bang
	BangEqual
	BangEqualEqual
	Plus
	PlusEqual
	PlusPlus
	Minus
	MinusEqual
	MinusMinus
	Arrow
	FatArrow
	Star
	StarEqual
	StarStar
	StarStarEqual
	Slash
	SlashEqual
	Percent
	PercentEqual
	Less
	LessEqual
	ShiftLeft
	ShiftLeftEqual
	Greater
	GreaterEqual
	ShiftRight
	ShiftRightEqual
	UnsignedShiftRight
	UnsignedShiftRightEqual
	And
	AndAnd
	AndEqual
	Or
	OrOr
	OrEqual
	Caret
	CaretEqual
	Tilde
	QuestionQuestion
	QuestionDot

	Comment
)

type Token struct {
	Kind Kind
	Text string
	Span source.Span
}

func New(kind Kind, text string, span source.Span) Token {
	return Token{Kind: kind, Text: text, Span: span}
}

func (t Token) IsTrivia() bool { return t.Kind == Comment }
func (t Token) IsEOF() bool    { return t.Kind == EOF }
func (t Token) IsInvalid() bool { return t.Kind == Invalid }

func (k Kind) String() string {
	switch k {
	case EOF: return "end of file"
	case Identifier: return "identifier"
	case Number: return "number"
	case String: return "string"
	case Invalid: return "invalid token"
	case LBrace: return "{"
	case RBrace: return "}"
	case LParen: return "("
	case RParen: return ")"
	case LBracket: return "["
	case RBracket: return "]"
	case Comma: return ","
	case Dot: return "."
	case Colon: return ":"
	case Semicolon: return ";"
	case Question: return "?"
	case Ellipsis: return "..."
	case Equal: return "="
	case EqualEqual: return "=="
	case EqualEqualEqual: return "==="
	case Bang: return "!"
	case BangEqual: return "!="
	case BangEqualEqual: return "!=="
	case Plus: return "+"
	case PlusEqual: return "+="
	case PlusPlus: return "++"
	case Minus: return "-"
	case MinusEqual: return "-="
	case MinusMinus: return "--"
	case Arrow: return "->"
	case FatArrow: return "=>"
	case Star: return "*"
	case StarEqual: return "*="
	case StarStar: return "**"
	case StarStarEqual: return "**="
	case Slash: return "/"
	case SlashEqual: return "/="
	case Percent: return "%"
	case PercentEqual: return "%="
	case Less: return "<"
	case LessEqual: return "<="
	case ShiftLeft: return "<<"
	case ShiftLeftEqual: return "<<="
	case Greater: return ">"
	case GreaterEqual: return ">="
	case ShiftRight: return ">>"
	case ShiftRightEqual: return ">>="
	case UnsignedShiftRight: return ">>>"
	case UnsignedShiftRightEqual: return ">>>="
	case And: return "&"
	case AndAnd: return "&&"
	case AndEqual: return "&="
	case Or: return "|"
	case OrOr: return "||"
	case OrEqual: return "|="
	case Caret: return "^"
	case CaretEqual: return "^="
	case Tilde: return "~"
	case QuestionQuestion: return "??"
	case QuestionDot: return "?."
	case Comment: return "comment"
	default: return "unknown token"
	}
}
